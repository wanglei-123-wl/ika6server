package auth

import (
	"database/sql"
	"errors"
	"time"
)

type SQLRevocationStore struct {
	db *sql.DB
}

var _ TokenRevocationStore = (*SQLRevocationStore)(nil)

func NewSQLRevocationStore(db *sql.DB) (*SQLRevocationStore, error) {
	if db == nil {
		return nil, errors.New("sql database is required")
	}
	return &SQLRevocationStore{db: db}, nil
}

func (s *SQLRevocationStore) RevokeTokenDigest(digest string, expiresAt time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO revoked_tokens (token_digest, expires_at)
		VALUES ($1, $2)
		ON CONFLICT (token_digest) DO UPDATE SET
			expires_at = EXCLUDED.expires_at,
			revoked_at = now()`,
		digest, expiresAt.UTC())
	return err
}

func (s *SQLRevocationStore) IsTokenDigestRevoked(digest string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM revoked_tokens
			WHERE token_digest = $1 AND expires_at > now()
			)`, digest).Scan(&exists)
	return exists, err
}
