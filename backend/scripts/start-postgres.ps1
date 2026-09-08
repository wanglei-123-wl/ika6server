param(
    [string]$DatabaseUrl = "postgres://ika6:ika6-local-password@localhost:5432/ika6"
)

$ErrorActionPreference = "Stop"

$BackendRoot = Split-Path -Parent $PSScriptRoot
Set-Location $BackendRoot

Write-Host "Starting ika6 PostgreSQL container..."
docker compose -f .\docker-compose.postgres.yml up -d
if ($LASTEXITCODE -ne 0) {
    throw "Failed to start PostgreSQL with Docker Compose. Please start Docker Desktop first, then rerun this script."
}

$env:IKA6_DATABASE_URL = $DatabaseUrl
$env:IKA6_ROOT = $BackendRoot
$env:IKA6_MIGRATIONS_DIR = Join-Path $BackendRoot "migrations"

Write-Host "Waiting for PostgreSQL on localhost:5432..."
$deadline = (Get-Date).AddSeconds(60)
do {
    try {
        $client = [System.Net.Sockets.TcpClient]::new()
        $connect = $client.BeginConnect("127.0.0.1", 5432, $null, $null)
        if ($connect.AsyncWaitHandle.WaitOne(1000)) {
            $client.EndConnect($connect)
            $client.Close()
            Write-Host "PostgreSQL port is reachable."
            break
        }
        $client.Close()
    } catch {
        Start-Sleep -Seconds 1
    }

    if ((Get-Date) -ge $deadline) {
        throw "PostgreSQL did not become reachable on localhost:5432 within 60 seconds."
    }
} while ($true)

Write-Host "Applying database migrations..."
go run .\cmd\migrate

Write-Host "Database is ready. Use this in the same shell when starting the API:"
Write-Host "`$env:IKA6_DATABASE_URL=`"$DatabaseUrl`""
