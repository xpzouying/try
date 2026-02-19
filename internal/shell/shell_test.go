package shell

import (
	"os"
	"strings"
	"testing"
)

func TestDetect_Bash(t *testing.T) {
	t.Setenv("SHELL", "/bin/bash")

	result := Detect()
	if result != "bash" {
		t.Errorf("expected 'bash', got %s", result)
	}
}

func TestDetect_Zsh(t *testing.T) {
	t.Setenv("SHELL", "/usr/local/bin/zsh")

	result := Detect()
	if result != "zsh" {
		t.Errorf("expected 'zsh', got %s", result)
	}
}

func TestDetect_Fish(t *testing.T) {
	t.Setenv("SHELL", "/opt/homebrew/bin/fish")

	result := Detect()
	if result != "fish" {
		t.Errorf("expected 'fish', got %s", result)
	}
}

func TestDetect_Empty(t *testing.T) {
	// Unset by setting to empty and checking Detect handles it
	// Note: t.Setenv doesn't support unset, so we test the default case differently
	originalShell := os.Getenv("SHELL")
	if err := os.Unsetenv("SHELL"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if originalShell != "" {
			_ = os.Setenv("SHELL", originalShell)
		}
	}()

	result := Detect()
	if result != "bash" {
		t.Errorf("expected 'bash' as default, got %s", result)
	}
}

func TestWrapper_Bash(t *testing.T) {
	wrapper, err := Wrapper("bash")
	if err != nil {
		t.Fatal(err)
	}

	// Check key elements of bash wrapper
	if !strings.Contains(wrapper, "try()") {
		t.Error("bash wrapper should define try function")
	}
	if !strings.Contains(wrapper, "eval \"$output\"") {
		t.Error("bash wrapper should eval output")
	}
	if !strings.Contains(wrapper, "exec") {
		t.Error("bash wrapper should call exec subcommand")
	}
	if !strings.Contains(wrapper, "~/.bashrc") {
		t.Error("bash wrapper should mention .bashrc")
	}
}

func TestWrapper_Zsh(t *testing.T) {
	wrapper, err := Wrapper("zsh")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(wrapper, "try()") {
		t.Error("zsh wrapper should define try function")
	}
	if !strings.Contains(wrapper, "~/.zshrc") {
		t.Error("zsh wrapper should mention .zshrc")
	}
}

func TestWrapper_Fish(t *testing.T) {
	wrapper, err := Wrapper("fish")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(wrapper, "function try") {
		t.Error("fish wrapper should define try function")
	}
	if !strings.Contains(wrapper, "eval $output") {
		t.Error("fish wrapper should eval output")
	}
	if !strings.Contains(wrapper, "config.fish") {
		t.Error("fish wrapper should mention config.fish")
	}
}

func TestWrapper_Sh(t *testing.T) {
	// "sh" should use bash wrapper
	wrapper, err := Wrapper("sh")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(wrapper, "try()") {
		t.Error("sh wrapper should define try function")
	}
}

func TestWrapper_CaseInsensitive(t *testing.T) {
	tests := []string{"BASH", "Bash", "ZSH", "Zsh", "FISH", "Fish"}
	for _, shell := range tests {
		_, err := Wrapper(shell)
		if err != nil {
			t.Errorf("Wrapper(%s) should not error: %v", shell, err)
		}
	}
}

func TestWrapper_Unsupported(t *testing.T) {
	_, err := Wrapper("powershell")
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("error should mention 'unsupported shell', got: %v", err)
	}
}

func TestWrapper_InitBypass(t *testing.T) {
	// All wrappers should have special handling for 'init'
	shells := []string{"bash", "zsh", "fish"}
	for _, shell := range shells {
		wrapper, err := Wrapper(shell)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(wrapper, "init") {
			t.Errorf("%s wrapper should handle 'init' specially", shell)
		}
	}
}

func TestBashWrapper_ExitCode(t *testing.T) {
	wrapper, _ := Wrapper("bash")

	// Should capture and return exit code
	if !strings.Contains(wrapper, "exit_code") {
		t.Error("bash wrapper should handle exit_code")
	}
	if !strings.Contains(wrapper, "return $exit_code") {
		t.Error("bash wrapper should return exit_code")
	}
}

func TestFishWrapper_Syntax(t *testing.T) {
	wrapper, _ := Wrapper("fish")

	// Fish uses 'end' instead of braces
	if !strings.Contains(wrapper, "end") {
		t.Error("fish wrapper should use 'end' keyword")
	}
	// Fish uses 'set -l' for local variables
	if !strings.Contains(wrapper, "set -l") {
		t.Error("fish wrapper should use 'set -l' for local vars")
	}
	// Fish uses $status instead of $?
	if !strings.Contains(wrapper, "$status") {
		t.Error("fish wrapper should use $status")
	}
}
