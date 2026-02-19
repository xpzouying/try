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
  try version          Show version

Examples:
  eval "$(try init bash)"   # Add to ~/.bashrc
  try redis                 # Create or jump to redis experiment
  try clone https://github.com/user/repo

Environment:
  TRY_PATH      Root directory (default: ~/tries)
  TRY_PROJECTS  Graduate destination (default: parent of TRY_PATH)`)
}
