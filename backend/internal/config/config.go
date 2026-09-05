package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Addr        string
	DatabaseURL string
	UploadDir   string
	TokenSecret string
}

func Load() Config {
	addr := env("IKA6_ADDR", ":8080")
	root := env("IKA6_ROOT", ".")

	return Config{
		Addr:        addr,
		DatabaseURL: os.Getenv("IKA6_DATABASE_URL"),
		UploadDir:   env("IKA6_UPLOAD_DIR", filepath.Join(root, "..", "storage", "uploads")),
		TokenSecret: env("IKA6_TOKEN_SECRET", "dev-secret-change-me"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
