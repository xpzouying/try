package entry

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

// Entry represents a directory in the tries folder.
type Entry struct {
	Name     string    // Directory name (e.g., "2024-01-15-redis")
	Path     string    // Full path
	ModTime  time.Time // Last modification time
	HasDate  bool      // Whether name starts with date prefix
	BaseName string    // Name without date prefix (e.g., "redis")
}

var datePrefix = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

// NewEntry creates an Entry from a directory path.
func NewEntry(path string) (*Entry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	name := filepath.Base(path)
	hasDate := datePrefix.MatchString(name)
	baseName := name
	if hasDate {
		baseName = name[11:] // Remove "YYYY-MM-DD-" prefix
	}

	return &Entry{
		Name:     name,
		Path:     path,
		ModTime:  info.ModTime(),
		HasDate:  hasDate,
		BaseName: baseName,
	}, nil
}

// LoadEntries loads all directories from the tries path.
func LoadEntries(triesPath string) ([]*Entry, error) {
	entries, err := os.ReadDir(triesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var result []*Entry
	for _, e := range entries {
		if !e.IsDir() || e.Name()[0] == '.' {
			continue
		}
		entry, err := NewEntry(filepath.Join(triesPath, e.Name()))
		if err != nil {
			continue
		}
		if entry != nil {
			result = append(result, entry)
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ModTime.After(result[j].ModTime)
	})

	return result, nil
}

// Score calculates relevance score for an entry.
// Higher score = more relevant.
func (e *Entry) Score(now time.Time) float64 {
	score := 0.0

	// Recency bonus: exponential decay
	age := now.Sub(e.ModTime).Hours()
	if age < 24 {
		score += 100 // Used today
	} else if age < 24*7 {
		score += 50 // Used this week
	} else if age < 24*30 {
		score += 20 // Used this month
	}

	// Date prefix bonus
	if e.HasDate {
		score += 10
	}

	return score
}

// TriesPath returns the configured tries directory path.
func TriesPath() string {
	if path := os.Getenv("TRY_PATH"); path != "" {
		return expandHome(path)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "src", "tries")
}

// ProjectsPath returns the configured projects directory path.
func ProjectsPath() string {
	if path := os.Getenv("TRY_PROJECTS"); path != "" {
		return expandHome(path)
	}
	return filepath.Dir(TriesPath())
}

func expandHome(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
