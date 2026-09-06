package users

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

type SQLRepository struct {
	db                  *sql.DB
	bootstrapAdminEmail string
}

var _ Repository = (*SQLRepository)(nil)

type nullableTimeValue struct {
	target **time.Time
}

func (v nullableTimeValue) Scan(src any) error {
	var value sql.NullTime
	if err := value.Scan(src); err != nil {
		return err
	}
	if value.Valid {
		utc := value.Time.UTC()
		*v.target = &utc
	} else {
		*v.target = nil
	}
	return nil
}

func NewSQLRepository(db *sql.DB, bootstrapAdminEmail ...string) (*SQLRepository, error) {
	if db == nil {
		return nil, errors.New("sql database is required")
	}
	adminEmail := ""
	if len(bootstrapAdminEmail) > 0 {
		adminEmail = strings.ToLower(strings.TrimSpace(bootstrapAdminEmail[0]))
	}
	return &SQLRepository{db: db, bootstrapAdminEmail: adminEmail}, nil
}

func (r *SQLRepository) Create(username, email, passwordHash string) (User, error) {
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))
	if username == "" || email == "" || passwordHash == "" {
		return User{}, errors.New("username, email and password are required")
	}
	var user User
	var role string
	err := r.db.QueryRow(`
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, CASE WHEN $4 <> '' AND $2 = $4 THEN 'admin' ELSE 'user' END)
		RETURNING id, username, email, role, password_hash, created_at`,
		username, email, passwordHash, r.bootstrapAdminEmail,
	).Scan(&user.ID, &user.Username, &user.Email, &role, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}
	user.Role = Role(role)
	return user, nil
}

func (r *SQLRepository) FindByEmail(email string) (User, bool) {
	var user User
	var role string
	err := r.db.QueryRow(`
		SELECT id, username, email, role, password_hash, created_at, banned_until, ban_reason
		FROM users WHERE email = $1`, strings.ToLower(strings.TrimSpace(email))).
		Scan(&user.ID, &user.Username, &user.Email, &role, &user.PasswordHash, &user.CreatedAt, nullableTime(&user.BannedUntil), &user.BanReason)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, false
	}
	if err != nil {
		return User{}, false
	}
	user.Role = Role(role)
	return user, true
}

func (r *SQLRepository) FindByID(id int64) (User, bool) {
	var user User
	var role string
	err := r.db.QueryRow(`
		SELECT id, username, email, role, password_hash, created_at, banned_until, ban_reason
		FROM users WHERE id = $1`, id).
		Scan(&user.ID, &user.Username, &user.Email, &role, &user.PasswordHash, &user.CreatedAt, nullableTime(&user.BannedUntil), &user.BanReason)
	if errors.Is(err, sql.ErrNoRows) || err != nil {
		return User{}, false
	}
	user.Role = Role(role)
	return user, true
}

func (r *SQLRepository) Ban(id int64, until time.Time, reason string) (User, error) {
	var user User
	var role string
	err := r.db.QueryRow(`
		UPDATE users
		SET banned_until = $1, ban_reason = $2
		WHERE id = $3
		RETURNING id, username, email, role, password_hash, created_at, banned_until, ban_reason`,
		until.UTC(), strings.TrimSpace(reason), id).
		Scan(&user.ID, &user.Username, &user.Email, &role, &user.PasswordHash, &user.CreatedAt, nullableTime(&user.BannedUntil), &user.BanReason)
	if err != nil {
		return User{}, err
	}
	user.Role = Role(role)
	return user, nil
}

func (r *SQLRepository) IsBanned(user User, now time.Time) bool {
	return user.BannedUntil != nil && now.UTC().Before(*user.BannedUntil)
}

func (r *SQLRepository) Counts(now time.Time) (total, active int) {
	err := r.db.QueryRow(`
		SELECT COUNT(*), COUNT(*) FILTER (
			WHERE banned_until IS NULL OR banned_until <= $1
		) FROM users`, now.UTC()).Scan(&total, &active)
	if err != nil {
		return 0, 0
	}
	return total, active
}

func nullableTime(target **time.Time) nullableTimeValue {
	return nullableTimeValue{target: target}
}
