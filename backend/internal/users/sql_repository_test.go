package users

import "testing"

func TestNewSQLRepositoryRequiresConnection(t *testing.T) {
	if _, err := NewSQLRepository(nil); err == nil {
		t.Fatal("expected nil sql database to fail")
	}
}
