package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Addr        string
	DatabaseURL string
	UploadDir   string
	TempDir     string
	TokenSecret string
	ClamScanBin string
	ClamAVDBDir string
	SevenZipBin string
	YaraBin     string
	YaraRules   string
}

func Load() Config {
	addr := env("IKA6_ADDR", ":8080")
	root := env("IKA6_ROOT", ".")

	return Config{
		Addr:        addr,
		DatabaseURL: os.Getenv("IKA6_DATABASE_URL"),
		UploadDir:   env("IKA6_UPLOAD_DIR", filepath.Join(root, "..", "storage", "uploads")),
		TempDir:     env("IKA6_TEMP_DIR", filepath.Join(root, "..", "storage", "tmp")),
		TokenSecret: env("IKA6_TOKEN_SECRET", "dev-secret-change-me"),
		ClamScanBin: env("IKA6_CLAMSCAN_BIN", `C:\Program Files\ClamAV\clamscan.exe`),
		ClamAVDBDir: env("IKA6_CLAMAV_DB_DIR", filepath.Join(root, "..", "storage", "clamav-db")),
		SevenZipBin: env("IKA6_7ZIP_BIN", `C:\Program Files\7-Zip\7z.exe`),
		YaraBin:     env("IKA6_YARA_BIN", `C:\Users\Administrator\AppData\Local\Microsoft\WinGet\Packages\VirusTotal.YARA_Microsoft.Winget.Source_8wekyb3d8bbwe\yara64.exe`),
		YaraRules:   env("IKA6_YARA_RULES", filepath.Join(root, "..", "deployments", "yara", "source_rules.yar")),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
