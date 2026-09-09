package users

import (
	"database/sql"
	"encoding/json"
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

func (r *SQLRepository) EnsureUniqueAdmin(account, passwordHash string) error {
	account = strings.ToLower(strings.TrimSpace(account))
	passwordHash = strings.TrimSpace(passwordHash)
	if account == "" || passwordHash == "" {
		return errors.New("admin account and password hash are required")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`
		UPDATE users
		SET role = 'user'
		WHERE LOWER(email) <> $1 AND role = 'admin'`, account); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, 'admin')
		ON CONFLICT (email) DO UPDATE SET
			role = 'admin'`,
		adminUsername(account), account, passwordHash); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLRepository) FindByEmail(email string) (User, error) {
	user, err := scanUser(r.db.QueryRow(`
		SELECT id, username, email, role, password_hash, created_at, banned_until, ban_reason,
		       bio, location, avatar_url, avatar_stored_name, developer_engines
		FROM users WHERE email = $1`, strings.ToLower(strings.TrimSpace(email))))
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) FindByUsername(username string) (User, bool) {
	user, err := scanUser(r.db.QueryRow(`
		SELECT id, username, email, role, password_hash, created_at, banned_until, ban_reason,
		       bio, location, avatar_url, avatar_stored_name, developer_engines
		FROM users WHERE username = $1`, strings.TrimSpace(username)))
	if errors.Is(err, sql.ErrNoRows) || err != nil {
		return User{}, false
	}
	return user, true
}

func (r *SQLRepository) FindByID(id int64) (User, bool) {
	user, err := scanUser(r.db.QueryRow(`
		SELECT id, username, email, role, password_hash, created_at, banned_until, ban_reason,
		       bio, location, avatar_url, avatar_stored_name, developer_engines
		FROM users WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) || err != nil {
		return User{}, false
	}
	return user, true
}

func (r *SQLRepository) Ban(id int64, until time.Time, reason string) (User, error) {
	user, err := scanUser(r.db.QueryRow(`
		UPDATE users
		SET banned_until = $1, ban_reason = $2
		WHERE id = $3
		RETURNING id, username, email, role, password_hash, created_at, banned_until, ban_reason,
		          bio, location, avatar_url, avatar_stored_name, developer_engines`,
		until.UTC(), strings.TrimSpace(reason), id))
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) Unban(id int64) (User, error) {
	user, err := scanUser(r.db.QueryRow(`
		UPDATE users
		SET banned_until = NULL, ban_reason = ''
		WHERE id = $1
		RETURNING id, username, email, role, password_hash, created_at, banned_until, ban_reason,
		          bio, location, avatar_url, avatar_stored_name, developer_engines`, id))
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (r *SQLRepository) UpdateProfile(id int64, bio, location string, engines []string) (User, error) {
	enginesJSON, err := json.Marshal(cleanEngines(engines))
	if err != nil {
		return User{}, err
	}
	return scanUser(r.db.QueryRow(`
		UPDATE users
		SET bio = $1, location = $2, developer_engines = $3
		WHERE id = $4
		RETURNING id, username, email, role, password_hash, created_at, banned_until, ban_reason,
		          bio, location, avatar_url, avatar_stored_name, developer_engines`,
		strings.TrimSpace(bio), strings.TrimSpace(location), enginesJSON, id))
}

func (r *SQLRepository) UpdateAvatar(id int64, avatarURL, storedName string) (User, error) {
	return scanUser(r.db.QueryRow(`
		UPDATE users
		SET avatar_url = $1, avatar_stored_name = $2
		WHERE id = $3
		RETURNING id, username, email, role, password_hash, created_at, banned_until, ban_reason,
		          bio, location, avatar_url, avatar_stored_name, developer_engines`,
		strings.TrimSpace(avatarURL), strings.TrimSpace(storedName), id))
}

func (r *SQLRepository) IsBanned(user User, now time.Time) bool {
	return user.BannedUntil != nil && now.UTC().Before(*user.BannedUntil)
}

func (r *SQLRepository) ListAdminUsers(page, pageSize int, now time.Time) ([]AdminUser, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
		SELECT u.id, u.username, u.email, u.role, u.created_at, u.banned_until, u.ban_reason,
		       COUNT(DISTINCT g.id), COUNT(DISTINCT p.id)
		FROM users u
		LEFT JOIN games g ON g.owner_id = u.id
		LEFT JOIN forum_posts p ON p.author_id = u.id
		GROUP BY u.id
		ORDER BY u.created_at DESC, u.id DESC
		LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]AdminUser, 0)
	for rows.Next() {
		var user User
		var role string
		var games, posts int64
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &role, &user.CreatedAt, nullableTime(&user.BannedUntil), &user.BanReason, &games, &posts); err != nil {
			return nil, 0, err
		}
		user.Role = Role(role)
		items = append(items, adminUserFromUser(user, r.IsBanned(user, now), games, posts))
	}
	return items, total, rows.Err()
}

func adminUsername(account string) string {
	name := strings.Split(strings.TrimSpace(account), "@")[0]
	if name == "" {
		return "admin"
	}
	return name
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

type userRowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row userRowScanner) (User, error) {
	var user User
	var role string
	var enginesJSON []byte
	err := row.Scan(&user.ID, &user.Username, &user.Email, &role, &user.PasswordHash, &user.CreatedAt,
		nullableTime(&user.BannedUntil), &user.BanReason, &user.Bio, &user.Location,
		&user.AvatarURL, &user.AvatarStoredName, &enginesJSON)
	if err != nil {
		return User{}, err
	}
	user.Role = Role(role)
	user.DeveloperEngines = decodeEngines(enginesJSON)
	return user, nil
}

func decodeEngines(data []byte) []string {
	var engines []string
	if len(data) > 0 && json.Unmarshal(data, &engines) == nil {
		return cleanEngines(engines)
	}
	return []string{}
}
