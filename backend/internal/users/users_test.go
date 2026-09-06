package users

import (
	"testing"
	"time"
)

func TestDefaultRegistrationDoesNotGrantAdmin(t *testing.T) {
	store := NewStore()
	user, err := store.Create("member", "member@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != RoleUser {
		t.Fatalf("role = %q, want %q", user.Role, RoleUser)
	}
}

func TestBootstrapAccountGetsAdminRole(t *testing.T) {
	store := NewStoreWithAdmin("admin@test.com")
	user, err := store.Create("admin", "admin@test.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != RoleAdmin {
		t.Fatalf("role = %q, want %q", user.Role, RoleAdmin)
	}
	if store.IsBanned(user, time.Now()) {
		t.Fatal("new user should not be banned")
	}
}
