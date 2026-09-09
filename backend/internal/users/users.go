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
	ID               int64      `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	Role             Role       `json:"role"`
	PasswordHash     string     `json:"-"`
	CreatedAt        time.Time  `json:"createdAt"`
	BannedUntil      *time.Time `json:"bannedUntil,omitempty"`
	BanReason        string     `json:"banReason,omitempty"`
	Bio              string     `json:"bio,omitempty"`
	Location         string     `json:"location,omitempty"`
	AvatarURL        string     `json:"avatarUrl,omitempty"`
	AvatarStoredName string     `json:"-"`
	DeveloperEngines []string   `json:"engines,omitempty"`
}

type AdminUser struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Role        Role       `json:"role"`
	Level       string     `json:"level"`
	Status      string     `json:"status"`
	Games       int64      `json:"games"`
	Posts       int64      `json:"posts"`
	CreatedAt   time.Time  `json:"createdAt"`
	BannedUntil *time.Time `json:"bannedUntil"`
	BanReason   string     `json:"banReason"`
}

type Repository interface {
	Create(username, email, passwordHash string) (User, error)
	FindByEmail(email string) (User, bool)
	FindByUsername(username string) (User, bool)
	FindByID(id int64) (User, bool)
	Ban(id int64, until time.Time, reason string) (User, error)
	Unban(id int64) (User, error)
	UpdateProfile(id int64, bio, location string, engines []string) (User, error)
	UpdateAvatar(id int64, avatarURL, storedName string) (User, error)
	IsBanned(user User, now time.Time) bool
	Counts(now time.Time) (total, active int)
	ListAdminUsers(page, pageSize int, now time.Time) ([]AdminUser, int, error)
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

func (s *Store) FindByUsername(username string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	username = strings.TrimSpace(username)
	for _, user := range s.byID {
		if user.Username == username {
			return user, true
		}
	}
	return User{}, false
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

func (s *Store) Unban(id int64) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.byID[id]
	if !ok {
		return User{}, errors.New("user not found")
	}
	user.BannedUntil = nil
	user.BanReason = ""
	s.byID[id] = user
	return user, nil
}

func (s *Store) UpdateProfile(id int64, bio, location string, engines []string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.byID[id]
	if !ok {
		return User{}, errors.New("user not found")
	}
	user.Bio = strings.TrimSpace(bio)
	user.Location = strings.TrimSpace(location)
	user.DeveloperEngines = cleanEngines(engines)
	s.byID[id] = user
	return user, nil
}

func (s *Store) UpdateAvatar(id int64, avatarURL, storedName string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.byID[id]
	if !ok {
		return User{}, errors.New("user not found")
	}
	user.AvatarURL = strings.TrimSpace(avatarURL)
	user.AvatarStoredName = strings.TrimSpace(storedName)
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

func (s *Store) ListAdminUsers(page, pageSize int, now time.Time) ([]AdminUser, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	users := make([]User, 0, len(s.byID))
	for _, user := range s.byID {
		users = append(users, user)
	}
	start := (page - 1) * pageSize
	if start > len(users) {
		start = len(users)
	}
	end := start + pageSize
	if end > len(users) {
		end = len(users)
	}
	items := make([]AdminUser, 0, end-start)
	for _, user := range users[start:end] {
		items = append(items, adminUserFromUser(user, s.IsBanned(user, now), 0, 0))
	}
	return items, len(users), nil
}

func adminUserFromUser(user User, banned bool, games, posts int64) AdminUser {
	status := "normal"
	if banned {
		status = "banned"
	}
	return AdminUser{
		ID:          user.ID,
		Name:        user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Level:       "lv1",
		Status:      status,
		Games:       games,
		Posts:       posts,
		CreatedAt:   user.CreatedAt.UTC(),
		BannedUntil: user.BannedUntil,
		BanReason:   user.BanReason,
	}
}

func cleanEngines(engines []string) []string {
	result := make([]string, 0, len(engines))
	seen := make(map[string]struct{}, len(engines))
	for _, engine := range engines {
		engine = strings.TrimSpace(engine)
		if engine == "" {
			continue
		}
		key := strings.ToLower(engine)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, engine)
	}
	return result
}
