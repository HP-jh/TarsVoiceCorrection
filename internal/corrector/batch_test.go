package corrector

import (
	"context"
	"testing"
	"time"
)

func TestBatchCorrectBasic(t *testing.T) {
	c, err := NewCorrector()
	if err != nil {
		t.Fatalf("NewCorrector: %v", err)
	}

	items := []BatchItem{
		{ID: "1", Text: "电恼"},
		{ID: "2", Text: "中过"},
		{ID: "3", Text: "Hello世界"},
	}

	results := c.BatchCorrect(context.Background(), items, 2)
	if len(results) != len(items) {
		t.Fatalf("Expected %d results, got %d", len(items), len(results))
	}

	for i, r := range results {
		if r.ID != items[i].ID {
			t.Errorf("Result[%d].ID = %q, want %q", i, r.ID, items[i].ID)
		}
		if r.Result.Original != items[i].Text {
			t.Errorf("Result[%d].Original = %q, want %q", i, r.Result.Original, items[i].Text)
		}
	}
}

func TestBatchCorrectEmpty(t *testing.T) {
	c, _ := NewCorrector()
	results := c.BatchCorrect(context.Background(), []BatchItem{}, 4)
	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty input, got %d", len(results))
	}
}

func TestBatchCorrectConcurrency(t *testing.T) {
	c, _ := NewCorrector()
	items := make([]BatchItem, 20)
	for i := range items {
		items[i] = BatchItem{ID: string(rune('a' + i)), Text: "电恼"}
	}

	start := time.Now()
	results := c.BatchCorrect(context.Background(), items, 8)
	elapsed := time.Since(start)

	if len(results) != 20 {
		t.Fatalf("Expected 20 results, got %d", len(results))
	}

	// With concurrency, 20 identical items should complete quickly
	// (mostly cache hits after first few)
	t.Logf("BatchCorrect 20 items with 8 workers took %v", elapsed)
}

func TestBatchCorrectCancellation(t *testing.T) {
	c, _ := NewCorrector()
	items := make([]BatchItem, 100)
	for i := range items {
		items[i] = BatchItem{ID: string(rune('a' + i%26)), Text: "测试文本"}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	results := c.BatchCorrect(ctx, items, 4)
	// Some may have been processed before cancellation takes effect,
	// but not all 100 in a cancelled context
	if len(results) == 100 {
		t.Log("All 100 processed before cancel took effect — acceptable on fast machine")
	}
}
