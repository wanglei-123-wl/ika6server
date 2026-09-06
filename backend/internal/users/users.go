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
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Role         Role       `json:"role"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"createdAt"`
	BannedUntil  *time.Time `json:"bannedUntil,omitempty"`
	BanReason    string     `json:"banReason,omitempty"`
}

type Repository interface {
	Create(username, email, passwordHash string) (User, error)
	FindByEmail(email string) (User, bool)
	FindByID(id int64) (User, bool)
	Ban(id int64, until time.Time, reason string) (User, error)
	IsBanned(user User, now time.Time) bool
	Counts(now time.Time) (total, active int)
}

type Store struct {
	mu                  sync.RWMutex
	nextID              int64
	byID                map[int64]User
	byEmail             map[string]int64
	bootstrapAdminEmail string
}

var _ Repository = (*Store)(nil)

func NewStore() *Store {
	return NewStoreWithAdmin("")
}

func NewStoreWithAdmin(email string) *Store {
	return &Store{
		nextID:              1,
		byID:                make(map[int64]User),
		byEmail:             make(map[string]int64),
		bootstrapAdminEmail: strings.ToLower(strings.TrimSpace(email)),
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
	if email == s.bootstrapAdminEmail && s.bootstrapAdminEmail != "" {
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

func (s *Store) Ban(id int64, until time.Time, reason string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.byID[id]
	if !ok {
		return User{}, errors.New("user not found")
	}
	until = until.UTC()
	user.BannedUntil = &until
	user.BanReason = strings.TrimSpace(reason)
	s.byID[id] = user
	return user, nil
}

func (s *Store) IsBanned(user User, now time.Time) bool {
	return user.BannedUntil != nil && now.UTC().Before(*user.BannedUntil)
}

func (s *Store) Counts(now time.Time) (total, active int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, user := range s.byID {
		total++
		if !s.IsBanned(user, now) {
			active++
		}
	}
	return total, active
}
