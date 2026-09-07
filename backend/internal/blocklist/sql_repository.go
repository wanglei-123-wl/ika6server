package blocklist

import (
	"database/sql"
	"errors"
)

type SQLRepository struct {
	db *sql.DB
}

var _ Repository = (*SQLRepository)(nil)

func NewSQLRepository(db *sql.DB) (*SQLRepository, error) {
	if db == nil {
		return nil, errors.New("sql database is required")
	}
	return &SQLRepository{db: db}, nil
}

func (r *SQLRepository) Add(sha256, reason string, createdBy int64) (Entry, error) {
	sha256 = normalizeSHA256(sha256)
	if sha256 == "" || reason == "" {
		return Entry{}, errors.New("sha256 and reason are required")
	}
	var entry Entry
	err := r.db.QueryRow(`
		INSERT INTO download_blocklist (sha256, reason, created_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (sha256) DO UPDATE SET
			reason = EXCLUDED.reason,
			created_by = EXCLUDED.created_by,
			created_at = now()
		RETURNING sha256, reason, created_by, created_at`,
		sha256, reason, createdBy).
		Scan(&entry.SHA256, &entry.Reason, &entry.CreatedBy, &entry.CreatedAt)
	return entry, err
}

func (r *SQLRepository) Contains(sha256 string) (Entry, bool) {
	var entry Entry
	err := r.db.QueryRow(`
		SELECT sha256, reason, created_by, created_at
		FROM download_blocklist
		WHERE sha256 = $1`, normalizeSHA256(sha256)).
		Scan(&entry.SHA256, &entry.Reason, &entry.CreatedBy, &entry.CreatedAt)
	return entry, err == nil
}

func (r *SQLRepository) List() []Entry {
	rows, err := r.db.Query(`
		SELECT sha256, reason, created_by, created_at
		FROM download_blocklist
		ORDER BY created_at DESC`)
	if err != nil {
		return []Entry{}
	}
	defer rows.Close()
	items := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.SHA256, &entry.Reason, &entry.CreatedBy, &entry.CreatedAt); err != nil {
			return []Entry{}
		}
		items = append(items, entry)
	}
	return items
}
