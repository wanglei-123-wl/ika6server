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
	"time"

	"github.com/wanglei-123-wl/ika6server/backend/internal/users"
)

const passwordIterations = 210000

type Service struct {
	users       users.Repository
	tokenSecret []byte
}

func NewService(store users.Repository, tokenSecret string) *Service {
	return &Service{
		users:       store,
		tokenSecret: []byte(tokenSecret),
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
	user, ok := s.users.FindByEmail(email)
	if !ok || !CheckPassword(password, user.PasswordHash) {
		return users.User{}, "", errors.New("invalid email or password")
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
	payload := fmt.Sprintf("%d.%d", userID, exp)
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
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(token))
	if err != nil {
		return 0, errors.New("invalid token")
	}

	parts := strings.Split(string(raw), ".")
	if len(parts) != 3 {
		return 0, errors.New("invalid token")
	}

	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, s.tokenSecret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[2])) != 1 {
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

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := stretch([]byte(password), salt, passwordIterations)
	return fmt.Sprintf("hmac-sha256-stretch$%d$%s$%s", passwordIterations, base64.RawURLEncoding.EncodeToString(salt), base64.RawURLEncoding.EncodeToString(hash)), nil
}

func CheckPassword(password, stored string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "hmac-sha256-stretch" {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}

	salt, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	expected, err := base64.RawURLEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	actual := stretch([]byte(password), salt, iterations)
	return subtle.ConstantTimeCompare(actual, expected) == 1
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
