package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContainsAny(t *testing.T) {
	tests := []struct {
		args     []string
		targets  []string
		expected bool
	}{
		{[]string{"-h"}, []string{"-h", "--help"}, true},
		{[]string{"--help"}, []string{"-h", "--help"}, true},
		{[]string{"foo", "-h"}, []string{"-h", "--help"}, true},
		{[]string{"foo", "bar"}, []string{"-h", "--help"}, false},
		{[]string{}, []string{"-h", "--help"}, false},
	}

	for _, tc := range tests {
		result := containsAny(tc.args, tc.targets...)
		if result != tc.expected {
			t.Errorf("containsAny(%v, %v) = %v, expected %v",
				tc.args, tc.targets, result, tc.expected)
		}
	}
}

func TestParseGitURI_HTTPS(t *testing.T) {
	tests := []struct {
		uri      string
		user     string
		repo     string
		hasError bool
	}{
		{"https://github.com/tobi/try", "tobi", "try", false},
		{"https://github.com/user/repo.git", "user", "repo", false},
		{"https://gitlab.com/org/project", "org", "project", false},
		{"https://example.com/user/repo", "user", "repo", false},
	}

	for _, tc := range tests {
		user, repo, err := parseGitURI(tc.uri)
		if tc.hasError {
			if err == nil {
				t.Errorf("parseGitURI(%s) expected error", tc.uri)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseGitURI(%s) unexpected error: %v", tc.uri, err)
			continue
		}
		if user != tc.user || repo != tc.repo {
			t.Errorf("parseGitURI(%s) = (%s, %s), expected (%s, %s)",
				tc.uri, user, repo, tc.user, tc.repo)
		}
	}
}

func TestParseGitURI_SSH(t *testing.T) {
	tests := []struct {
		uri  string
		user string
		repo string
	}{
		{"git@github.com:tobi/try", "tobi", "try"},
		{"git@github.com:user/repo.git", "user", "repo"},
		{"git@gitlab.com:org/project", "org", "project"},
	}

	for _, tc := range tests {
		user, repo, err := parseGitURI(tc.uri)
		if err != nil {
			t.Errorf("parseGitURI(%s) unexpected error: %v", tc.uri, err)
			continue
		}
		if user != tc.user || repo != tc.repo {
			t.Errorf("parseGitURI(%s) = (%s, %s), expected (%s, %s)",
				tc.uri, user, repo, tc.user, tc.repo)
		}
	}
}

func TestParseGitURI_Invalid(t *testing.T) {
	invalids := []string{
		"not-a-url",
		"redis",
		"http://github.com",          // No user/repo
		"ftp://github.com/user/repo", // Wrong scheme
	}

	for _, uri := range invalids {
		_, _, err := parseGitURI(uri)
		if err == nil {
			t.Errorf("parseGitURI(%s) expected error", uri)
		}
	}
}

func TestIsGitURL(t *testing.T) {
	gitURLs := []string{
		"https://github.com/tobi/try",
		"git@github.com:tobi/try.git",
		"https://gitlab.com/org/project",
	}

	for _, url := range gitURLs {
		if !isGitURL(url) {
			t.Errorf("isGitURL(%s) should be true", url)
		}
	}

	notGitURLs := []string{
		"redis",
		"my-project",
		"./path",
		".",
	}

	for _, s := range notGitURLs {
		if isGitURL(s) {
			t.Errorf("isGitURL(%s) should be false", s)
		}
	}
}

func TestResolveUniqueName(t *testing.T) {
	tmpDir := t.TempDir()
	datePrefix := "2024-01-15"

	// First name should be as-is
	name1 := resolveUniqueName(tmpDir, datePrefix, "redis")
	if name1 != "redis" {
		t.Errorf("expected 'redis', got %s", name1)
	}

	// Create the directory
	if err := os.Mkdir(filepath.Join(tmpDir, datePrefix+"-redis"), 0755); err != nil {
		t.Fatal(err)
	}

	// Second name should be versioned
	name2 := resolveUniqueName(tmpDir, datePrefix, "redis")
	if name2 != "redis-2" {
		t.Errorf("expected 'redis-2', got %s", name2)
	}

	// Create that too
	if err := os.Mkdir(filepath.Join(tmpDir, datePrefix+"-redis-2"), 0755); err != nil {
		t.Fatal(err)
	}

	// Third should be -3
	name3 := resolveUniqueName(tmpDir, datePrefix, "redis")
	if name3 != "redis-3" {
		t.Errorf("expected 'redis-3', got %s", name3)
	}
}

func TestResolveUniqueName_NumericSuffix(t *testing.T) {
	tmpDir := t.TempDir()
	datePrefix := "2024-01-15"

	// Create project2
	if err := os.Mkdir(filepath.Join(tmpDir, datePrefix+"-project2"), 0755); err != nil {
		t.Fatal(err)
	}

	// Resolving "project2" should give "project3"
	name := resolveUniqueName(tmpDir, datePrefix, "project2")
	if name != "project3" {
		t.Errorf("expected 'project3', got %s", name)
	}
}

// Integration test: verify help output
func TestRun_Help(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := run([]string{"-h"})

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Errorf("run(-h) returned error: %v", err)
	}

	// Read captured output
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "try - Manage experimental project directories") {
		t.Error("help output should contain description")
	}
	if !strings.Contains(output, "try <git-url>") {
		t.Error("help output should mention git URL auto-detection")
	}
}

