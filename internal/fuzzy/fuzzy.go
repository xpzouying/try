package fuzzy

import (
	"strings"
	"unicode"
)

// Match represents a fuzzy match result.
type Match struct {
	Text       string
	Score      float64
	Positions  []int // Matched character positions
	StartIndex int   // Original index in the list
}

// Search performs fuzzy matching on a list of strings.
// Returns matches sorted by score (highest first).
func Search(query string, items []string) []Match {
	if query == "" {
		result := make([]Match, len(items))
		for i, item := range items {
			result[i] = Match{Text: item, Score: 0, StartIndex: i}
		}
		return result
	}

	query = strings.ToLower(query)
	var matches []Match

	for i, item := range items {
		score, positions := score(query, strings.ToLower(item))
		if score > 0 {
			matches = append(matches, Match{
				Text:       item,
				Score:      score,
				Positions:  positions,
				StartIndex: i,
			})
		}
	}

	// Sort by score descending
	sortByScore(matches)
	return matches
}

// score calculates the fuzzy match score between query and text.
// Returns 0 if no match.
func score(query, text string) (float64, []int) {
	if len(query) == 0 {
		return 0, nil
	}
	if len(text) == 0 {
		return 0, nil
	}

	var positions []int
	queryRunes := []rune(query)
	textRunes := []rune(text)
	queryIdx := 0
	lastMatchPos := -1
	totalScore := 0.0

	for i, r := range textRunes {
		if queryIdx >= len(queryRunes) {
			break
		}

		if r == queryRunes[queryIdx] {
			positions = append(positions, i)

			// Base score for match
			matchScore := 1.0

			// Bonus for word boundary match
			if i == 0 || !unicode.IsLetter(textRunes[i-1]) || unicode.IsUpper(rune(text[i])) {
				matchScore += 1.0
			}

			// Bonus for consecutive matches
			if lastMatchPos >= 0 && i == lastMatchPos+1 {
				matchScore += 2.0
			}

			totalScore += matchScore
			lastMatchPos = i
			queryIdx++
		}
	}

	// All query characters must be matched
	if queryIdx < len(queryRunes) {
		return 0, nil
	}

	// Penalty for length (prefer shorter matches)
	totalScore *= float64(len(queryRunes)) / float64(len(textRunes)+10)

	return totalScore, positions
}

func sortByScore(matches []Match) {
	// Simple insertion sort (good enough for small lists)
	for i := 1; i < len(matches); i++ {
		j := i
		for j > 0 && matches[j].Score > matches[j-1].Score {
			matches[j], matches[j-1] = matches[j-1], matches[j]
			j--
		}
	}
}
