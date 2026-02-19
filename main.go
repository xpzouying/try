package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/xpzouying/try/internal/entry"
	"github.com/xpzouying/try/internal/selector"
	"github.com/xpzouying/try/internal/shell"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Check global flags first (like Ruby version)
	// This handles: try exec -h, try -h, try foo -h, etc.
	if containsAny(args, "-h", "--help", "help") {
		printUsage()
		return nil
	}
	if containsAny(args, "-v", "--version", "version") {
		fmt.Fprintln(os.Stderr, "try", version)
		return nil
	}

	if len(args) == 0 {
		return runExec("")
	}

	// Handle . and ./path for worktree creation
	if strings.HasPrefix(args[0], ".") {
		return runWorktree(args)
	}

	switch args[0] {
	case "init":
		return runInit(args[1:])
	case "exec":
		query := ""
		if len(args) > 1 {
			query = args[1]
		}
		return runExec(query)
	case "clone":
		if len(args) < 2 {
			return fmt.Errorf("clone requires a URL argument")
		}
		return runClone(args[1])
	default:
		// Treat as search query
		return runExec(args[0])
	}
}

func containsAny(args []string, targets ...string) bool {
	for _, arg := range args {
		for _, target := range targets {
			if arg == target {
				return true
			}
		}
	}
	return false
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	fs.Parse(args)

	shellName := fs.Arg(0)
	if shellName == "" {
		// Auto-detect from SHELL env
		shellName = shell.Detect()
	}

	wrapper, err := shell.Wrapper(shellName)
	if err != nil {
		return err
	}

	fmt.Print(wrapper)
	return nil
}

func runExec(query string) error {
	// Ensure tries directory exists
	if err := selector.EnsureTriesDir(); err != nil {
		return fmt.Errorf("create tries directory: %w", err)
	}

	result, err := selector.Run(query)
	if err != nil {
		return err
	}

	if result == nil || result.Action == "cancel" {
		return nil
	}

	switch result.Action {
	case "cd":
		// Output cd command for shell to eval
		fmt.Printf("cd %q\n", result.Path)
	case "mkdir":
		// Create directory and cd into it
		fmt.Printf("mkdir -p %q && cd %q\n", result.Path, result.Path)
	case "graduate":
		// Move directory to projects and create symlink
		symlinkPath := filepath.Join(entry.TriesPath(), result.BaseName)
		// Check if source is a git worktree (has .git file, not directory)
		gitFile := filepath.Join(result.Path, ".git")
		info, err := os.Stat(gitFile)
		isWorktree := err == nil && !info.IsDir()

		if isWorktree {
			// Use git worktree move for proper bookkeeping
			fmt.Printf("git worktree move %q %q && ", result.Path, result.DestPath)
		} else {
			fmt.Printf("mv %q %q && ", result.Path, result.DestPath)
		}
		fmt.Printf("ln -s %q %q && ", result.DestPath, symlinkPath)
		fmt.Printf("echo %q && ", fmt.Sprintf("Graduated: %s → %s", result.BaseName, result.DestPath))
		fmt.Printf("cd %q\n", result.DestPath)
	case "delete":
		// Delete directory (stay in current directory or go to tries root)
		triesPath := entry.TriesPath()
		fmt.Printf("rm -rf %q && ", result.Path)
		fmt.Printf("echo %q && ", fmt.Sprintf("Deleted: %s", result.BaseName))
		// If we're in the deleted directory, go to tries root
		fmt.Printf("( cd %q 2>/dev/null || cd %q )\n", os.Getenv("PWD"), triesPath)
	case "rename":
		// Rename directory and cd into it
		triesPath := entry.TriesPath()
		fmt.Printf("cd %q && ", triesPath)
		fmt.Printf("mv %q %q && ", result.BaseName, result.NewName)
		fmt.Printf("echo %q && ", fmt.Sprintf("Renamed: %s → %s", result.BaseName, result.NewName))
		fmt.Printf("cd %q\n", result.DestPath)
	}

	return nil
}

