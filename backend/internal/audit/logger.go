package audit

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

type Entry struct {
	ID        int64     `json:"id"`
	ActorID   int64     `json:"actorId"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	TargetID  int64     `json:"targetId"`
	Details   any       `json:"details"`
	CreatedAt time.Time `json:"createdAt"`
}

type Logger interface {
	Log(actorID int64, action, target string, targetID int64, details any)
	List() []Entry
}

type MemoryLogger struct {
	mu      sync.RWMutex
	entries []Entry
}

func NewMemoryLogger() *MemoryLogger {
	return &MemoryLogger{}
}

func (l *MemoryLogger) Log(actorID int64, action, target string, targetID int64, details any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, Entry{ActorID: actorID, Action: action, Target: target, TargetID: targetID, Details: details, CreatedAt: time.Now().UTC()})
}

func (l *MemoryLogger) List() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	items := append([]Entry(nil), l.entries...)
	for index := 0; index < len(items)/2; index++ {
		opposite := len(items) - 1 - index
		items[index], items[opposite] = items[opposite], items[index]
	}
	return items
}

type SQLLogger struct {
	db *sql.DB
}

var _ Logger = (*SQLLogger)(nil)

func NewSQLLogger(db *sql.DB) (*SQLLogger, error) {
	if db == nil {
		return nil, errors.New("sql database is required")
	}
	return &SQLLogger{db: db}, nil
}

func (l *SQLLogger) Log(actorID int64, action, target string, targetID int64, details any) {
	payload, err := json.Marshal(details)
	if err != nil {
		payload = []byte(`{}`)
	}
	_, _ = l.db.Exec(`
		INSERT INTO audit_logs (actor_id, action, target, target_id, details)
		VALUES ($1, $2, $3, $4, $5)`,
		actorID, action, target, targetID, payload)
}

func (l *SQLLogger) List() []Entry {
	rows, err := l.db.Query(`
		SELECT id, actor_id, action, target, target_id, details, created_at
		FROM audit_logs
		ORDER BY created_at DESC, id DESC
		LIMIT 200`)
	if err != nil {
		return []Entry{}
	}
	defer rows.Close()
	items := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		var details []byte
		if err := rows.Scan(&entry.ID, &entry.ActorID, &entry.Action, &entry.Target, &entry.TargetID, &details, &entry.CreatedAt); err != nil {
			return []Entry{}
		}
		var decoded any
		if err := json.Unmarshal(details, &decoded); err == nil {
			entry.Details = decoded
		}
		items = append(items, entry)
	}
	return items
}
