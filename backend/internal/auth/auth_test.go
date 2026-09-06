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
