package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Detect returns the current shell name from SHELL environment variable.
func Detect() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "bash"
	}
	return filepath.Base(shell)
}

// Wrapper returns the shell wrapper function for the given shell.
func Wrapper(shellName string) (string, error) {
	// Get the path to the try binary
	executable, err := os.Executable()
	if err != nil {
		executable = "try"
	}

	switch strings.ToLower(shellName) {
	case "bash", "sh":
		return bashWrapper(executable), nil
	case "zsh":
		return zshWrapper(executable), nil
	case "fish":
		return fishWrapper(executable), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shellName)
	}
}

func bashWrapper(tryPath string) string {
	return fmt.Sprintf(`# try - experimental project directory manager
# Add this to your ~/.bashrc

try() {
  case "$1" in
    init|help|version|-h|--help|-v|--version)
      %[1]q "$@"
      return $?
      ;;
  esac
  local output
  output=$(%[1]q exec "$@")
  local exit_code=$?
  if [[ $exit_code -eq 0 && -n "$output" ]]; then
    eval "$output"
  fi
  return $exit_code
}
`, tryPath)
}

func zshWrapper(tryPath string) string {
	return fmt.Sprintf(`# try - experimental project directory manager
# Add this to your ~/.zshrc

try() {
  case "$1" in
    init|help|version|-h|--help|-v|--version)
      %[1]q "$@"
      return $?
      ;;
  esac
  local output
  output=$(%[1]q exec "$@")
  local exit_code=$?
  if [[ $exit_code -eq 0 && -n "$output" ]]; then
    eval "$output"
  fi
  return $exit_code
}
`, tryPath)
}

func fishWrapper(tryPath string) string {
	return fmt.Sprintf(`# try - experimental project directory manager
# Add this to your ~/.config/fish/config.fish

function try
  switch "$argv[1]"
    case init help version -h --help -v --version
      %[1]s $argv
      return $status
  end
  set -l output (%[1]s exec $argv)
  set -l exit_code $status
  if test $exit_code -eq 0 -a -n "$output"
    eval $output
  end
  return $exit_code
end
`, tryPath)
}
