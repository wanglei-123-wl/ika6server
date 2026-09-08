param(
    [string]$Addr = "0.0.0.0:8080",
    [string]$DatabaseUrl = "postgres://ika6:ika6-local-password@localhost:5432/ika6"
)

$ErrorActionPreference = "Stop"

$BackendRoot = Split-Path -Parent $PSScriptRoot
Set-Location $BackendRoot

$env:IKA6_ADDR = $Addr
$env:IKA6_DATABASE_URL = $DatabaseUrl
$env:IKA6_ROOT = $BackendRoot
$env:IKA6_MIGRATIONS_DIR = Join-Path $BackendRoot "migrations"
$env:IKA6_UPLOAD_DIR = Join-Path (Split-Path -Parent $BackendRoot) "storage\uploads"
$env:IKA6_TEMP_DIR = Join-Path (Split-Path -Parent $BackendRoot) "storage\tmp"
$env:IKA6_PLAY_DIR = Join-Path (Split-Path -Parent $BackendRoot) "storage\play"

Write-Host "Starting PostgreSQL if needed..."
docker compose -f .\docker-compose.postgres.yml up -d
if ($LASTEXITCODE -ne 0) {
    throw "Failed to start PostgreSQL with Docker Compose. Please start Docker Desktop first, then rerun this script."
}

Write-Host "Applying migrations before API startup..."
go run .\cmd\migrate

Write-Host "Starting ika6 API with PostgreSQL persistence on $Addr..."
go run .\cmd\api
