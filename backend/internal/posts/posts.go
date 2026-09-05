package posts

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type Post struct {
	ID          int64     `json:"id"`
	AuthorID    int64     `json:"authorId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Status      Status    `json:"status"`
	Downloads   int64     `json:"downloads"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Store struct {
	mu     sync.RWMutex
	nextID int64
	items  map[int64]Post
}

func NewStore() *Store {
	return &Store{
		nextID: 1,
		items:  make(map[int64]Post),
	}
}

func (s *Store) Create(authorID int64, title, description, category string) (Post, error) {
	title = strings.TrimSpace(title)
	if authorID <= 0 || title == "" {
		return Post{}, errors.New("author and title are required")
	}

	now := time.Now().UTC()
	post := Post{
		ID:          s.nextID,
		AuthorID:    authorID,
		Title:       title,
		Description: strings.TrimSpace(description),
		Category:    strings.TrimSpace(category),
		Status:      StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	s.items[post.ID] = post
	return post, nil
}

func (s *Store) List(status Status) []Post {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Post, 0, len(s.items))
	for _, post := range s.items {
		if status == "" || post.Status == status {
			result = append(result, post)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID > result[j].ID
	})
	return result
}

func (s *Store) Find(id int64) (Post, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, ok := s.items[id]
	return post, ok
}

func (s *Store) SetStatus(id int64, status Status) (Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, ok := s.items[id]
	if !ok {
		return Post{}, errors.New("post not found")
	}

	post.Status = status
	post.UpdatedAt = time.Now().UTC()
	s.items[id] = post
	return post, nil
}

func (s *Store) AddDownload(id int64) (Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, ok := s.items[id]
	if !ok {
		return Post{}, errors.New("post not found")
	}

	post.Downloads++
	post.UpdatedAt = time.Now().UTC()
	s.items[id] = post
	return post, nil
}
