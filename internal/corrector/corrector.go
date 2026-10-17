package corrector

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/HP-jh/TarsVoiceCorrection/internal/dict"
)

// CorrectionResult is the output of a correction operation.
type CorrectionResult struct {
	Original   string            `json:"original"`
	Corrected  string            `json:"corrected"`
	Confidence float64           `json:"confidence"`
	Candidates []Candidate       `json:"candidates"`
	Strategies []StrategyDetail  `json:"strategies"`
	Message    string            `json:"message"`
}

// Candidate represents a correction candidate.
type Candidate struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

// StrategyDetail describes what a strategy found.
type StrategyDetail struct {
	Name     string   `json:"name"`
	Score    float64  `json:"score"`
	Matched  []string `json:"matched,omitempty"`
	Action   string   `json:"action"`
}

// Corrector is the main voice correction engine.
type Corrector struct {
	mu         sync.RWMutex
	synonyms   *SynonymMapper
	pinyin     *PinyinMatcher
	opencc     *OpenCCConverter
	customDict *CustomDictionary
	wordList   []string
	maxLen     int
}

// NewCorrector creates a new corrector with all strategies initialized.
func NewCorrector() (*Corrector, error) {
	sm := NewSynonymMapper()
	wl := sm.AllWords()
	wl = append(wl, dict.CommonWords...)
	wl = deduplicateStrings(wl)

	pm := NewPinyinMatcher(wl)

	occ, err := NewOpenCCConverter()
	if err != nil {
		return nil, fmt.Errorf("opencc init: %w", err)
	}

	cd := NewCustomDictionary(DefaultCustomDictPath())
	_ = cd.Load() // non-fatal

	c := &Corrector{
		synonyms:   sm,
		pinyin:     pm,
		opencc:     occ,
		customDict: cd,
		wordList:   wl,
		maxLen:     1000,
	}
	return c, nil
}

// Correct applies all correction strategies and returns the best result.
func (c *Corrector) Correct(text string) CorrectionResult {
	if len(text) > c.maxLen {
		text = text[:c.maxLen]
	}
	original := text

	result := CorrectionResult{
		Original:   original,
		Corrected:  original,
		Confidence: 1.0,
		Candidates: nil,
		Strategies: []StrategyDetail{},
	}

	// Strategy 1: OpenCC - normalize Simplified/Traditional
	s2tText, _ := c.opencc.AutoConvert(text)
	if s2tText != text {
		result.Strategies = append(result.Strategies, StrategyDetail{
			Name:    "opencc",
			Score:   0.95,
			Action:  "normalized_s2t",
			Matched: []string{text, s2tText},
		})
		text = s2tText
	}

	// Strategy 2: Mixed language normalization
	normalized := NormalizeMixedLang(text)
	if normalized != text {
		result.Strategies = append(result.Strategies, StrategyDetail{
			Name:    "mixed_lang",
			Score:   0.90,
			Action:  "normalized_spacing",
			Matched: []string{text, normalized},
		})
		text = normalized
	}

	// Strategy 3: Word-level correction (fuzzy + pinyin + synonym)
	words := tokenize(text)
	var correctedWords []string
	var totalConf float64
	var wordCount int

	for _, w := range words {
		cw, conf, strategy := c.correctWord(w)
		correctedWords = append(correctedWords, cw)
		totalConf += conf
		wordCount++
		if strategy != "" {
			result.Strategies = append(result.Strategies, StrategyDetail{
				Name:    strategy,
				Score:   conf,
				Matched: []string{w, cw},
				Action:  "word_corrected",
			})
		}
	}

	correctedText := strings.Join(correctedWords, "")
	// Re-normalize spaces for readability
	correctedText = regexp.MustCompile(`\s+`).ReplaceAllString(correctedText, " ")
	correctedText = strings.TrimSpace(correctedText)

	result.Corrected = correctedText

	// Confidence calculation
	if wordCount > 0 {
		result.Confidence = totalConf / float64(wordCount)
	}

	// Generate candidates
	candidates := c.generateCandidates(original, correctedText)
	result.Candidates = candidates

	// Human-friendly message
	if result.Corrected != result.Original {
		if len(candidates) > 1 {
			result.Message = fmt.Sprintf("您是不是想说「%s」？其它候选：", candidates[0].Text)
			for i := 1; i < minInt(len(candidates), 4); i++ {
				result.Message += fmt.Sprintf("%s(%.0f%%) ", candidates[i].Text, candidates[i].Confidence*100)
			}
		} else {
			result.Message = fmt.Sprintf("已纠错为「%s」，置信度 %.0f%%", result.Corrected, result.Confidence*100)
		}
	} else {
		result.Message = "文本看起来没问题，无需纠错"
	}

	return result
}

