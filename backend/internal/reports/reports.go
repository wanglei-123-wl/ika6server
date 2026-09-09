package reports

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrNotFound      = errors.New("report not found")
	ErrConflict      = errors.New("report is already processed")
	ErrInvalidStatus = errors.New("invalid report status")
	ErrInvalidTarget = errors.New("report target is required")
	ErrInvalidReason = errors.New("report reason is required")
)

type Report struct {
	ID         int64      `json:"id"`
	ReporterID int64      `json:"reporterId"`
	Reporter   string     `json:"reporter"`
	TargetType string     `json:"type"`
	TargetID   int64      `json:"targetId"`
	Target     string     `json:"target"`
	Reason     string     `json:"reason"`
	Details    string     `json:"details,omitempty"`
	Status     string     `json:"status"`
	Resolution string     `json:"resolution,omitempty"`
	ResolvedBy *int64     `json:"resolvedBy,omitempty"`
	ResolvedAt *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type Repository interface {
	Create(ctx context.Context, reporterID int64, targetType string, targetID int64, reason, details string) (Report, error)
	List(ctx context.Context, page, pageSize int, status string) ([]Report, int, error)
	Resolve(ctx context.Context, id, resolverID int64, status, resolution string) (Report, error)
	PendingCount(ctx context.Context) (int, error)
}

type Store struct {
	mu     sync.RWMutex
	nextID int64
	items  []Report
}

var _ Repository = (*Store)(nil)

func NewStore() *Store {
	return &Store{nextID: 1, items: make([]Report, 0)}
}

func (s *Store) Create(_ context.Context, reporterID int64, targetType string, targetID int64, reason, details string) (Report, error) {
	if reporterID <= 0 || strings.TrimSpace(targetType) == "" || targetID <= 0 {
		return Report{}, ErrInvalidTarget
	}
	if strings.TrimSpace(reason) == "" {
		return Report{}, ErrInvalidReason
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	item := Report{ID: s.nextID, ReporterID: reporterID, TargetType: strings.TrimSpace(targetType), TargetID: targetID, Reason: strings.TrimSpace(reason), Details: strings.TrimSpace(details), Status: "pending", CreatedAt: time.Now().UTC()}
	s.nextID++
	s.items = append(s.items, item)
	return item, nil
}

func (s *Store) List(_ context.Context, page, pageSize int, status string) ([]Report, int, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "pending"
	}
	if err := validateStatus(status); err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	filtered := make([]Report, 0, len(s.items))
	for _, item := range s.items {
		if status != "all" && item.Status != status {
			continue
		}
		filtered = append(filtered, item)
	}
	for left, right := 0, len(filtered)-1; left < right; left, right = left+1, right-1 {
		filtered[left], filtered[right] = filtered[right], filtered[left]
	}
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	items := make([]Report, end-start)
	copy(items, filtered[start:end])
	return items, len(filtered), nil
}

func (s *Store) Resolve(_ context.Context, id, resolverID int64, status, resolution string) (Report, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "resolved"
	}
	if err := validateStatus(status); err != nil || status == "all" || status == "pending" {
		return Report{}, ErrInvalidStatus
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.items {
		if s.items[index].ID != id {
			continue
		}
		if s.items[index].Status != "pending" {
			return Report{}, ErrConflict
		}
		now := time.Now().UTC()
		s.items[index].Status = status
		s.items[index].Resolution = strings.TrimSpace(resolution)
		s.items[index].ResolvedBy = &resolverID
		s.items[index].ResolvedAt = &now
		return s.items[index], nil
	}
	return Report{}, ErrNotFound
}

func (s *Store) PendingCount(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, item := range s.items {
		if item.Status == "pending" {
			count++
		}
	}
	return count, nil
}

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

func (r *SQLRepository) Create(ctx context.Context, reporterID int64, targetType string, targetID int64, reason, details string) (Report, error) {
	if reporterID <= 0 || strings.TrimSpace(targetType) == "" || targetID <= 0 {
		return Report{}, ErrInvalidTarget
	}
	if strings.TrimSpace(reason) == "" {
		return Report{}, ErrInvalidReason
	}
	return scanReport(r.db.QueryRowContext(ctx, `
		WITH inserted AS (
			INSERT INTO reports (reporter_id, target_type, target_id, reason, details)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, reporter_id, target_type, target_id, reason, details, status, resolution, resolved_by, resolved_at, created_at
		)
		SELECT i.id, i.reporter_id, u.username, i.target_type, i.target_id, i.reason, i.details,
		       i.status, i.resolution, i.resolved_by, i.resolved_at, i.created_at
		FROM inserted i
		JOIN users u ON u.id = i.reporter_id`,
		reporterID, strings.TrimSpace(targetType), targetID, strings.TrimSpace(reason), strings.TrimSpace(details)))
}

func (r *SQLRepository) List(ctx context.Context, page, pageSize int, status string) ([]Report, int, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "pending"
	}
	if err := validateStatus(status); err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	args := []any{}
	filter := ""
	if status != "all" {
		filter = "WHERE r.status = $1"
		args = append(args, status)
	}
	var total int
	countQuery := `SELECT COUNT(*) FROM reports r ` + filter
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.reporter_id, u.username, r.target_type, r.target_id, r.reason, r.details,
		       r.status, r.resolution, r.resolved_by, r.resolved_at, r.created_at
		FROM reports r
		JOIN users u ON u.id = r.reporter_id
		`+filter+`
		ORDER BY r.created_at DESC, r.id DESC
		LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args))+``, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]Report, 0)
	for rows.Next() {
		item, err := scanReport(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *SQLRepository) Resolve(ctx context.Context, id, resolverID int64, status, resolution string) (Report, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "resolved"
	}
	if err := validateStatus(status); err != nil || status == "all" || status == "pending" {
		return Report{}, ErrInvalidStatus
	}
	item, err := scanReport(r.db.QueryRowContext(ctx, `
		WITH updated AS (
			UPDATE reports
			SET status = $1, resolution = $2, resolved_by = $3, resolved_at = now()
			WHERE id = $4 AND status = 'pending'
			RETURNING id, reporter_id, target_type, target_id, reason, details, status, resolution, resolved_by, resolved_at, created_at
		)
		SELECT udata.id, udata.reporter_id, reporter.username, udata.target_type, udata.target_id,
		       udata.reason, udata.details, udata.status, udata.resolution, udata.resolved_by,
		       udata.resolved_at, udata.created_at
		FROM updated udata
		JOIN users reporter ON reporter.id = udata.reporter_id`,
		status, strings.TrimSpace(resolution), resolverID, id))
	if err == nil {
		return item, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Report{}, err
	}
	var existingStatus string
	if lookupErr := r.db.QueryRowContext(ctx, `SELECT status FROM reports WHERE id = $1`, id).Scan(&existingStatus); errors.Is(lookupErr, sql.ErrNoRows) {
		return Report{}, ErrNotFound
	} else if lookupErr != nil {
		return Report{}, lookupErr
	}
	return Report{}, ErrConflict
}

func (r *SQLRepository) PendingCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reports WHERE status = 'pending'`).Scan(&count)
	return count, err
}

func scanReport(row interface{ Scan(dest ...any) error }) (Report, error) {
	var item Report
	var resolvedBy sql.NullInt64
	var resolvedAt sql.NullTime
	if err := row.Scan(&item.ID, &item.ReporterID, &item.Reporter, &item.TargetType, &item.TargetID, &item.Reason, &item.Details, &item.Status, &item.Resolution, &resolvedBy, &resolvedAt, &item.CreatedAt); err != nil {
		return Report{}, err
	}
	if resolvedBy.Valid {
		item.ResolvedBy = &resolvedBy.Int64
	}
	if resolvedAt.Valid {
		value := resolvedAt.Time.UTC()
		item.ResolvedAt = &value
	}
	return item, nil
}

func validateStatus(status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "pending"
	}
	switch status {
	case "all", "pending", "resolved", "dismissed":
		return nil
	default:
		return ErrInvalidStatus
	}
}
