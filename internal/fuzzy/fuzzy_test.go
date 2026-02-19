package fuzzy

import (
	"testing"
)

func TestSearch_EmptyQuery(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	matches := Search("", items)

	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(matches))
	}
	for i, m := range matches {
		if m.Text != items[i] {
			t.Errorf("expected %s at index %d, got %s", items[i], i, m.Text)
		}
		if m.Score != 0 {
			t.Errorf("expected score 0 for empty query, got %f", m.Score)
		}
	}
}

func TestSearch_ExactMatch(t *testing.T) {
	items := []string{"redis", "postgres", "mysql"}
	matches := Search("redis", items)

	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].Text != "redis" {
		t.Errorf("expected redis as first match, got %s", matches[0].Text)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	items := []string{"Redis", "POSTGRES", "MySQL"}
	matches := Search("redis", items)

	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].Text != "Redis" {
		t.Errorf("expected Redis as first match, got %s", matches[0].Text)
	}
}

func TestSearch_PrefixMatch(t *testing.T) {
	items := []string{"redis-cache", "redis-queue", "postgres"}
	matches := Search("red", items)

	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestSearch_SubstringMatch(t *testing.T) {
	items := []string{"my-redis-app", "postgres", "redis"}
	matches := Search("redis", items)

	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
	// "redis" should score higher than "my-redis-app" due to length penalty
	if matches[0].Text != "redis" {
		t.Errorf("expected redis as first match, got %s", matches[0].Text)
	}
}

func TestSearch_NoMatch(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	matches := Search("xyz", items)

	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestSearch_FuzzyMatch(t *testing.T) {
	items := []string{"2024-01-15-redis", "2024-01-15-postgres", "2024-01-15-mysql"}
	matches := Search("rds", items)

	// "rds" should match "redis" (r-e-d-i-s contains r, d, s)
	if len(matches) == 0 {
		t.Fatal("expected fuzzy match for 'rds' in 'redis'")
	}
	if matches[0].Text != "2024-01-15-redis" {
		t.Errorf("expected redis entry as first match, got %s", matches[0].Text)
	}
}

func TestSearch_Positions(t *testing.T) {
	items := []string{"redis"}
	matches := Search("rds", items)

	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	// Positions should be indices of 'r', 'd', 's' in "redis"
	// r=0, e=1, d=2, i=3, s=4
	expected := []int{0, 2, 4}
	if len(matches[0].Positions) != len(expected) {
		t.Fatalf("expected %d positions, got %d", len(expected), len(matches[0].Positions))
	}
	for i, pos := range matches[0].Positions {
		if pos != expected[i] {
			t.Errorf("expected position %d at index %d, got %d", expected[i], i, pos)
		}
	}
}

func TestSearch_ConsecutiveBonus(t *testing.T) {
	items := []string{"abcdef", "aXbXcXdXeXf"}
	matches := Search("abc", items)

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	// "abcdef" should score higher due to consecutive match bonus
	if matches[0].Text != "abcdef" {
		t.Errorf("expected abcdef to score higher due to consecutive bonus, got %s first", matches[0].Text)
	}
}

func TestSearch_WordBoundaryBonus(t *testing.T) {
	items := []string{"my-redis-app", "myxredisxapp"}
	matches := Search("redis", items)

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	// "my-redis-app" should score higher due to word boundary bonus
	if matches[0].Text != "my-redis-app" {
		t.Errorf("expected my-redis-app to score higher due to word boundary, got %s first", matches[0].Text)
	}
}

func TestSearch_StartIndexPreserved(t *testing.T) {
	items := []string{"zzz", "aaa", "mmm"}
	matches := Search("a", items)

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].StartIndex != 1 {
		t.Errorf("expected StartIndex 1, got %d", matches[0].StartIndex)
	}
}

func TestSearch_EmptyItems(t *testing.T) {
	matches := Search("test", []string{})

	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty items, got %d", len(matches))
	}
}

func TestSearch_SortedByScore(t *testing.T) {
	items := []string{"xxxredisxxx", "redis", "redisabc"}
	matches := Search("redis", items)

	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(matches))
	}

	// Verify sorted by score descending
	for i := 1; i < len(matches); i++ {
		if matches[i].Score > matches[i-1].Score {
			t.Errorf("matches not sorted by score: %f > %f at index %d",
				matches[i].Score, matches[i-1].Score, i)
		}
	}
}
