package database

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type recordingExecutor struct {
	queries []string
}

func (r *recordingExecutor) ExecContext(_ context.Context, query string, _ ...any) (sql.Result, error) {
	r.queries = append(r.queries, query)
	return nil, nil
}

func TestValidateDSN(t *testing.T) {
	if err := ValidateDSN("postgres://user:pass@localhost:5432/ika6"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateDSN("sqlite://local"); err == nil {
		t.Fatal("expected unsupported scheme to fail")
	}
}

func TestOpenWithoutDSNUsesMemoryMode(t *testing.T) {
	db, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestOpenWithDSNUsesPostgresDriver(t *testing.T) {
	db, err := Open("postgres://user:pass@localhost:5432/ika6")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if db.SQL() == nil {
		t.Fatal("expected sql database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := db.Ping(ctx); err == nil {
		t.Fatal("expected unavailable test database to fail ping")
	}
}

func TestOpenSQLRequiresConnection(t *testing.T) {
	if _, err := OpenSQL(nil, "postgres://user:pass@localhost:5432/ika6"); err == nil {
		t.Fatal("expected nil sql database to fail")
	}
}

func TestLoadMigrations(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "002_second.sql"), []byte("SELECT 2"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "001_first.sql"), []byte("SELECT 1"), 0600); err != nil {
		t.Fatal(err)
	}
	items, err := LoadMigrations(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Name != "001_first.sql" || items[1].Name != "002_second.sql" {
		t.Fatalf("unexpected migrations: %#v", items)
	}
}

func TestApplyMigrationsIsIdempotentAtDatabaseLevel(t *testing.T) {
	executor := &recordingExecutor{}
	err := ApplyMigrations(context.Background(), executor, []Migration{{Name: "001_initial_schema.sql", SQL: "CREATE TABLE users (id BIGINT)"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(executor.queries) != 3 {
		t.Fatalf("executed %d queries, want 3", len(executor.queries))
	}
	if err := ApplyMigrations(context.Background(), nil, nil); err == nil {
		t.Fatal("expected nil executor to fail")
	}
}
