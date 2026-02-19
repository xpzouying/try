package main

import (
	"flag"
	"fmt"
	"os"

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
	case "version", "--version", "-v":
		fmt.Printf("try %s\n", version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		// Treat as search query
		return runExec(args[0])
	}
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
	}

	return nil
}

func runClone(url string) error {
	// TODO: Implement git clone
	fmt.Fprintf(os.Stderr, "Clone not implemented yet. URL: %s\n", url)
	return nil
}

func printUsage() {
	fmt.Println(`try - Manage experimental project directories

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
