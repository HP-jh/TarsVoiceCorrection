package main

import (
	"testing"

	"github.com/HP-jh/TarsVoiceCorrection/internal/corrector"
)

func FuzzCorrectorCorrect(f *testing.F) {
	c, err := corrector.NewCorrector()
	if err != nil {
		f.Fatalf("NewCorrector: %v", err)
	}

	// Seed corpus
	f.Add("电脑")
	f.Add("中过")
	f.Add("Hello世界")
	f.Add("電腦")
	f.Add("")
	f.Add("123")
	f.Add("<script>alert(1)</script>")
	f.Add("../../../etc/passwd")
	f.Add("🎉🎊")
	f.Add("a" + string(make([]byte, 1000)))

	f.Fuzz(func(t *testing.T, text string) {
		// Should not panic on any input
		result := c.Correct(text)
		_ = result.Corrected
		_ = result.Confidence
		_ = result.Message
	})
}

func FuzzPinyinMatch(f *testing.F) {
	c, _ := corrector.NewCorrector()
	f.Add("中国", 3)
	f.Add("", 5)
	f.Add("<script>", 1)
	f.Fuzz(func(t *testing.T, word string, topK int) {
		if topK < 0 || topK > 100 {
			return
		}
		matches := c.PinyinMatch(word, topK)
		_ = matches
	})
}

func FuzzFuzzySearch(f *testing.F) {
	c, _ := corrector.NewCorrector()
	f.Add("电恼", 3)
	f.Add("", 1)
	f.Fuzz(func(t *testing.T, query string, topK int) {
		if topK < 0 || topK > 100 {
			return
		}
		matches := c.FuzzySearch(query, topK)
		_ = matches
	})
}
