package blocklist

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	SHA256    string    `json:"sha256"`
	Reason    string    `json:"reason"`
	CreatedBy int64     `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

func NewStore() *Store {
	return &Store{entries: make(map[string]Entry)}
}

func (s *Store) Add(sha256, reason string, createdBy int64) (Entry, error) {
	sha256 = normalizeSHA256(sha256)
	reason = strings.TrimSpace(reason)
	if sha256 == "" || reason == "" {
		return Entry{}, errors.New("sha256 and reason are required")
	}

	entry := Entry{
		SHA256:    sha256,
		Reason:    reason,
		CreatedBy: createdBy,
		CreatedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[sha256] = entry
	return entry, nil
}

func (s *Store) Contains(sha256 string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[normalizeSHA256(sha256)]
	return entry, ok
}

func (s *Store) List() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Entry, 0, len(s.entries))
	for _, entry := range s.entries {
		items = append(items, entry)
	}
	return items
}

func normalizeSHA256(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != 64 {
		return ""
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return ""
		}
	}
	return value
}
