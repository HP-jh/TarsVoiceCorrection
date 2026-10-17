package corrector

import (
	"strings"
	"sync"

	"github.com/HP-jh/TarsVoiceCorrection/internal/dict"
)

// SynonymMapper provides synonym lookup and expansion.
type SynonymMapper struct {
	mu       sync.RWMutex
	toSyns   map[string][]string // word -> synonyms
	fromSyns map[string][]string // synonym -> original words (reverse index)
}

// NewSynonymMapper creates a synonym mapper with built-in dictionary.
func NewSynonymMapper() *SynonymMapper {
	sm := &SynonymMapper{
		toSyns:   make(map[string][]string),
		fromSyns: make(map[string][]string),
	}
	sm.loadBuiltins()
	return sm
}

func (sm *SynonymMapper) loadBuiltins() {
	for word, syns := range dict.BuiltinSynonyms {
		sm.toSyns[word] = syns
		for _, syn := range syns {
			sm.fromSyns[syn] = append(sm.fromSyns[syn], word)
		}
	}
}

// AddCustom adds custom synonym entries.
func (sm *SynonymMapper) AddCustom(word string, synonyms []string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.toSyns[word] = synonyms
	for _, syn := range synonyms {
		sm.fromSyns[syn] = append(sm.fromSyns[syn], word)
	}
}

// Lookup finds synonyms for a word.
func (sm *SynonymMapper) Lookup(word string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	word = strings.TrimSpace(word)
	var result []string
	if syns, ok := sm.toSyns[word]; ok {
		result = append(result, syns...)
	}
	if originals, ok := sm.fromSyns[word]; ok {
		result = append(result, originals...)
	}
	return deduplicate(result)
}

// ReverseLookup finds words that have the given word as a synonym.
func (sm *SynonymMapper) ReverseLookup(word string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	word = strings.TrimSpace(word)
	return sm.fromSyns[word]
}

// Expand replaces words in text with their canonical forms (first synonym).
func (sm *SynonymMapper) Expand(text string) string {
	// Simple word-by-word replacement
	words := strings.Fields(text)
	for i, w := range words {
		if syns := sm.Lookup(w); len(syns) > 0 {
			words[i] = syns[0]
		}
	}
	return strings.Join(words, " ")
}

// AllWords returns all known words for dictionary building.
func (sm *SynonymMapper) AllWords() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	seen := make(map[string]struct{})
	var words []string
	for w := range sm.toSyns {
		if _, ok := seen[w]; !ok {
			seen[w] = struct{}{}
			words = append(words, w)
		}
	}
	for w := range sm.fromSyns {
		if _, ok := seen[w]; !ok {
			seen[w] = struct{}{}
			words = append(words, w)
		}
	}
	return words
}

func deduplicate(ss []string) []string {
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