// Integration test: verify version output
func TestRun_Version(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	err := run([]string{"--version"})

	w.Close()
	os.Stderr = oldStderr

	if err != nil {
		t.Errorf("run(--version) returned error: %v", err)
	}

	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "try") {
		t.Error("version output should contain 'try'")
	}
}

// Integration test: verify clone command
func TestRun_Clone(t *testing.T) {
	// Set temp tries path
	tmpDir := t.TempDir()
	os.Setenv("TRY_PATH", tmpDir)
	defer os.Unsetenv("TRY_PATH")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := run([]string{"clone", "https://github.com/tobi/try"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("run(clone) returned error: %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "git clone") {
		t.Error("clone output should contain 'git clone'")
	}
	if !strings.Contains(output, "tobi-try") {
		t.Error("clone output should contain repo name")
	}
}

// Integration test: git URL auto-detection
func TestRun_GitURLAutoDetect(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TRY_PATH", tmpDir)
	defer os.Unsetenv("TRY_PATH")

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Pass git URL directly (without 'clone' subcommand)
	err := run([]string{"https://github.com/tobi/try"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("run(git-url) returned error: %v", err)
	}

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "git clone") {
		t.Error("auto-detect should trigger git clone")
	}
}

// Integration test: clone requires URL
func TestRun_CloneNoURL(t *testing.T) {
	err := run([]string{"clone"})
	if err == nil {
		t.Error("clone without URL should return error")
	}
	if !strings.Contains(err.Error(), "URL") {
		t.Errorf("error should mention URL: %v", err)
	}
}

// Integration test: init command
func TestRun_Init(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := run([]string{"init", "bash"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("run(init bash) returned error: %v", err)
	}

	buf := make([]byte, 2048)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "try()") {
		t.Error("init bash should output try function")
	}
}

// Integration test: worktree on non-git directory
func TestRun_WorktreeNonGit(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("TRY_PATH", tmpDir)
	defer os.Unsetenv("TRY_PATH")

	// Create a non-git directory
	nonGitDir := filepath.Join(tmpDir, "non-git")
	if err := os.Mkdir(nonGitDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Capture stderr for the note
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	// Capture stdout for the commands
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Change to non-git dir and run try .
	oldWd, _ := os.Getwd()
	os.Chdir(nonGitDir)
	defer os.Chdir(oldWd)

	err := run([]string{"."})

	w.Close()
	wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	if err != nil {
		t.Errorf("run(.) returned error: %v", err)
	}

	// Check stderr for note
	bufErr := make([]byte, 512)
	nErr, _ := rErr.Read(bufErr)
	stderrOutput := string(bufErr[:nErr])

	if !strings.Contains(stderrOutput, "not a git repository") {
		t.Error("should note that directory is not a git repository")
	}

	// Check stdout for mkdir command
	buf := make([]byte, 512)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "mkdir") {
		t.Error("non-git worktree should create directory")
	}
}
