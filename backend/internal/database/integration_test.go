package database

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("IKA6_DATABASE_URL")
	if dsn == "" {
		t.Skip("IKA6_DATABASE_URL is not configured")
	}
	db, err := Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	migrations, err := LoadMigrations("../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyMigrations(ctx, db.SQL(), migrations); err != nil {
		t.Fatal(err)
	}
}
