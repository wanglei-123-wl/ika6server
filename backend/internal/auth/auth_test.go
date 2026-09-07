package auth

import (
	"strings"
	"testing"

	"github.com/wanglei-123-wl/ika6server/backend/internal/users"
)

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