func runClone(gitURL string) error {
	// Ensure tries directory exists
	if err := selector.EnsureTriesDir(); err != nil {
		return fmt.Errorf("create tries directory: %w", err)
	}

	// Parse git URL to get user and repo
	user, repo, err := parseGitURI(gitURL)
	if err != nil {
		return fmt.Errorf("invalid git URL: %w", err)
	}

	// Generate directory name: {date}-{user}-{repo}
	datePrefix := time.Now().Format("2006-01-02")
	dirName := fmt.Sprintf("%s-%s-%s", datePrefix, user, repo)
	fullPath := filepath.Join(entry.TriesPath(), dirName)

	// Output shell commands for clone
	fmt.Printf("mkdir -p %q && ", fullPath)
	fmt.Printf("echo %q && ", fmt.Sprintf("Using git clone to create this trial from %s.", gitURL))
	fmt.Printf("git clone %q %q && ", gitURL, fullPath)
	fmt.Printf("cd %q\n", fullPath)

	return nil
}

// runWorktree handles "try ." and "try ./path" commands
func runWorktree(args []string) error {
	// Ensure tries directory exists
	if err := selector.EnsureTriesDir(); err != nil {
		return fmt.Errorf("create tries directory: %w", err)
	}

	pathArg := args[0]
	customName := ""
	if len(args) > 1 {
		customName = strings.Join(args[1:], "-")
	}

	// Expand the path
	repoDir, err := filepath.Abs(pathArg)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(repoDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("not a directory: %s", pathArg)
	}

	// Check if it's a git repository
	gitPath := filepath.Join(repoDir, ".git")
	isGitRepo := false
	if _, err := os.Stat(gitPath); err == nil {
		isGitRepo = true
	}

	// If git repo and no custom name → show TUI with existing worktrees
	if isGitRepo && customName == "" {
		return runWorktreeTUI(repoDir)
	}

	// Direct creation mode (with custom name or non-git directory)
	return runWorktreeDirect(repoDir, pathArg, customName, isGitRepo)
}

// runWorktreeTUI shows TUI with existing worktrees from the repo
func runWorktreeTUI(repoDir string) error {
	result, err := selector.RunWorktree(repoDir)
	if err != nil {
		return err
	}

	if result == nil || result.Action == "cancel" {
		return nil
	}

	switch result.Action {
	case "cd":
		// Jump to existing worktree
		fmt.Printf("cd %q\n", result.Path)
	case "worktree":
		// Create new worktree
		triesPath := entry.TriesPath()
		finalName := resolveUniqueName(triesPath, time.Now().Format("2006-01-02"), result.NewName)
		fullPath := filepath.Join(triesPath, fmt.Sprintf("%s-%s", time.Now().Format("2006-01-02"), finalName))

		fmt.Printf("mkdir -p %q && ", fullPath)
		fmt.Printf("echo %q && ", fmt.Sprintf("Using git worktree to create this trial from %s.", result.RepoPath))
		fmt.Printf("(cd %q && git worktree add --detach %q 2>/dev/null || true) && ", result.RepoPath, fullPath)
		fmt.Printf("cd %q\n", fullPath)
	}

	return nil
}

// runWorktreeDirect creates a worktree directly without TUI
func runWorktreeDirect(repoDir, pathArg, customName string, isGitRepo bool) error {
	// Determine the base name
	baseName := customName
	if baseName == "" {
		// Use the repository name
		baseName = filepath.Base(repoDir)
		// Try to resolve symlinks for better name
		if realPath, err := filepath.EvalSymlinks(repoDir); err == nil {
			baseName = filepath.Base(realPath)
		}
	}

	// Replace spaces with dashes
	baseName = strings.ReplaceAll(baseName, " ", "-")

	// Generate unique directory name
	datePrefix := time.Now().Format("2006-01-02")
	triesPath := entry.TriesPath()
	finalName := resolveUniqueName(triesPath, datePrefix, baseName)
	fullPath := filepath.Join(triesPath, fmt.Sprintf("%s-%s", datePrefix, finalName))

	// Output shell commands
	if isGitRepo {
		// Create worktree
		fmt.Printf("mkdir -p %q && ", fullPath)
		fmt.Printf("echo %q && ", fmt.Sprintf("Using git worktree to create this trial from %s.", repoDir))
		// Use a subshell to handle worktree creation gracefully
		fmt.Printf("(cd %q && git worktree add --detach %q 2>/dev/null || true) && ", repoDir, fullPath)
		fmt.Printf("cd %q\n", fullPath)
	} else {
		// Not a git repo, just create directory
		fmt.Fprintf(os.Stderr, "Note: %s is not a git repository, creating plain directory.\n", pathArg)
		fmt.Printf("mkdir -p %q && cd %q\n", fullPath, fullPath)
	}

	return nil
}

