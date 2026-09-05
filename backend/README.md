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
IKA6_UPLOAD_DIR=F:\ika6server\storage\uploads
IKA6_TOKEN_SECRET=change-me
```

## Current Status

The first version uses in-memory stores so the API shape can be tested before wiring PostgreSQL into repositories.