// correctWord applies strategies to a single word.
func (c *Corrector) correctWord(word string) (corrected string, confidence float64, strategy string) {
	// Skip pure punctuation/space
	if len(strings.TrimSpace(word)) == 0 {
		return word, 1.0, ""
	}

	// Strategy A: Custom dictionary exact match (highest priority)
	if c.customDict.Has(word) {
		return word, 1.0, "custom_dict"
	}

	// Strategy B: Synonym lookup
	if syns := c.synonyms.Lookup(word); len(syns) > 0 {
		return syns[0], 0.85, "synonym"
	}

	// Strategy C: Fuzzy search against word list
	if matches := FuzzySearch(word, c.wordList, 3, 0.6); len(matches) > 0 {
		best := matches[0]
		if best.Score >= 0.75 {
			return best.Word, best.Score, "fuzzy"
		}
	}

	// Strategy D: Pinyin homophone matching
	if homophones := c.pinyin.FindHomophones(word, 3, 0.75); len(homophones) > 0 {
		best := homophones[0]
		if best.Score >= 0.75 {
			return best.Word, best.Score * 0.9, "pinyin"
		}
	}

	return word, 1.0, ""
}

// generateCandidates produces top-N correction candidates with confidence.
func (c *Corrector) generateCandidates(original, corrected string) []Candidate {
	cands := []Candidate{{Text: corrected, Confidence: 1.0, Source: "combined"}}

	// Add OpenCC variant
	if t2s, _ := c.opencc.T2S(original); t2s != original && t2s != corrected {
		cands = append(cands, Candidate{Text: t2s, Confidence: 0.92, Source: "opencc"})
	}

	// Add fuzzy matches for the whole text
	if matches := FuzzySearch(original, c.wordList, 3, 0.5); len(matches) > 0 {
		for _, m := range matches {
			if m.Word != corrected {
				cands = append(cands, Candidate{Text: m.Word, Confidence: m.Score, Source: "fuzzy"})
			}
		}
	}

	// Add pinyin homophones
	if homos := c.pinyin.FindHomophones(original, 3, 0.5); len(homos) > 0 {
		for _, h := range homos {
			found := false
			for _, c := range cands {
				if c.Text == h.Word {
					found = true
					break
				}
			}
			if !found {
				cands = append(cands, Candidate{Text: h.Word, Confidence: h.Score * 0.9, Source: "pinyin"})
			}
		}
	}

	// Sort by confidence descending
	for i := 0; i < len(cands); i++ {
		for j := i + 1; j < len(cands); j++ {
			if cands[j].Confidence > cands[i].Confidence {
				cands[i], cands[j] = cands[j], cands[i]
			}
		}
	}

	// Deduplicate
	seen := make(map[string]struct{})
	var out []Candidate
	for _, cand := range cands {
		if _, ok := seen[cand.Text]; !ok {
			seen[cand.Text] = struct{}{}
			out = append(out, cand)
		}
	}
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

// tokenize splits text into word-like segments.
func tokenize(text string) []string {
	var words []string
	var current strings.Builder
	for _, r := range text {
		if IsCJK(r) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			current.WriteRune(r)
		} else if IsLatin(r) || IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			words = append(words, string(r))
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

// PinyinMatch finds homophones for a word.
func (c *Corrector) PinyinMatch(word string, topK int) []MatchResult {
	return c.pinyin.FindHomophones(word, topK, 0.6)
}

// SynonymLookup finds synonyms for a word.
func (c *Corrector) SynonymLookup(word string) []string {
	return c.synonyms.Lookup(word)
}

// FuzzySearch performs fuzzy search against the word list.
func (c *Corrector) FuzzySearch(query string, topK int) []MatchResult {
	return FuzzySearch(query, c.wordList, topK, 0.35)
}

// LangDetect detects the language of text.
func (c *Corrector) LangDetect(text string) string {
	lt := DetectLanguage(text)
	switch lt {
	case LangChinese:
		return "zh"
	case LangEnglish:
		return "en"
	case LangMixed:
		return "mixed"
	case LangNumeric:
		return "numeric"
	default:
		return "unknown"
	}
}

// AddCustomWord adds a word to the custom dictionary.
func (c *Corrector) AddCustomWord(word string) error {
	c.customDict.Add(word)
	return c.customDict.Save()
}

func deduplicateStrings(ss []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, s := range ss {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
