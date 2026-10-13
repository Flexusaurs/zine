# ZINE — System Architecture

## API Service (cmd/api)

- chi v5 router with chi middleware: `RequestID`, `RealIP`, `Logger`, `Recoverer`, 30s timeout
- Route groups: public (no auth) and authenticated (JWT guard middleware)
- Graceful shutdown via `signal.NotifyContext` (SIGTERM / SIGINT), 10s drain window
- Config loaded from environment variables via `internal/config`; `JWT_SECRET` required in production

---

## Auth Service (internal/auth)

- Registration: bcrypt (`DefaultCost`), unique constraint on email + username
- Access tokens: HS256 JWT, 15 min TTL, claims: `uid` (user UUID), `usr` (username)
- Refresh tokens: 32 random bytes, stored as SHA-256 hash in `refresh_tokens` table, 30-day TTL
- Refresh is atomic DELETE-then-INSERT (single-use, replay-safe)
- Logout: DELETE refresh token by hash
- `ValidateAccessToken` returns typed `Claims` struct; used by `RequireAuth` middleware

---

## Feed Service (internal/feed)

- **Fan-out on write**: when a post is created, `ZADD` to each follower's `feed:<userID>` sorted set (score = unix timestamp)
- Feed capped at 200 items (`ZREMRANGEBYRANK`), TTL 48 h
- **Read**: `ZREVRANGEBYSCORE` with cursor (before timestamp) — Redis cache hit, DB fallback on miss
- Follow/unfollow: INSERT/DELETE in `follows` table, immediate effect on future fan-outs

---

## Rate Limiting (internal/middleware/ratelimit)

- Redis token bucket implemented as a single Lua script (atomic — no race conditions)
- Bucket key: `rl:u:<userID>:<method>:<path>` (authenticated) or `rl:ip:<ip>:<method>:<path>` (anonymous)
- Per-route rules: `POST /api/posts` (10 burst, 0.5/s), `POST /api/auth/login` (5 burst, 0.05/s), default (100 burst, 10/s)
- **Fail-open**: if Redis is unreachable, request passes through (logged, not blocked)
- 429 response sets `Retry-After: 5` header

---

## Database Schema

### users
```sql
id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
username    TEXT UNIQUE NOT NULL,  -- ^[a-zA-Z0-9_-]{2,30}$
email       TEXT UNIQUE NOT NULL,
password_hash TEXT NOT NULL,
created_at  TIMESTAMPTZ DEFAULT NOW(),
updated_at  TIMESTAMPTZ DEFAULT NOW()
```
Indexes: `users_email` (lower(email)), `users_username` (lower(username))
Trigger: `set_updated_at()` on UPDATE

### refresh_tokens
```sql
id          UUID PRIMARY KEY,
user_id     UUID REFERENCES users,
token_hash  TEXT UNIQUE NOT NULL,  -- SHA-256 of raw token
user_agent  TEXT,
expires_at  TIMESTAMPTZ NOT NULL
```
Indexes: `refresh_tokens_user`, `refresh_tokens_expiry`
Helper: `prune_expired_tokens()` function

### posts
```sql
id          UUID PRIMARY KEY,
user_id     UUID REFERENCES users,
body        TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 10000),
visibility  TEXT NOT NULL DEFAULT 'public'
              CHECK (visibility IN ('public','followers','private')),
created_at  TIMESTAMPTZ DEFAULT NOW(),
edited_at   TIMESTAMPTZ,
deleted_at  TIMESTAMPTZ  -- soft delete
```
Indexes: `posts_user_created` (user_id, created_at DESC WHERE deleted_at IS NULL), `posts_public_created` (created_at DESC WHERE visibility = 'public' AND deleted_at IS NULL)

### follows
```sql
follower_id UUID REFERENCES users,
followee_id UUID REFERENCES users,
PRIMARY KEY (follower_id, followee_id),
CHECK (follower_id != followee_id)
```
Index: `follows_followee`

### content_ids (P2P — 002_p2p.sql)
```sql
cid         TEXT PRIMARY KEY,  -- "sha256-<hex>"
file_id     UUID REFERENCES files,
mime_type   TEXT,
size_bytes  BIGINT,
created_at  TIMESTAMPTZ DEFAULT NOW()
```

### seeder_nodes (P2P — 002_p2p.sql)
```sql
id          TEXT PRIMARY KEY,  -- self-assigned
addr        TEXT NOT NULL,
pubkey      TEXT,
storage_gb  NUMERIC,
certified_at TIMESTAMPTZ,
last_seen   TIMESTAMPTZ
```
Helper: `prune_stale_seeders(cutoff INTERVAL DEFAULT '10 minutes')`

