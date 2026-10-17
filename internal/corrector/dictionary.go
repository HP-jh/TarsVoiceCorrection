package corrector

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// CustomDictionary manages user-defined custom words.
type CustomDictionary struct {
	mu     sync.RWMutex
	words  map[string]struct{}
	path   string
	loaded bool
}

// NewCustomDictionary creates a custom dictionary.
func NewCustomDictionary(path string) *CustomDictionary {
	return &CustomDictionary{
		words: make(map[string]struct{}),
		path:  path,
	}
}

// Load reads custom words from file.
func (cd *CustomDictionary) Load() error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	f, err := os.Open(cd.path)
	if err != nil {
		if os.IsNotExist(err) {
			cd.loaded = true
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			cd.words[word] = struct{}{}
		}
	}
	cd.loaded = true
	return scanner.Err()
}

// Add adds a word to the custom dictionary (in-memory only).
func (cd *CustomDictionary) Add(word string) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	cd.words[strings.TrimSpace(word)] = struct{}{}
}

// Has checks if a word is in the custom dictionary.
func (cd *CustomDictionary) Has(word string) bool {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	_, ok := cd.words[word]
	return ok
}

// Words returns all custom words.
func (cd *CustomDictionary) Words() []string {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	var out []string
	for w := range cd.words {
		out = append(out, w)
	}
	return out
}

// Save persists the dictionary to disk.
func (cd *CustomDictionary) Save() error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	dir := filepath.Dir(cd.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(cd.path)
	if err != nil {
		return err
	}
	defer f.Close()

	for w := range cd.words {
		if _, err := f.WriteString(w + "\n"); err != nil {
			return err
		}
	}
	return nil
}

// DefaultCustomDictPath returns the default path for custom dictionary.
func DefaultCustomDictPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".tars", "dict", "custom.txt")
}
