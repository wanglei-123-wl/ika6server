package reputation

import (
	"database/sql"
	"errors"
	"sort"
	"time"
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

func (r *SQLRepository) Add(userID int64, eventType EventType, delta int, reason string) Profile {
	if userID <= 0 {
		return Profile{}
	}
	_, _ = r.db.Exec(`
		INSERT INTO reputation_events (user_id, event_type, delta, reason)
		VALUES ($1, $2, $3, $4)`, userID, string(eventType), delta, reason)
	return r.Get(userID)
}

func (r *SQLRepository) Get(userID int64) Profile {
	if userID <= 0 {
		return Profile{}
	}
	rows, err := r.db.Query(`
		SELECT user_id, event_type, delta, reason, created_at
		FROM reputation_events
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return Profile{UserID: userID, Level: level(0)}
	}
	defer rows.Close()
	return profileFromRows(userID, rows)
}

func (r *SQLRepository) List() []Profile {
	rows, err := r.db.Query(`
		SELECT user_id, event_type, delta, reason, created_at
		FROM reputation_events
		ORDER BY user_id ASC, created_at DESC, id DESC`)
	if err != nil {
		return []Profile{}
	}
	defer rows.Close()
	profiles := make(map[int64]Profile)
	for rows.Next() {
		var event Event
		var eventType string
		if err := rows.Scan(&event.UserID, &eventType, &event.Delta, &event.Reason, &event.CreatedAt); err != nil {
			return []Profile{}
		}
		event.Type = EventType(eventType)
		profile := profiles[event.UserID]
		if profile.UserID == 0 {
			profile.UserID = event.UserID
		}
		profile.Score += event.Delta
		if event.Type == EventDownload {
			profile.Downloads++
		}
		profile.Events = append(profile.Events, event)
		profiles[event.UserID] = profile
	}
	items := make([]Profile, 0, len(profiles))
	for _, profile := range profiles {
		profile.Level = level(profile.Score)
		items = append(items, profile)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Score > items[j].Score
	})
	return items
}

func profileFromRows(userID int64, rows *sql.Rows) Profile {
	profile := Profile{UserID: userID}
	for rows.Next() {
		var event Event
		var eventType string
		if err := rows.Scan(&event.UserID, &eventType, &event.Delta, &event.Reason, &event.CreatedAt); err != nil {
			return Profile{UserID: userID, Level: level(0)}
		}
		event.Type = EventType(eventType)
		event.CreatedAt = event.CreatedAt.UTC()
		profile.Score += event.Delta
		if event.Type == EventDownload {
			profile.Downloads++
		}
		profile.Events = append(profile.Events, event)
	}
	profile.Level = level(profile.Score)
	return profile
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
