package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformMigrationSeedsRequiredDefaults(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "migrations", "002_platform_schema.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(content)
	for _, fragment := range []string{"INSERT INTO forum_bars", "INSERT INTO developer_docs", "ON CONFLICT (id) DO NOTHING"} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("migration is missing %q", fragment)
		}
	}
}

func TestFinalizeMigrationExists(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "migrations", "003_finalize_platform_schema.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(content)
	for _, fragment := range []string{"CREATE TABLE IF NOT EXISTS game_files", "CREATE TABLE IF NOT EXISTS developer_docs", "ON CONFLICT (id) DO NOTHING"} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("final migration is missing %q", fragment)
		}
	}
}

func TestRevokedTokensMigrationExists(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "migrations", "004_revoked_tokens.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(content)
	for _, fragment := range []string{"CREATE TABLE IF NOT EXISTS revoked_tokens", "token_digest TEXT PRIMARY KEY", "expires_at TIMESTAMPTZ NOT NULL"} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("revoked tokens migration is missing %q", fragment)
		}
	}
}

func TestInteractionCountsMigrationExists(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "migrations", "005_interaction_counts.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(content)
	for _, fragment := range []string{"CREATE TABLE IF NOT EXISTS comment_likes", "PRIMARY KEY (comment_id, user_id)", "idx_comments_post_id"} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("interaction counts migration is missing %q", fragment)
		}
	}
}

func TestAdminAuditLogsMigrationExists(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "migrations", "006_admin_audit_logs.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(content)
	for _, fragment := range []string{"CREATE TABLE IF NOT EXISTS audit_logs", "details JSONB NOT NULL", "idx_audit_logs_target"} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("audit logs migration is missing %q", fragment)
		}
	}
}
