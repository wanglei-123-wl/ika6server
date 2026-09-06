package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrDriverUnavailable = errors.New("postgres driver is not installed")

type Database struct {
	DSN   string
	sqlDB *sql.DB
}

func Open(dsn string) (*Database, error) {
	if strings.TrimSpace(dsn) == "" {
		return &Database{}, nil
	}
	if err := ValidateDSN(dsn); err != nil {
		return nil, err
	}
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres database: %w", err)
	}
	return &Database{DSN: dsn, sqlDB: sqlDB}, nil
}

func (db *Database) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if db == nil || db.DSN == "" {
		return nil
	}
	if db.sqlDB != nil {
		return db.sqlDB.PingContext(ctx)
	}
	return ErrDriverUnavailable
}

func (db *Database) SQL() *sql.DB {
	if db == nil {
		return nil
	}
	return db.sqlDB
}

func OpenSQL(db *sql.DB, dsn string) (*Database, error) {
	if db == nil {
		return nil, errors.New("sql database is required")
	}
	if err := ValidateDSN(dsn); err != nil {
		return nil, err
	}
	return &Database{DSN: dsn, sqlDB: db}, nil
}

func (db *Database) Migrate(ctx context.Context, dir string) error {
	if db == nil || db.sqlDB == nil {
		return ErrDriverUnavailable
	}
	migrations, err := LoadMigrations(dir)
	if err != nil {
		return err
	}
	return ApplyMigrations(ctx, db.sqlDB, migrations)
}

func (db *Database) Close() error {
	if db == nil || db.sqlDB == nil {
		return nil
	}
	return db.sqlDB.Close()
}

func ValidateDSN(dsn string) error {
	parsed, err := url.Parse(strings.TrimSpace(dsn))
	if err != nil || (parsed.Scheme != "postgres" && parsed.Scheme != "postgresql") {
		return errors.New("database URL must use postgres:// or postgresql://")
	}
	if parsed.Host == "" {
		return errors.New("database URL must include a host")
	}
	return nil
}

type Migration struct {
	Name string
	SQL  string
}

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type migrationQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func ApplyMigrations(ctx context.Context, executor SQLExecutor, migrations []Migration) error {
	if executor == nil {
		return errors.New("migration executor is required")
	}
	if _, err := executor.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	applied := make(map[string]struct{})
	if queryer, ok := executor.(migrationQueryer); ok {
		rows, err := queryer.QueryContext(ctx, `SELECT name FROM schema_migrations`)
		if err != nil {
			return fmt.Errorf("load applied migrations: %w", err)
		}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				rows.Close()
				return fmt.Errorf("scan applied migration: %w", err)
			}
			applied[name] = struct{}{}
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("close applied migrations: %w", err)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("read applied migrations: %w", err)
		}
	}
	for _, migration := range migrations {
		if strings.TrimSpace(migration.Name) == "" || strings.TrimSpace(migration.SQL) == "" {
			return errors.New("migration name and SQL are required")
		}
		if _, ok := applied[migration.Name]; ok {
			continue
		}
		if _, err := executor.ExecContext(ctx, migration.SQL); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.Name, err)
		}
		if _, err := executor.ExecContext(ctx, `
			INSERT INTO schema_migrations (name) VALUES ($1)
			ON CONFLICT (name) DO NOTHING`, migration.Name); err != nil {
			return fmt.Errorf("record migration %s: %w", migration.Name, err)
		}
		applied[migration.Name] = struct{}{}
	}
	return nil
}

func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		sql, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, Migration{Name: entry.Name(), SQL: string(sql)})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Name < migrations[j].Name })
	return migrations, nil
}
