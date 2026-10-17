package corrector

import (
	"math"
	"strings"
	"unicode/utf8"
)

// DamerauLevenshtein computes the Damerau-Levenshtein distance between two strings.
// Supports transpositions (swapping adjacent characters).
func DamerauLevenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Full matrix for correct transposition handling
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
	}
	for i := 0; i <= la; i++ {
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			d[i][j] = min3(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)

			// Transposition
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				if trans := d[i-2][j-2] + 1; trans < d[i][j] {
					d[i][j] = trans
				}
			}
		}
	}
	return d[la][lb]
}

// NormalizedSimilarity returns similarity score [0,1] based on edit distance.
func NormalizedSimilarity(a, b string) float64 {
	maxLen := utf8.RuneCountInString(a)
	if l := utf8.RuneCountInString(b); l > maxLen {
		maxLen = l
	}
	if maxLen == 0 {
		return 1.0
	}
	dist := DamerauLevenshtein(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

// NGramTokenize splits text into character n-grams.
func NGramTokenize(text string, n int) []string {
	runes := []rune(text)
	if len(runes) < n {
		return []string{text}
	}
	var grams []string
	for i := 0; i <= len(runes)-n; i++ {
		grams = append(grams, string(runes[i:i+n]))
	}
	return grams
}

// NGramSimilarity computes Jaccard-like similarity using n-grams.
func NGramSimilarity(a, b string, n int) float64 {
	ga := NGramTokenize(a, n)
	gb := NGramTokenize(b, n)
	if len(ga) == 0 && len(gb) == 0 {
		return 1.0
	}
	if len(ga) == 0 || len(gb) == 0 {
		return 0.0
	}
	setA := make(map[string]struct{}, len(ga))
	for _, g := range ga {
		setA[g] = struct{}{}
	}
	intersection := 0
	for _, g := range gb {
		if _, ok := setA[g]; ok {
			intersection++
		}
	}
	return float64(intersection) / float64(len(ga)+len(gb)-intersection)
}

// FuzzySearch finds the best matching candidates from a dictionary.
func FuzzySearch(query string, candidates []string, topK int, threshold float64) []MatchResult {
	type scored struct {
		word  string
		score float64
	}
	var results []scored
	qLower := strings.ToLower(query)
	qRunes := utf8.RuneCountInString(qLower)
	for _, c := range candidates {
		cLower := strings.ToLower(c)
		cRunes := utf8.RuneCountInString(cLower)
		// Combine multiple similarity signals
		s1 := NormalizedSimilarity(qLower, cLower)
		// Use unigrams for short strings, bigrams for longer ones
		n := 2
		if qRunes < 3 || cRunes < 3 {
			n = 1
		}
		s2 := NGramSimilarity(qLower, cLower, n)
		score := s1*0.6 + s2*0.4
		if score >= threshold {
			results = append(results, scored{word: c, score: score})
		}
	}

	// Sort by score descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	out := make([]MatchResult, len(results))
	for i, r := range results {
		out[i] = MatchResult{Word: r.word, Score: r.score}
	}
	return out
}

// MatchResult represents a fuzzy match.
type MatchResult struct {
	Word  string  `json:"word"`
	Score float64 `json:"score"`
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

// ConfidenceFromDistance converts edit distance to a confidence score.
func ConfidenceFromDistance(dist, maxLen int) float64 {
	if maxLen == 0 {
		return 1.0
	}
	ratio := float64(dist) / float64(maxLen)
	return math.Max(0, math.Min(1, 1.0-ratio))
}
