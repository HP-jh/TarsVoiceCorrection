package corrector

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// CJKRegex matches CJK characters.
	CJKRegex = regexp.MustCompile(`[\x{4e00}-\x{9fff}\x{3400}-\x{4dbf}\x{20000}-\x{2a6df}\x{2a700}-\x{2b73f}\x{2b740}-\x{2b81f}\x{2b820}-\x{2ceaf}\x{f900}-\x{faff}]`)
	// LatinRegex matches Latin letters.
	LatinRegex = regexp.MustCompile(`[a-zA-Z]+`)
	// DigitRegex matches digits.
	DigitRegex = regexp.MustCompile(`\d+`)
	// SpaceAroundCJKEnglish inserts spaces between CJK and English/digits.
	mixedLangPattern = regexp.MustCompile(`([\x{4e00}-\x{9fff}])([a-zA-Z0-9])|([a-zA-Z0-9])([\x{4e00}-\x{9fff}])`)
)

// LanguageType identifies the dominant language of text.
type LanguageType int

const (
	LangUnknown LanguageType = iota
	LangChinese
	LangEnglish
	LangMixed
	LangNumeric
)

// DetectLanguage detects the primary language of text.
func DetectLanguage(text string) LanguageType {
	cjkCount := len(CJKRegex.FindAllString(text, -1))
	latinCount := len(LatinRegex.FindAllString(text, -1))
	digitCount := len(DigitRegex.FindAllString(text, -1))
	total := utf8Len(text)
	if total == 0 {
		return LangUnknown
	}

	if cjkCount > 0 && latinCount == 0 && digitCount == 0 {
		return LangChinese
	}
	if cjkCount == 0 && latinCount > 0 {
		return LangEnglish
	}
	if cjkCount == 0 && latinCount == 0 && digitCount > 0 {
		return LangNumeric
	}
	return LangMixed
}

// NormalizeMixedLang inserts spaces between CJK and English/number boundaries
// and normalizes consecutive whitespace.
func NormalizeMixedLang(text string) string {
	// Insert space between CJK and ASCII
	text = mixedLangPattern.ReplaceAllString(text, "$1$3 $2$4")
	// Normalize whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// ExtractCJK extracts CJK segments from mixed text.
func ExtractCJK(text string) []string {
	return CJKRegex.FindAllString(text, -1)
}

// ExtractLatin extracts Latin word segments.
func ExtractLatin(text string) []string {
	return LatinRegex.FindAllString(text, -1)
}

// IsCJK checks if a rune is CJK.
func IsCJK(r rune) bool {
	return unicode.Is(unicode.Han, r)
}

// IsLatin checks if a rune is Latin letter.
func IsLatin(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// IsDigit checks if a rune is digit.
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// IsPunctuation checks if a rune is punctuation.
func IsPunctuation(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

// SplitMixedText splits mixed CJK-English text into segments.
func SplitMixedText(text string) []string {
	var segments []string
	var current strings.Builder
	var lastType int // 0=none, 1=CJK, 2=Latin, 3=other

	for _, r := range text {
		var t int
		switch {
		case IsCJK(r):
			t = 1
		case IsLatin(r) || IsDigit(r):
			t = 2
		default:
			t = 3
		}

		if t != lastType && current.Len() > 0 {
			segments = append(segments, current.String())
			current.Reset()
		}
		current.WriteRune(r)
		lastType = t
	}
	if current.Len() > 0 {
		segments = append(segments, current.String())
	}
	return segments
}

func utf8Len(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
