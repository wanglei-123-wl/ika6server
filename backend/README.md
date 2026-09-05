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
IKA6_TEMP_DIR=F:\ika6server\storage\tmp
IKA6_TOKEN_SECRET=change-me
IKA6_CLAMSCAN_BIN=C:\Program Files\ClamAV\clamscan.exe
IKA6_CLAMAV_DB_DIR=F:\ika6server\storage\clamav-db
IKA6_7ZIP_BIN=C:\Program Files\7-Zip\7z.exe
IKA6_YARA_BIN=C:\Users\Administrator\AppData\Local\Microsoft\WinGet\Packages\VirusTotal.YARA_Microsoft.Winget.Source_8wekyb3d8bbwe\yara64.exe
IKA6_YARA_RULES=F:\ika6server\deployments\yara\source_rules.yar
```

## Current Status

The first version uses in-memory stores so the API shape can be tested before wiring PostgreSQL into repositories.

Uploaded files are written to a temporary directory, scanned with ClamAV and optional YARA rules, then moved into the upload directory only when the scan is clean. Archive uploads are extracted with 7-Zip into a temporary scan directory and scanned recursively before approval.

Second-stage safety controls are included in the in-memory development backend:

- YARA rules run as part of upload scanning.
- Static sandbox reports inspect uploaded files and `.zip` archives without executing user code.
- Download blocklist rejects known-bad SHA256 values at upload and download time.
- User reputation records registration, clean uploads, rejected uploads, approvals, rejections, and downloads.
