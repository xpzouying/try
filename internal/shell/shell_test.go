package shell

import (
	"os"
	"strings"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		shell    string
		expected string
	}{
		{"/bin/bash", "bash"},
		{"/usr/local/bin/zsh", "zsh"},
		{"/opt/homebrew/bin/fish", "fish"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Setenv("SHELL", tc.shell)
			if result := Detect(); result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
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

func TestWrapper(t *testing.T) {
	tests := []struct {
		shell    string
		contains []string
	}{
		{"bash", []string{"try()", "eval \"$output\"", "exec", "~/.bashrc"}},
		{"zsh", []string{"try()", "~/.zshrc"}},
		{"fish", []string{"function try", "eval $output", "config.fish"}},
		{"sh", []string{"try()"}},
	}

	for _, tc := range tests {
		t.Run(tc.shell, func(t *testing.T) {
			wrapper, err := Wrapper(tc.shell)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range tc.contains {
				if !strings.Contains(wrapper, s) {
					t.Errorf("wrapper should contain %q", s)
				}
			}
		})
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
