package reputation

import (
	"sort"
	"sync"
	"time"
)

type EventType string

const (
	EventRegister       EventType = "register"
	EventUploadClean    EventType = "upload_clean"
	EventUploadRejected EventType = "upload_rejected"
	EventPostApproved   EventType = "post_approved"
	EventPostRejected   EventType = "post_rejected"
	EventDownload       EventType = "download"
)

type Event struct {
	UserID    int64     `json:"userId"`
	Type      EventType `json:"type"`
	Delta     int       `json:"delta"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
}

type Profile struct {
	UserID    int64   `json:"userId"`
	Score     int     `json:"score"`
	Level     string  `json:"level"`
	Downloads int64   `json:"downloads"`
	Events    []Event `json:"events"`
}

type Store struct {
	mu        sync.RWMutex
	scores    map[int64]int
	downloads map[int64]int64
	events    map[int64][]Event
}

func NewStore() *Store {
	return &Store{
		scores:    make(map[int64]int),
		downloads: make(map[int64]int64),
		events:    make(map[int64][]Event),
	}
}

func (s *Store) Add(userID int64, eventType EventType, delta int, reason string) Profile {
	if userID <= 0 {
		return Profile{}
	}

	event := Event{
		UserID:    userID,
		Type:      eventType,
		Delta:     delta,
		Reason:    reason,
		CreatedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.scores[userID] += delta
	if eventType == EventDownload {
		s.downloads[userID]++
	}
	s.events[userID] = append(s.events[userID], event)
	return s.profileLocked(userID)
}

func (s *Store) Get(userID int64) Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.profileLocked(userID)
}

func (s *Store) List() []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Profile, 0, len(s.scores))
	for userID := range s.scores {
		items = append(items, s.profileLocked(userID))
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})
	return items
}

func (s *Store) profileLocked(userID int64) Profile {
	events := append([]Event(nil), s.events[userID]...)
	return Profile{
		UserID:    userID,
		Score:     s.scores[userID],
		Level:     level(s.scores[userID]),
		Downloads: s.downloads[userID],
		Events:    events,
	}
}

func level(score int) string {
	switch {
	case score >= 100:
		return "trusted"
	case score >= 30:
		return "active"
	case score < 0:
		return "restricted"
	default:
		return "new"
	}
}
