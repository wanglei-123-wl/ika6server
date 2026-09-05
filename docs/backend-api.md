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

`GET /api/posts/{id}/download`

Downloads the file after the post is approved.

## Search

`GET /api/search?q=keyword`

Searches approved posts.

## Audit

`POST /api/admin/posts/{id}/approve`

`POST /api/admin/posts/{id}/reject`

Requires admin role.
