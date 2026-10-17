package corrector

import (
	"strings"

	"github.com/HP-jh/TarsVoiceCorrection/internal/pinyin"
)

// PinyinMatcher provides pinyin-based matching for homophone correction.
type PinyinMatcher struct {
	cache map[string]string // word -> pinyin string (toneless)
	dict  map[string][]string // pinyin -> words
}

// NewPinyinMatcher creates a new pinyin matcher with a word dictionary.
func NewPinyinMatcher(words []string) *PinyinMatcher {
	pm := &PinyinMatcher{
		cache: make(map[string]string),
		dict:  make(map[string][]string),
	}
	for _, w := range words {
		pyStr := pm.toPinyin(w)
		if pyStr == "" {
			continue
		}
		pm.cache[w] = pyStr
		pm.dict[pyStr] = append(pm.dict[pyStr], w)
	}
	return pm
}

func (pm *PinyinMatcher) toPinyin(s string) string {
	var parts []string
	for _, r := range s {
		if py := pinyin.Of(r); py != "" {
			parts = append(parts, py)
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			parts = append(parts, string(r))
		}
	}
	return strings.Join(parts, " ")
}

// PinyinOf returns the pinyin representation of a word.
func (pm *PinyinMatcher) PinyinOf(word string) string {
	if py, ok := pm.cache[word]; ok {
		return py
	}
	return pm.toPinyin(word)
}

// FindHomophones finds words with similar pinyin to the input.
func (pm *PinyinMatcher) FindHomophones(word string, topK int, threshold float64) []MatchResult {
	inputPy := pm.PinyinOf(word)
	if inputPy == "" {
		return nil
	}

	type scored struct {
		word  string
		score float64
	}
	var results []scored

	for w, py := range pm.cache {
		if w == word {
			continue
		}
		sim := NormalizedSimilarity(inputPy, py)
		if sim >= threshold {
			results = append(results, scored{word: w, score: sim})
		}
	}

	// Sort descending
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

// PinyinDistance computes edit distance between pinyin strings.
func (pm *PinyinMatcher) PinyinDistance(a, b string) int {
	pyA := pm.PinyinOf(a)
	pyB := pm.PinyinOf(b)
	return DamerauLevenshtein(pyA, pyB)
}

// IsHomophone checks if two words have similar pinyin (distance <= threshold).
func (pm *PinyinMatcher) IsHomophone(a, b string, maxDist int) bool {
	return pm.PinyinDistance(a, b) <= maxDist
}
