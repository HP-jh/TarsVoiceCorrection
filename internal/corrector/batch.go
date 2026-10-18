package corrector

import (
	"context"
	"sync"
)

// BatchItem represents a single text to be corrected in a batch.
type BatchItem struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// BatchResult is the correction result for a single batch item.
type BatchResult struct {
	ID     string           `json:"id"`
	Result CorrectionResult `json:"result"`
}

// BatchCorrect processes multiple texts concurrently with a configurable worker pool.
// The context can be used to cancel the operation mid-way.
func (c *Corrector) BatchCorrect(ctx context.Context, items []BatchItem, workers int) []BatchResult {
	if workers <= 0 {
		workers = 4
	}
	if workers > len(items) {
		workers = len(items)
	}

	results := make([]BatchResult, len(items))
	itemCh := make(chan int, len(items))
	for i := range items {
		itemCh <- i
	}
	close(itemCh)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case idx, ok := <-itemCh:
					if !ok {
						return
					}
					results[idx] = BatchResult{
						ID:     items[idx].ID,
						Result: c.Correct(items[idx].Text),
					}
				}
			}
		}()
	}
	wg.Wait()
	return results
}
