package auth

import (
	"errors"
	"strings"
	"testing"

	"github.com/wanglei-123-wl/ika6server/backend/internal/users"
)

type failingUserLookup struct {
	users.Repository
	err error
}

func (r failingUserLookup) FindByEmail(string) (users.User, error) {
	return users.User{}, r.err
}

func TestLoginPreservesLookupErrors(t *testing.T) {
	lookupErr := errors.New("database unavailable")
	service := NewService(failingUserLookup{err: lookupErr}, "test-secret")
	user, token, err := service.Login("admin@test.com", "test-password")
	if !errors.Is(err, lookupErr) || errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unexpected error classification: %v", err)
	}
	if user.ID != 0 || token != "" {
		t.Fatal("failed lookup must not issue a user or token")
	}
}

func TestLoginClassifiesCredentialFailures(t *testing.T) {
	store := users.NewStore()
	service := NewService(store, "test-secret")
	if _, _, err := service.Register("member", "member@test.com", "test-password"); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		account string
		want    error
	}{
		{"missing@test.com", ErrAccountNotFound},
		{"member@test.com", ErrPasswordMismatch},
	} {
		_, token, err := service.Login(tc.account, "wrong-password")
		if !errors.Is(err, tc.want) || !errors.Is(err, ErrInvalidCredentials) || token != "" {
			t.Fatalf("unexpected credential failure for %s: %v", tc.account, err)
		}
	}
}

func TestLoginRejectsMalformedStoredHashesAsServiceErrors(t *testing.T) {
	for _, hash := range []string{
		"broken-hash",
		"hmac-sha256-stretch$0$c2FsdA$ZGlnaWVzdA",
		"hmac-sha256-stretch$1$!$ZGlnaWVzdA",
		"hmac-sha256-stretch$1$$ZGlnaWVzdA",
		"hmac-sha256-stretch$1$c2FsdA$!",
		"hmac-sha256-stretch$1$c2FsdA$c2hvcnQ",
	} {
		store := users.NewStore()
		if _, err := store.Create("member", "member@test.com", hash); err != nil {
			t.Fatal(err)
		}
		service := NewService(store, "test-secret")
		_, token, err := service.Login("member@test.com", "test-password")
		if !errors.Is(err, ErrInvalidPasswordHash) || errors.Is(err, ErrInvalidCredentials) || token != "" {
			t.Fatalf("malformed stored hash was not classified correctly: %v", err)
		}
		if CheckPassword("test-password", hash) {
			t.Fatal("CheckPassword must reject a malformed stored hash")
		}
	}
}

func TestRememberLoginCreatesParsableToken(t *testing.T) {
	store := users.NewStore()
	service := NewService(store, "test-secret")
	if _, _, err := service.Register("member", "member@test.com", "12345678"); err != nil {
		t.Fatal(err)
	}
	_, regularToken, err := service.LoginWithRemember("member@test.com", "12345678", false)
	if err != nil {
		t.Fatal(err)
	}
	_, rememberedToken, err := service.LoginWithRemember("member@test.com", "12345678", true)
	if err != nil {
		t.Fatal(err)
	}
	if regularToken == rememberedToken || strings.TrimSpace(regularToken) == "" {
		t.Fatal("expected distinct non-empty tokens")
	}
	if _, err := service.ParseToken(rememberedToken); err != nil {
		t.Fatal(err)
	}
}

func TestTokenRevocationAndFreshLoginToken(t *testing.T) {
	store := users.NewStore()
	service := NewService(store, "test-secret")
	if _, _, err := service.Register("member", "member@test.com", "12345678"); err != nil {
		t.Fatal(err)
	}
	_, firstToken, err := service.Login("member@test.com", "12345678")
	if err != nil {
		t.Fatal(err)
	}
	service.RevokeToken(firstToken)
	if _, err := service.ParseToken(firstToken); err == nil {
		t.Fatal("expected revoked token to fail")
	}
	_, secondToken, err := service.Login("member@test.com", "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if secondToken == firstToken {
		t.Fatal("expected login to issue a fresh token")
	}
	if _, err := service.ParseToken(secondToken); err != nil {
		t.Fatal(err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	store := users.NewStore()
	service := NewService(store, "test-secret")
	if _, _, err := service.Register("member", "member@test.com", "12345678"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := service.Login("member@test.com", "wrong-password"); err == nil {
		t.Fatal("expected login with wrong password to fail")
	}
}

func TestLoginRejectsWrongAccountWithKnownPassword(t *testing.T) {
	store := users.NewStore()
	service := NewService(store, "test-secret")
	if _, _, err := service.Register("member", "member@test.com", "12345678"); err != nil {
		t.Fatal(err)
	}

	if _, _, err := service.Login("unknown@test.com", "12345678"); err == nil {
		t.Fatal("expected login with wrong account to fail")
	}
}
