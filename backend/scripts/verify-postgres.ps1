param(
    [string]$DatabaseUrl = "postgres://ika6:ika6-local-password@localhost:5432/ika6"
)

$ErrorActionPreference = "Stop"

$BackendRoot = Split-Path -Parent $PSScriptRoot
Set-Location $BackendRoot

$env:IKA6_DATABASE_URL = $DatabaseUrl
$env:IKA6_ROOT = $BackendRoot
$env:IKA6_MIGRATIONS_DIR = Join-Path $BackendRoot "migrations"

Write-Host "Checking Docker PostgreSQL container..."
docker compose -f .\docker-compose.postgres.yml ps
if ($LASTEXITCODE -ne 0) {
    throw "Docker Compose check failed. Please start Docker Desktop first, then rerun this script."
}

Write-Host "Checking PostgreSQL TCP port..."
$client = [System.Net.Sockets.TcpClient]::new()
try {
    $client.Connect("127.0.0.1", 5432)
    Write-Host "localhost:5432 is reachable."
} finally {
    $client.Close()
}

Write-Host "Running migrations as a persistence smoke test..."
go run .\cmd\migrate

Write-Host "Running database integration tests..."
go test -count=1 .\internal\database -run TestPostgresIntegration

Write-Host "PostgreSQL persistence preflight passed."
