package users

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func TestNewSQLRepositoryRequiresConnection(t *testing.T) {
	if _, err := NewSQLRepository(nil); err == nil {
		t.Fatal("expected nil sql database to fail")
	}
}

func TestPostgresAdminBootstrapAndLoginLookup(t *testing.T) {
	dsn := os.Getenv("IKA6_AUTH_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set IKA6_AUTH_TEST_DATABASE_URL to run isolated PostgreSQL auth tests")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid auth test database configuration")
	}
	// Every connection can resolve only temporary tables, never public.users.
	cfg.RuntimeParams["search_path"] = "pg_temp"
	cfg.RuntimeParams["statement_timeout"] = "5000"
	cfg.ConnectTimeout = 5 * time.Second
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal("auth test database is not reachable")
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TEMP TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			banned_until TIMESTAMPTZ,
			ban_reason TEXT DEFAULT '',
			bio TEXT NOT NULL DEFAULT '',
			location TEXT NOT NULL DEFAULT '',
			avatar_url TEXT NOT NULL DEFAULT '',
			avatar_stored_name TEXT NOT NULL DEFAULT '',
			developer_engines JSONB NOT NULL DEFAULT '[]'
		)`); err != nil {
		t.Fatal(err)
	}
	repository, err := NewSQLRepository(db, "admin@test.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.EnsureUniqueAdmin(" ADMIN@TEST.COM ", "original-hash"); err != nil {
		t.Fatal(err)
	}
	first, err := repository.FindByEmail(" ADMIN@TEST.COM ")
	if err != nil || first.Role != RoleAdmin || first.PasswordHash != "original-hash" {
		t.Fatalf("initial admin was not created correctly: %v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE pg_temp.users SET password_hash = 'updated-hash' WHERE id = $1`, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := repository.EnsureUniqueAdmin("admin@test.com", "stale-startup-hash"); err != nil {
		t.Fatal(err)
	}
	after, err := repository.FindByEmail("admin@test.com")
	if err != nil || after.ID != first.ID || after.PasswordHash != "updated-hash" || after.Role != RoleAdmin {
		t.Fatalf("startup must preserve the existing administrator password: %v", err)
	}
	if _, err := repository.FindByEmail("missing@test.com"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing user should return ErrNotFound: %v", err)
	}

	// Simulate a client-side scan failure independently of a database query error.
	if _, err := db.ExecContext(ctx, `UPDATE pg_temp.users SET ban_reason = NULL WHERE id = $1`, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindByEmail("admin@test.com"); err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("scan errors must not look like a missing account: %v", err)
	}
	if _, err := db.ExecContext(ctx, `DROP TABLE pg_temp.users`); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindByEmail("admin@test.com"); err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("query errors must not look like a missing account: %v", err)
	}
}