// resolveUniqueName generates a unique directory name with versioning
func resolveUniqueName(triesPath, datePrefix, baseName string) string {
	initial := fmt.Sprintf("%s-%s", datePrefix, baseName)
	initialPath := filepath.Join(triesPath, initial)

	// If doesn't exist, use as-is
	if _, err := os.Stat(initialPath); os.IsNotExist(err) {
		return baseName
	}

	// Check if baseName ends with a number
	re := regexp.MustCompile(`^(.*?)(\d+)$`)
	if matches := re.FindStringSubmatch(baseName); matches != nil {
		stem := matches[1]
		n, _ := fmt.Sscanf(matches[2], "%d", new(int))
		candidateNum := n + 1
		for {
			candidateName := fmt.Sprintf("%s%d", stem, candidateNum)
			candidatePath := filepath.Join(triesPath, fmt.Sprintf("%s-%s", datePrefix, candidateName))
			if _, err := os.Stat(candidatePath); os.IsNotExist(err) {
				return candidateName
			}
			candidateNum++
		}
	}

	// No numeric suffix, add -2, -3, etc.
	for i := 2; ; i++ {
		candidateName := fmt.Sprintf("%s-%d", baseName, i)
		candidatePath := filepath.Join(triesPath, fmt.Sprintf("%s-%s", datePrefix, candidateName))
		if _, err := os.Stat(candidatePath); os.IsNotExist(err) {
			return candidateName
		}
	}
}

// parseGitURI extracts user and repo from various git URL formats
func parseGitURI(uri string) (user, repo string, err error) {
	// Remove .git suffix if present
	uri = strings.TrimSuffix(uri, ".git")

	// https://github.com/user/repo
	if re := regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)`); re.MatchString(uri) {
		matches := re.FindStringSubmatch(uri)
		return matches[1], matches[2], nil
	}

	// git@github.com:user/repo
	if re := regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+)`); re.MatchString(uri) {
		matches := re.FindStringSubmatch(uri)
		return matches[1], matches[2], nil
	}

	// https://host/user/repo (gitlab, etc.)
	if re := regexp.MustCompile(`^https?://[^/]+/([^/]+)/([^/]+)`); re.MatchString(uri) {
		matches := re.FindStringSubmatch(uri)
		return matches[1], matches[2], nil
	}

	// git@host:user/repo
	if re := regexp.MustCompile(`^git@[^:]+:([^/]+)/([^/]+)`); re.MatchString(uri) {
		matches := re.FindStringSubmatch(uri)
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("could not parse git URL: %s", uri)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `try - Manage experimental project directories

Usage:
  try                  Interactive selector
  try <name>           Jump to or create experiment
  try init [shell]     Output shell wrapper function
  try clone <url>      Clone repository into tries directory
  try .                Show worktrees from current repo (TUI)
  try . <name>         Create worktree with custom name (direct)
  try ./path           Create worktree from specified path
  try version          Show version

Examples:
  eval "$(try init bash)"   # Add to ~/.bashrc
  try redis                 # Create or jump to redis experiment
  try clone https://github.com/user/repo
  try .                     # TUI: show existing worktrees + create
  try . experiment          # Direct: create 2024-01-15-experiment

Environment:
  TRY_PATH      Root directory (default: ~/tries)
  TRY_PROJECTS  Graduate destination (default: parent of TRY_PATH)`)
}
