package corrector

import (
	"strings"
	"testing"
)

func TestDamerauLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "b", 1},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"ca", "abc", 3},
		{"ab", "ba", 1}, // transposition
		{"中国", "中國", 1}, // 国 vs 國 are different characters
		{"电脑", "计算机", 3},
	}
	for _, tt := range tests {
		got := DamerauLevenshtein(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("DamerauLevenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestNormalizedSimilarity(t *testing.T) {
	sim := NormalizedSimilarity("电脑", "电恼")
	if sim < 0.4 || sim > 0.6 {
		t.Errorf("Expected similarity ~0.5 for 电脑/电恼, got %.2f", sim)
	}
	sim2 := NormalizedSimilarity("电脑", "电脑")
	if sim2 != 1.0 {
		t.Errorf("Expected 1.0 for identical, got %.2f", sim2)
	}
}

func TestNGramSimilarity(t *testing.T) {
	sim := NGramSimilarity("电脑", "电恼", 1)
	if sim < 0.3 || sim > 0.7 {
		t.Errorf("Expected n-gram similarity ~0.5, got %.2f", sim)
	}
}

func TestFuzzySearch(t *testing.T) {
	candidates := []string{"电脑", "手机", "网络", "数据", "服务器"}
	results := FuzzySearch("电恼", candidates, 3, 0.3)
	if len(results) == 0 {
		t.Error("Expected at least one fuzzy match for 电恼")
	}
	found := false
	for _, r := range results {
		if r.Word == "电脑" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected '电脑' in fuzzy results, got %v", results)
	}
}

func TestPinyinMatcher(t *testing.T) {
	pm := NewPinyinMatcher([]string{"中国", "中心", "重要", "钟国"})
	py := pm.PinyinOf("中国")
	if py == "" {
		t.Error("Expected pinyin for 中国")
	}
	matches := pm.FindHomophones("中国", 3, 0.5)
	if len(matches) == 0 {
		t.Error("Expected homophone matches")
	}
}

func TestSynonymMapper(t *testing.T) {
	sm := NewSynonymMapper()
	syns := sm.Lookup("电脑")
	found := false
	for _, s := range syns {
		if s == "计算机" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected '计算机' as synonym of '电脑', got %v", syns)
	}

	sm.AddCustom("网关", []string{"Gateway", "入口"})
	customSyns := sm.Lookup("网关")
	if len(customSyns) == 0 {
		t.Error("Expected custom synonyms for 网关")
	}
}

func TestDetectLanguage(t *testing.T) {
	if DetectLanguage("中文文本") != LangChinese {
		t.Error("Expected Chinese")
	}
	if DetectLanguage("Hello world") != LangEnglish {
		t.Error("Expected English")
	}
	if DetectLanguage("Hello 世界") != LangMixed {
		t.Error("Expected Mixed")
	}
}

func TestNormalizeMixedLang(t *testing.T) {
	got := NormalizeMixedLang("Hello世界123")
	if !strings.Contains(got, " ") {
		t.Errorf("Expected spaces in mixed text, got %q", got)
	}
}

func TestCustomDictionary(t *testing.T) {
	cd := NewCustomDictionary("/tmp/test_tars_dict.txt")
	cd.Add("自定义词")
	if !cd.Has("自定义词") {
		t.Error("Expected custom word to be found")
	}
	words := cd.Words()
	if len(words) != 1 || words[0] != "自定义词" {
		t.Errorf("Expected [自定义词], got %v", words)
	}
}

func TestCorrectorCorrect(t *testing.T) {
	c, err := NewCorrector()
	if err != nil {
		t.Fatalf("NewCorrector failed: %v", err)
	}

	// Test basic correction
	result := c.Correct("电恼")
	if result.Corrected != result.Original {
		t.Logf("Corrected '%s' -> '%s' (conf=%.2f)", result.Original, result.Corrected, result.Confidence)
	}

	// Test mixed language
	result2 := c.Correct("Hello世界")
	t.Logf("Mixed: '%s' -> '%s'", result2.Original, result2.Corrected)

	// Test pinyin match
	matches := c.PinyinMatch("中过", 3)
	t.Logf("Pinyin matches for 中过: %v", matches)

	// Test synonym lookup
	syns := c.SynonymLookup("电脑")
	if len(syns) == 0 {
		t.Error("Expected synonyms for 电脑")
	}

	// Test fuzzy search
	fuzzy := c.FuzzySearch("电恼", 3)
	if len(fuzzy) == 0 {
		t.Error("Expected fuzzy matches")
	}

	// Test language detection
	lang := c.LangDetect("中文")
	if lang != "zh" {
		t.Errorf("Expected zh, got %s", lang)
	}
}

func TestCorrectorConfidence(t *testing.T) {
	c, _ := NewCorrector()
	result := c.Correct("电脑")
	if result.Confidence < 0.8 {
		t.Errorf("Expected high confidence for correct text, got %.2f", result.Confidence)
	}
}

func TestSplitMixedText(t *testing.T) {
	segs := SplitMixedText("Hello世界123")
	if len(segs) < 3 {
		t.Errorf("Expected at least 3 segments, got %d: %v", len(segs), segs)
	}
}

func TestTokenize(t *testing.T) {
	words := tokenize("Hello 世界 123")
	if len(words) < 3 {
		t.Errorf("Expected at least 3 tokens, got %d: %v", len(words), words)
	}
}

func BenchmarkDamerauLevenshtein(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DamerauLevenshtein("kitten", "sitting")
	}
}
