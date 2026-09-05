package database

import "context"

type Database struct {
	DSN string
}

func Open(dsn string) (*Database, error) {
	return &Database{DSN: dsn}, nil
}

func (db *Database) Ping(ctx context.Context) error {
	return ctx.Err()
}
