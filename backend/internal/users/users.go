package users

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Role         Role      `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Store struct {
	mu      sync.RWMutex
	nextID  int64
	byID    map[int64]User
	byEmail map[string]int64
}

func NewStore() *Store {
	return &Store{
		nextID:  1,
		byID:    make(map[int64]User),
		byEmail: make(map[string]int64),
	}
}

func (s *Store) Create(username, email, passwordHash string) (User, error) {
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	if username == "" || email == "" || passwordHash == "" {
		return User{}, errors.New("username, email and password are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byEmail[email]; exists {
		return User{}, errors.New("email already exists")
	}

	user := User{
		ID:           s.nextID,
		Username:     username,
		Email:        email,
		Role:         RoleUser,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}
	if len(s.byID) == 0 {
		user.Role = RoleAdmin
	}
	s.nextID++
	s.byID[user.ID] = user
	s.byEmail[user.Email] = user.ID
	return user, nil
}

func (s *Store) FindByEmail(email string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.byEmail[strings.ToLower(strings.TrimSpace(email))]
	if !ok {
		return User{}, false
	}

	user, ok := s.byID[id]
	return user, ok
}

func (s *Store) FindByID(id int64) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.byID[id]
	return user, ok
}