### seeder_cids (P2P — 002_p2p.sql)
```sql
seeder_id   TEXT REFERENCES seeder_nodes,
cid         TEXT REFERENCES content_ids,
PRIMARY KEY (seeder_id, cid),
announced_at TIMESTAMPTZ DEFAULT NOW()
```
Index: `seeder_cids_by_cid`

---

## P2P Layer (internal/p2p + cmd/seeder + web/src/p2p)

### Registry (internal/p2p/pow.go)

| Method | Action |
|---|---|
| `IssueChallenge` | Random 16-byte prefix, difficulty 20, 10 min TTL in Redis |
| `VerifySolution` | Verify `SHA-256(prefix+nonce)` has 20 leading zero bits, then register peer |
| `AnnounceCID` | `SADD cid:peers:<cid> <peerID>` in Redis |
| `PeersForCID` | `SRANDMEMBERN` up to 10 peer IDs, hydrate from `peer:<peerID>` hash |
| `Heartbeat` | Refreshes peer TTL to 5 min (peers vanish if heartbeat stops) |

### Content Addressing (internal/p2p/content_address.go)

```go
CID([]byte) string         // SHA-256 → "sha256-<hex>"
StorePath(baseDir, cid)    // two-char bucket dirs to avoid flat directory explosion
PinFile(baseDir, ReadSeeker) // idempotent, returns CID
```

### WebRTC Signaling (internal/p2p/signaling.go)

- `SignalingHub` manages rooms keyed by page context (e.g. `"profile:username"`)
- Upgrades WS, assigns UUID peer IDs, relays offer/answer/ICE between peers
- Server sees **only** peer IDs and SDP/ICE metadata — never content
- Peer write pump: buffered channel (32), drops with warning on full

### Browser Content Router (web/src/p2p/content-router.js)

```
Singleton ContentRouter shared across all page elements

Resolution order (all racing):
  IndexedDB hit             → instant
  WebRTC mesh broadcast     → t+0ms  (fails fast with 0 peers)
  Seeder HTTP query         → t+50ms
  Origin fetch              → t+100ms (guaranteed floor)

All successful fetches: pin to IndexedDB + announce to mesh
```

### `<zine-img>` Web Component (web/src/p2p/zine-img.js)

```html
<zine-img cid="sha256-..." alt="..." width="..." height="...">
```

Drop-in replacement for `<img>`. Resolves via `ContentRouter`. Shows blur placeholder while loading, revokes object URL on disconnect.

### Uploader (web/src/p2p/uploader.js)

1. Computes SHA-256 CID client-side before upload (`crypto.subtle.digest`)
2. POSTs multipart to origin
3. Verifies server CID matches client CID (upload rejected on mismatch)
4. Immediately pins raw bytes locally and announces to mesh

---

## E2EE Chat (web/src/chat)

| Concern | Implementation |
|---|---|
| Key exchange | X25519 via WebCrypto API (`crypto.subtle`) — no external libraries |
| Key storage | Private keys stored in IndexedDB, **never sent to server** |
| Public keys | Uploaded to server (`users.pubkey` column) |
| Shared key derivation | X25519 ECDH → AES-GCM 256-bit key |
| Message encryption | AES-GCM with 12-byte random nonce |
| Server storage | `(ciphertext BYTEA, nonce BYTEA)` only — never sees plaintext |
| Self-destruct | `self_destruct_at` column; background worker runs every 30s |

Wire format (JSON over WebSocket):
```json
{
  "type": "msg",
  "room_id": "uuid",
  "ct": "<base64 ciphertext>",
  "nonce": "<base64 12-byte nonce>",
  "self_destruct_after": 60
}
```

`ChatClient` extends `EventTarget`. Auto-decrypts incoming messages, fires `'message'` `CustomEvent` with plaintext.

---

## Asset Pipeline (designed, not yet built)

```
Upload flow:
  Browser  →  POST multipart  →  Go API validates MIME
           →  upload raw to S3
           →  enqueue job via Redis LPUSH
           →  return file_id immediately (async)

Worker (cmd/worker):
  BRPOP jobs:asset
  → fetch raw from S3
  → libvips (govips): AVIF quality 72 + WebP quality 78 at 400w / 800w / 1200w
  → ffmpeg: WebM + MP4 thumbnail for video
  → BlurHash computed and stored
  → CID assigned from processed output (content-addressed after optimization)
  → upload to S3 opt/
  → update DB files table
```

CI additions needed: `libvips-dev` + `ffmpeg` on GitHub Actions `ubuntu-latest`.
