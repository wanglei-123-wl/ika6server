# ika6 Backend

This is the Go backend for the ika6 source-code community.

## Run

```powershell
cd F:\ika6server\backend
go run .\cmd\api
```

Default address:

```text
http://localhost:8080
```

Local network example:

```powershell
$env:IKA6_ADDR="0.0.0.0:8080"
go run .\cmd\api
```

## Environment

```text
IKA6_ADDR=0.0.0.0:8080
IKA6_DATABASE_URL=postgres://user:password@localhost:5432/ika6
IKA6_MIGRATIONS_DIR=F:\ika6server\backend\migrations
IKA6_ADMIN_ACCOUNT=yaochenAi.18700021044.com@#$%
IKA6_ADMIN_PASSWORD_HASH=replace-with-output-of-go-run-cmd-hash-password
IKA6_UPLOAD_DIR=F:\ika6server\storage\uploads
IKA6_TEMP_DIR=F:\ika6server\storage\tmp
IKA6_PLAY_DIR=F:\ika6server\storage\play
IKA6_TOKEN_SECRET=change-me
IKA6_CLAMSCAN_BIN=C:\Program Files\ClamAV\clamscan.exe
IKA6_CLAMAV_DB_DIR=F:\ika6server\storage\clamav-db
IKA6_7ZIP_BIN=C:\Program Files\7-Zip\7z.exe
IKA6_YARA_BIN=C:\Users\Administrator\AppData\Local\Microsoft\WinGet\Packages\VirusTotal.YARA_Microsoft.Winget.Source_8wekyb3d8bbwe\yara64.exe
IKA6_YARA_RULES=F:\ika6server\deployments\yara\source_rules.yar
```

## Database Migration

With PostgreSQL available, run:

```powershell
docker compose -f .\docker-compose.postgres.yml up -d
$env:IKA6_DATABASE_URL="postgres://ika6:ika6-local-password@localhost:5432/ika6"
go run .\cmd\migrate
```

The API applies pending migrations during startup. `IKA6_DATABASE_URL`, `IKA6_TOKEN_SECRET`, `IKA6_ADMIN_ACCOUNT`, and `IKA6_ADMIN_PASSWORD_HASH` are required for API startup; the backend no longer falls back to in-memory storage in `cmd/api`.

Generate the administrator password hash with:

```powershell
go run .\cmd\hash-password your-password
```

The PostgreSQL integration test is skipped when `IKA6_DATABASE_URL` is absent:

```powershell
go test .\internal\database -run TestPostgresIntegration
```

The Compose file is only a local verification environment; it is not started by the API.

## Local PostgreSQL Startup

For local testing with persistent users, games, comments, likes, and upload metadata, start the API through the PostgreSQL helper script:

```powershell
cd F:\游戏社区\ika6server\backend
.\scripts\start-api-postgres.ps1
```

This script starts `docker-compose.postgres.yml`, sets `IKA6_DATABASE_URL`, applies migrations, and then starts `cmd/api`.

To only start and migrate PostgreSQL:

```powershell
.\scripts\start-postgres.ps1
```

To verify that the database path is reachable:

```powershell
.\scripts\verify-postgres.ps1
```

If the API is started with plain `go run .\cmd\api`, set the required environment variables in that same shell/process first.

## Current Status

The backend requires PostgreSQL persistence for `cmd/api` startup. In-memory stores remain only for unit tests and isolated package-level development.

Uploaded files are written to a temporary directory, scanned with ClamAV and optional YARA rules, then moved into the upload directory only when the scan is clean. Archive uploads are extracted with 7-Zip into a temporary scan directory and scanned recursively before approval.

Online play builds are deployed from uploaded `.zip` build packages into `IKA6_PLAY_DIR`. The build package must contain an `index.html`; play assets are served from `/play/games/{gameId}/...` only after the game is published.

Second-stage safety controls are included in the in-memory development backend:

- YARA rules run as part of upload scanning.
- Static sandbox reports inspect uploaded files and `.zip` archives without executing user code.
- Download blocklist rejects known-bad SHA256 values at upload and download time.
- User reputation records registration, clean uploads, rejected uploads, approvals, rejections, and downloads.
