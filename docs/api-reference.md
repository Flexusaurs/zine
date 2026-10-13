# ZINE — API Reference

## Authentication

All authenticated routes require:
```
Authorization: Bearer <access_token>
```

Access tokens expire in **15 min**. Use the refresh endpoint to rotate.

---

## Auth Endpoints

### POST /api/auth/register

```json
// Request
{
  "username": "itay",
  "email": "itay@example.com",
  "password": "s3cure!pw"
}

// Response 201
{
  "user": {
    "id": "uuid",
    "username": "itay",
    "email": "itay@example.com",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "tokens": {
    "access_token": "eyJ...",
    "refresh_token": "hex...",
    "expires_at": "2024-01-01T00:15:00Z"
  }
}
```

Errors: `400` missing fields · `409` email or username taken · `422` weak password (< 8 chars)

---

### POST /api/auth/login

```json
// Request
{ "email": "itay@example.com", "password": "s3cure!pw" }

// Response 200 — same shape as register response
```

Errors: `400` bad body · `401` invalid credentials

---

### POST /api/auth/refresh

```json
// Request
{ "refresh_token": "<raw token>" }

// Response 200
{
  "tokens": {
    "access_token": "eyJ...",
    "refresh_token": "hex...",
    "expires_at": "..."
  }
}
```

Errors: `400` missing token · `401` invalid or expired token

---

### POST /api/auth/logout

```json
// Request
{ "refresh_token": "<raw token>" }

// Response 204 No Content
```

---

## Posts

### GET /api/posts

Public timeline. Query params: `before=<RFC3339>` (cursor), `limit=1-50` (default 20).

```json
// Response 200
{
  "posts": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "username": "itay",
      "body": "first post!",
      "visibility": "public",
      "created_at": "...",
      "edited_at": null
    }
  ],
  "next_cursor": "2024-01-01T00:00:00Z"
}
```

---

### GET /api/posts/{id}

Single post. Returns `404` if the post is private and the requester is not the owner.

---

### POST /api/posts *(auth required)*

```json
// Request
{
  "body": "first post!",
  "visibility": "public"
}
// visibility: "public" | "followers" | "private"  (default: "public")

// Response 201
{ "post": { ... } }
```

Errors: `400` bad body · `422` body empty or > 10,000 chars

---

### DELETE /api/posts/{id} *(auth required)*

Soft delete — sets `deleted_at`. Only the owner can delete.

```
Response 204 No Content
Errors: 404 not found or not owner
```

---

## Feed & Follows

### GET /api/feed *(auth required)*

Home feed — posts from followed users only, reverse-chronological.
Query params: `before=<RFC3339>` (cursor).

```json
// Response 200
{ "posts": [ ... ], "next_cursor": "..." }
```

Served from Redis cache (`ZREVRANGEBYSCORE`). Falls back to DB on cache miss.

---

### POST /api/follows *(auth required)*

```json
// Request
{ "followee_id": "uuid" }

// Response 201
{ "status": "following" }
```

Errors: `400` self-follow · `409` already following

---

### DELETE /api/follows *(auth required)*

```json
// Request
{ "followee_id": "uuid" }

// Response 204
```

---

## P2P (designed — not yet wired into API)

### GET /api/p2p/challenge

Returns a PoW puzzle for seeder node registration.

```json
{
  "id": "uuid",
  "prefix": "a3f9c2...",
  "difficulty": 20,
  "expires_at": "..."
}
```

---

### POST /api/p2p/register

```json
// Request
{
  "challenge_id": "uuid",
  "nonce": "000000a3...",
  "peer": {
    "id": "hex8",
    "addr": "host:9090",
    "storage_gb": 10
  },
  "cids": ["sha256-...", "sha256-..."]
}

// Response 200
{ "status": "registered" }
```

Errors: `403` bad PoW solution or expired challenge

---

### GET /api/p2p/resolve?cid=sha256-...

```json
// Response 200
{
  "peers": [
    { "id": "hex8", "addr": "host:9090", "certified_at": "..." }
  ]
}

// Response 404 (no peers) — client falls back to origin
{ "peers": [] }
```

---

### POST /api/p2p/heartbeat?peer_id=...

```json
// Body (optional)
{ "new_cids": ["sha256-..."] }

// Response 200
{ "status": "ok" }
```

---

### POST /api/p2p/announce

Browser clients announce newly cached CIDs.

```json
// Request
{ "cids": ["sha256-..."], "source": "browser" }

// Response 200
```

---

### WS /api/p2p/signaling?room=profile:username

WebRTC signaling relay. Message types: `hello`, `peer_joined`, `peer_left`, `offer`, `answer`, `ice`.

---

## Rate Limits

| Route | Burst | Sustained |
|---|---|---|
| `POST /api/posts` | 10 | 0.5 / s |
| `POST /api/auth/login` | 5 | 0.05 / s |
| `POST /api/files/upload` | 5 | 0.1 / s |
| `POST /api/chat/send` | 30 | 2 / s |
| `GET /api/feed` | 60 | 10 / s |
| everything else | 100 | 10 / s |

`429` response includes `Retry-After: 5`. Rate limiter **fails open** if Redis is unavailable.

---

## Error Shape

All errors use a consistent envelope:

```json
{ "error": "human-readable message" }
```

HTTP status codes follow REST conventions: `400` bad input, `401` unauthenticated, `403` forbidden, `404` not found, `409` conflict, `422` unprocessable, `429` rate limited, `500` internal error.
