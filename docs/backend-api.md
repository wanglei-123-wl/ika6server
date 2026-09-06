# ika6 Backend API

Base URL for local development:

```text
http://localhost:8080
```

## Health

`GET /api/health`

Returns backend status.

## Auth

`POST /api/auth/register`

```json
{
  "username": "demo",
  "email": "demo@example.com",
  "password": "password123"
}
```

`POST /api/auth/login`

```json
{
  "email": "demo@example.com",
  "password": "password123"
}
```

Use the returned token as:

```text
Authorization: Bearer <token>
```

## Users

`GET /api/users/me`

Returns the current authenticated user.

## Categories

`GET /api/categories`

Returns source-code categories.

## Posts

`GET /api/posts`

`GET /api/posts?status=pending`

`POST /api/posts`

```json
{
  "title": "Example source code",
  "description": "Short description",
  "category": "web"
}
```

New posts start as `pending` and require audit before public download.

In the current in-memory development version, the first registered user becomes `admin` so the audit flow can be tested locally.

## Files

`POST /api/posts/{id}/files`

Upload multipart form field:

```text
file
```

The backend saves the uploaded file to `storage/tmp`, scans it with ClamAV and YARA rules, then moves it to `storage/uploads` only when the scan result is clean. Archive uploads are extracted with 7-Zip into a temporary scan directory and scanned recursively before approval. The response includes the file SHA256 and scan metadata.

The upload response also includes a static sandbox report. This first sandbox stage never executes user code inside the API process; it inspects archives and file content for executable/script files, persistence indicators, credential-theft keywords, and network download execution patterns.

`GET /api/posts/{id}/download`

Downloads the file after the post is approved.

## Search

`GET /api/search?q=keyword`

Searches approved posts.

## Audit

`POST /api/admin/posts/{id}/approve`

`POST /api/admin/posts/{id}/reject`

Requires admin role.

## Download Blocklist

`GET /api/admin/blocklist`

`POST /api/admin/blocklist`

```json
{
  "sha256": "64-character file sha256",
  "reason": "malware confirmed by manual review"
}
```

Files matching a blocked SHA256 are rejected during upload and blocked during download.

## Reputation

`GET /api/users/me/reputation`

Returns the current user's reputation score, level, download count, and event history.

`GET /api/admin/reputation`

Returns all in-memory reputation profiles for admins.
