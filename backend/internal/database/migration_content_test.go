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
