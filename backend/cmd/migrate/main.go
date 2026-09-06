package main

import (
	"context"
	"log"
	"time"

	"github.com/wanglei-123-wl/ika6server/backend/internal/config"
	"github.com/wanglei-123-wl/ika6server/backend/internal/database"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("IKA6_DATABASE_URL is required")
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	migrations, err := database.LoadMigrations(cfg.MigrationsDir)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.ApplyMigrations(ctx, db.SQL(), migrations); err != nil {
		log.Fatal(err)
	}
	log.Printf("applied %d migrations", len(migrations))
}
