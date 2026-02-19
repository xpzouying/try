package entry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewEntry_BasicDirectory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test-project")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	entry, err := NewEntry(testDir)
	if err != nil {
		t.Fatal(err)
	}
	if entry == nil {
		t.Fatal("expected non-nil entry")
	}

	if entry.Name != "test-project" {
		t.Errorf("expected Name 'test-project', got %s", entry.Name)
	}
	if entry.Path != testDir {
		t.Errorf("expected Path %s, got %s", testDir, entry.Path)
	}
	if entry.HasDate {
		t.Error("expected HasDate false for non-dated directory")
	}
	if entry.BaseName != "test-project" {
		t.Errorf("expected BaseName 'test-project', got %s", entry.BaseName)
	}
	if entry.IsWorktree {
		t.Error("expected IsWorktree false")
	}
}

func TestNewEntry_DatePrefixed(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "2024-01-15-redis")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	entry, err := NewEntry(testDir)
	if err != nil {
		t.Fatal(err)
	}

	if !entry.HasDate {
		t.Error("expected HasDate true for dated directory")
	}
	if entry.BaseName != "redis" {
		t.Errorf("expected BaseName 'redis', got %s", entry.BaseName)
	}
	if entry.Name != "2024-01-15-redis" {
		t.Errorf("expected Name '2024-01-15-redis', got %s", entry.Name)
	}
}

func TestNewEntry_NonExistent(t *testing.T) {
	_, err := NewEntry("/nonexistent/path")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestNewEntry_File(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	entry, err := NewEntry(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if entry != nil {
		t.Error("expected nil entry for file")
	}
}

func TestLoadEntries_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test directories
	dirs := []string{"2024-01-15-redis", "2024-01-14-postgres", "plain-project"}
	for _, d := range dirs {
		if err := os.Mkdir(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create a hidden directory (should be ignored)
	if err := os.Mkdir(filepath.Join(tmpDir, ".hidden"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := LoadEntries(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestLoadEntries_NonExistent(t *testing.T) {
	entries, err := LoadEntries("/nonexistent/path")
	if err != nil {
		t.Fatal(err)
	}
	if entries != nil {
		t.Error("expected nil entries for non-existent path")
	}
}

func TestLoadEntries_SortedByModTime(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directories with different mod times
	dir1 := filepath.Join(tmpDir, "older")
	dir2 := filepath.Join(tmpDir, "newer")

	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	// Set older time
	oldTime := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(dir1, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	entries, err := LoadEntries(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Newer should be first
	if entries[0].Name != "newer" {
		t.Errorf("expected 'newer' first, got %s", entries[0].Name)
	}
}

func TestScore_Today(t *testing.T) {
	entry := &Entry{
		ModTime: time.Now().Add(-1 * time.Hour),
		HasDate: true,
	}

	score := entry.Score(time.Now())
	// Should get 100 (today) + 10 (date prefix) = 110
	if score != 110 {
		t.Errorf("expected score 110 for today with date prefix, got %f", score)
	}
}

func TestScore_ThisWeek(t *testing.T) {
	entry := &Entry{
		ModTime: time.Now().Add(-3 * 24 * time.Hour), // 3 days ago
		HasDate: false,
	}

	score := entry.Score(time.Now())
	// Should get 50 (this week) + 0 (no date prefix) = 50
	if score != 50 {
		t.Errorf("expected score 50 for this week, got %f", score)
	}
}

func TestScore_ThisMonth(t *testing.T) {
	entry := &Entry{
		ModTime: time.Now().Add(-10 * 24 * time.Hour), // 10 days ago
		HasDate: true,
	}

	score := entry.Score(time.Now())
	// Should get 20 (this month) + 10 (date prefix) = 30
	if score != 30 {
		t.Errorf("expected score 30 for this month with date prefix, got %f", score)
	}
}

func TestScore_Old(t *testing.T) {
	entry := &Entry{
		ModTime: time.Now().Add(-60 * 24 * time.Hour), // 60 days ago
		HasDate: false,
	}

	score := entry.Score(time.Now())
	// Should get 0 (old) + 0 (no date prefix) = 0
	if score != 0 {
		t.Errorf("expected score 0 for old entry, got %f", score)
	}
}

func TestTriesPath(t *testing.T) {
	// Test custom env
	t.Setenv("TRY_PATH", "/custom/tries")
	if path := TriesPath(); path != "/custom/tries" {
		t.Errorf("expected /custom/tries, got %s", path)
	}

	// Test tilde expansion
	t.Setenv("TRY_PATH", "~/my-tries")
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "my-tries")
	if path := TriesPath(); path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestProjectsPath(t *testing.T) {
	t.Setenv("TRY_PROJECTS", "/custom/projects")
	if path := ProjectsPath(); path != "/custom/projects" {
		t.Errorf("expected /custom/projects, got %s", path)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~", home},
	}

	for _, tc := range tests {
		result := expandHome(tc.input)
		if result != tc.expected {
			t.Errorf("expandHome(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestDatePrefixRegex(t *testing.T) {
	tests := []struct {
		input    string
		hasDate  bool
	}{
		{"2024-01-15-redis", true},
		{"2024-12-31-project", true},
		{"2024-1-15-redis", false},      // Single digit month
		{"24-01-15-redis", false},       // Two digit year
		{"2024-01-15redis", false},      // Missing dash after date
		{"redis", false},
		{"2024-01-15", false},           // Just date, no suffix
		{"", false},
	}

	for _, tc := range tests {
		result := datePrefix.MatchString(tc.input)
		if result != tc.hasDate {
			t.Errorf("datePrefix.MatchString(%s) = %v, expected %v", tc.input, result, tc.hasDate)
		}
	}
}

func TestNewEntry_Worktree(t *testing.T) {
	tmpDir := t.TempDir()
	worktreeDir := filepath.Join(tmpDir, "2024-01-15-feature")
	if err := os.Mkdir(worktreeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a .git file (not directory) to simulate worktree
	gitFile := filepath.Join(worktreeDir, ".git")
	gitContent := "gitdir: /path/to/repo/.git/worktrees/feature"
	if err := os.WriteFile(gitFile, []byte(gitContent), 0644); err != nil {
		t.Fatal(err)
	}

	entry, err := NewEntry(worktreeDir)
	if err != nil {
		t.Fatal(err)
	}

	if !entry.IsWorktree {
		t.Error("expected IsWorktree true")
	}
	if entry.SourceRepo != "repo" {
		t.Errorf("expected SourceRepo 'repo', got %s", entry.SourceRepo)
	}
}

func TestParseWorktreeSource(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		content  string
		expected string
	}{
		{"gitdir: /home/user/projects/myrepo/.git/worktrees/feature", "myrepo"},
		{"gitdir: /path/to/repo/.git/worktrees/branch", "repo"},
		{"gitdir: /single/.git/worktrees/wt", "single"},
		{"invalid content", ""},
		{"gitdir: /no/worktrees/path", ""},
	}

	for i, tc := range tests {
		gitFile := filepath.Join(tmpDir, "git"+string(rune('0'+i)))
		if err := os.WriteFile(gitFile, []byte(tc.content), 0644); err != nil {
			t.Fatal(err)
		}

		result := parseWorktreeSource(gitFile)
		if result != tc.expected {
			t.Errorf("parseWorktreeSource(%q) = %s, expected %s", tc.content, result, tc.expected)
		}
	}
}
