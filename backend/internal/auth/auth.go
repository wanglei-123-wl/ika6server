package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wanglei-123-wl/ika6server/backend/internal/users"
)

const passwordIterations = 210000

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrAccountNotFound     = fmt.Errorf("account not found: %w", ErrInvalidCredentials)
	ErrPasswordMismatch    = fmt.Errorf("password verification failed: %w", ErrInvalidCredentials)
	ErrInvalidPasswordHash = errors.New("stored password hash is invalid")
)

type Service struct {
	users       users.Repository
	tokenSecret []byte
	revocations TokenRevocationStore
}

type TokenRevocationStore interface {
	RevokeTokenDigest(digest string, expiresAt time.Time) error
	IsTokenDigestRevoked(digest string) (bool, error)
}

func NewService(store users.Repository, tokenSecret string) *Service {
	return NewServiceWithRevocationStore(store, tokenSecret, nil)
}

func NewServiceWithRevocationStore(store users.Repository, tokenSecret string, revocations TokenRevocationStore) *Service {
	if revocations == nil {
		revocations = newMemoryRevocationStore()
	}
	return &Service{
		users:       store,
		tokenSecret: []byte(tokenSecret),
		revocations: revocations,
	}
}

func (s *Service) Register(username, email, password string) (users.User, string, error) {
	if len(password) < 8 {
		return users.User{}, "", errors.New("password must be at least 8 characters")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return users.User{}, "", err
	}

	user, err := s.users.Create(username, email, hash)
	if err != nil {
		return users.User{}, "", err
	}

	token := s.SignToken(user.ID)
	return user, token, nil
}

func (s *Service) Login(email, password string) (users.User, string, error) {
	return s.LoginWithRemember(email, password, false)
}

func (s *Service) LoginWithRemember(email, password string, remember bool) (users.User, string, error) {
	user, err := s.users.FindByEmail(email)
	if errors.Is(err, users.ErrNotFound) {
		return users.User{}, "", ErrAccountNotFound
	}
	if err != nil {
		return users.User{}, "", fmt.Errorf("lookup login account: %w", err)
	}
	matches, err := checkPassword(password, user.PasswordHash)
	if err != nil {
		return users.User{}, "", err
	}
	if !matches {
		return users.User{}, "", ErrPasswordMismatch
	}

	token := s.SignTokenWithTTL(user.ID, tokenTTL(remember))
	return user, token, nil
}

func (s *Service) SignToken(userID int64) string {
	return s.SignTokenWithTTL(userID, 24*time.Hour)
}

func (s *Service) SignTokenWithTTL(userID int64, ttl time.Duration) string {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	exp := time.Now().UTC().Add(ttl).Unix()
	tokenID, err := randomTokenID()
	if err != nil {
		tokenID = strconv.FormatInt(time.Now().UTC().UnixNano(), 36)
	}
	payload := fmt.Sprintf("%d.%d.%s", userID, exp, tokenID)
	mac := hmac.New(sha256.New, s.tokenSecret)
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload + "." + signature))
}

func tokenTTL(remember bool) time.Duration {
	if remember {
		return 30 * 24 * time.Hour
	}
	return 24 * time.Hour
}

func (s *Service) ParseToken(token string) (int64, error) {
	if s.IsRevoked(token) {
		return 0, errors.New("token revoked")
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return 0, errors.New("invalid token")
	}

	parts := strings.Split(string(raw), ".")
	if len(parts) != 3 && len(parts) != 4 {
		return 0, errors.New("invalid token")
	}

	payload := strings.Join(parts[:len(parts)-1], ".")
	mac := hmac.New(sha256.New, s.tokenSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[len(parts)-1])) != 1 {
		return 0, errors.New("invalid token")
	}

	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().UTC().Unix() > exp {
		return 0, errors.New("token expired")
	}

	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid token")
	}

	return userID, nil
}

func (s *Service) RevokeToken(token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	_, expiresAtUnix, err := s.parseTokenClaims(token)
	if err != nil {
		return
	}
	_ = s.revocations.RevokeTokenDigest(tokenDigest(token), time.Unix(expiresAtUnix, 0).UTC())
}

func (s *Service) IsRevoked(token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	revoked, err := s.revocations.IsTokenDigestRevoked(tokenDigest(token))
	return err != nil || revoked
}

func randomTokenID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func tokenDigest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Service) parseTokenClaims(token string) (int64, int64, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return 0, 0, errors.New("invalid token")
	}
	parts := strings.Split(string(raw), ".")
	if len(parts) != 3 && len(parts) != 4 {
		return 0, 0, errors.New("invalid token")
	}
	payload := strings.Join(parts[:len(parts)-1], ".")
	mac := hmac.New(sha256.New, s.tokenSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[len(parts)-1])) != 1 {
		return 0, 0, errors.New("invalid token")
	}
	exp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, errors.New("invalid token")
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, errors.New("invalid token")
	}
	return userID, exp, nil
}

type memoryRevocationStore struct {
	mu      sync.RWMutex
	revoked map[string]time.Time
}

func newMemoryRevocationStore() *memoryRevocationStore {
	return &memoryRevocationStore{revoked: make(map[string]time.Time)}
}

func (s *memoryRevocationStore) RevokeTokenDigest(digest string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[digest] = expiresAt.UTC()
	return nil
}

func (s *memoryRevocationStore) IsTokenDigestRevoked(digest string) (bool, error) {
	s.mu.RLock()
	expiresAt, ok := s.revoked[digest]
	s.mu.RUnlock()
	if !ok {
		return false, nil
	}
	if time.Now().UTC().After(expiresAt) {
		s.mu.Lock()
		delete(s.revoked, digest)
		s.mu.Unlock()
		return false, nil
	}
	return true, nil
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := stretch([]byte(password), salt, passwordIterations)
	return fmt.Sprintf("hmac-sha256-stretch$%d$%s$%s", passwordIterations, base64.RawURLEncoding.EncodeToString(salt), base64.RawURLEncoding.EncodeToString(hash)), nil
}

func CheckPassword(password, stored string) bool {
	matches, err := checkPassword(password, stored)
	return err == nil && matches
}

func checkPassword(password, stored string) (bool, error) {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "hmac-sha256-stretch" {
		return false, ErrInvalidPasswordHash
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false, ErrInvalidPasswordHash
	}

	salt, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(salt) == 0 {
		return false, ErrInvalidPasswordHash
	}

	expected, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil || len(expected) != sha256.Size {
		return false, ErrInvalidPasswordHash
	}

	actual := stretch([]byte(password), salt, iterations)
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}

func stretch(password, salt []byte, iterations int) []byte {
	mac := hmac.New(sha256.New, password)
	mac.Write(salt)
	sum := mac.Sum(nil)

	for i := 1; i < iterations; i++ {
		mac = hmac.New(sha256.New, password)
		mac.Write(sum)
		sum = mac.Sum(nil)
	}

	return sum
}
