# ZINE — Project Overview

## Purpose & Context

ZINE is a MySpace-inspired DIY social network for the punk/DIY scene. It is self-hosted and infrastructure-light — no CDN, no managed cloud services. Aesthetic and community fit matter as much as technical execution.

Inspiration: MySpace-era user-controlled layouts, bringable CSS, segmented profile widgets.

---

## Tech Stack

| Layer | Choice |
|---|---|
| Backend | Go 1.23+, module path `github.com/zine-social/zine` |
| Router | chi v5 |
| Database | PostgreSQL 16 (pgx/v5, no ORM) |
| Cache / job queue | Redis 7 (Valkey) — sessions, rate limiting, feed cache, async jobs via LPUSH/BRPOP |
| Search | Typesense (self-hostable) |
| Object storage | S3-compatible (MinIO local dev / Cloudflare R2 prod) |
| Frontend | Vanilla JS + Web Components, no framework |

NATS was considered and explicitly dropped — Redis lists are sufficient at DIY-scene scale. Total infrastructure: 4 components (Postgres, Redis, MinIO, Typesense).

---

## Current Implementation State

### Vertical Slice (built, compiles, tests pass)

- User registration and login — bcrypt password hashing, JWT access tokens (15 min), hashed refresh tokens in DB (30 days)
- Post creation, retrieval, soft delete, cursor-based pagination
- Follow graph (follower_id / followee_id in follows table)
- Redis-cached home feed with fan-out on write (ZADD sorted sets, 48 h TTL, 200-item cap)
- Redis token-bucket rate limiter implemented as atomic Lua script
- Middleware: JWT auth guard, rate limiter
- API routes: `/api/auth/*`, `/api/posts/*`, `/api/feed`, `/api/follows`

### Test Architecture

- `testutil.NewTestDB(tb testing.TB)` — accepts both `*testing.T` and `*testing.B`
- Isolated per-test PostgreSQL schemas (`CREATE SCHEMA test_<uuid>`, dropped in `t.Cleanup`)
- Redis on DB index 1 to keep test data off dev data
- `testutil.NewBadRedis(tb)` — tests fail-open behaviour when Redis is unavailable
- Integration tests skip (not fail) when Postgres/Redis are unreachable — safe to run without Docker
- 70 test functions across 6 packages: config (4), auth (24), post (19), feed (10), middleware (4), integration (9) + 1 benchmark
- GitHub Actions CI: 4 jobs — unit, integration, build, lint

### Database Schema (001_init.sql)

- **users**: id (uuid), username (unique, regex-validated), email (unique), password_hash, created_at, updated_at
- **refresh_tokens**: id, user_id (FK), token_hash (SHA-256 of raw token), user_agent, expires_at
- **posts**: id, user_id, body (1–10000 chars), visibility (public/followers/private), created_at, edited_at, deleted_at (soft delete)
- **follows**: follower_id, followee_id (composite PK, self-follow CHECK constraint)

---

## P2P Content Distribution (designed, not yet built)

### Architecture — Three-Tier Hybrid

| Tier | Component | Notes |
|---|---|---|
| 1 | Browser peers | WebRTC DataChannels, ephemeral |
| 2 | Community seeder nodes | Standalone Go binary, PoW-certified, persistent |
| 3 | Origin server | Fallback only + peer registry |

### Content Addressing

All assets keyed by SHA-256 hash: `"sha256-<hex>"`. Deterministic — any two nodes with the same bytes advertise the same CID, natural dedup.

### Seeder Certification (PoW)

- Anyone runs the seeder binary — it solves a PoW challenge on first start
- `SHA-256(prefix + nonce)` must have 20 leading zero bits (~1M hashes, ~1s on modern CPU)
- PoW is Sybil resistance only — no token, no blockchain, no financialisation
- Peers registered in Redis with 5-minute TTL, refreshed by heartbeat every 2 min
- Stale peers expire automatically

### Browser Content Resolution (racing, not sequential)

```
t=0ms   — IndexedDB local cache check (synchronous-ish, zero network cost)
t=0ms   — WebRTC mesh broadcast (fails fast when 0 peers)
t=50ms  — seeder registry query fires
t=100ms — origin fetch fires (guaranteed floor, imperceptible overhead)
```

The original sequential design had a hidden ~300ms dead-time penalty even with 0 peers. The racing design fixes this. Origin delay is adaptive: half the measured origin RTT, clamped to [50ms, 300ms].

Every successful fetch pins to IndexedDB and announces to WebRTC mesh.

### WebRTC Signaling

- Peers on the same page connect to: `wss://origin/api/p2p/signaling?room=profile:username`
- Server is a dumb relay — sees only peer IDs and SDP/ICE, never content
- DataChannel wire protocol: text frames (JSON want/have), binary frames (`[4-byte CID len][CID][content]`)

### Seeder Node Binary (cmd/seeder)

- Self-contained Go binary — no full API stack needed
- Serves pinned CIDs at `GET /content/<cid>` with Range request support
- Heartbeats to origin every 2 min, announces newly pinned CIDs
- Configured via env vars: `ORIGIN_URL`, `PUBLIC_ADDR`, `LISTEN_ADDR`, `STORAGE_DIR`, `STORAGE_GB`

---

## File Structure

```
cmd/
  api/            — entry point, router wiring, graceful shutdown
  seeder/         — standalone seeder binary
  worker/         — (planned) asset processing worker
internal/
  auth/           — service + handler + tests
  post/           — service + handler + tests
  feed/           — service + handler + tests
  middleware/     — JWT auth guard, rate limiter + tests
  integration/    — end-to-end HTTP tests + benchmark
  db/             — pgx pool + Redis client constructors
  config/         — typed env-based config + tests
  p2p/            — pow.go, content_address.go, signaling.go
testutil/         — shared test helpers
migrations/       — 001_init.sql, 002_p2p.sql
web/src/
  p2p/            — content-router.js, zine-img.js, uploader.js
  chat/           — crypto.js (X25519 + AES-GCM E2EE), client.js
  segments/       — posts.js (Web Component)
```

---

## What's Next

1. Complete and refine the PowerShell developer workflow (make.ps1 is primary)
2. Implement the P2P distribution layer (architecture done, code scaffolded)
3. Frontend: profile layout engine, segment system, CSS sandbox
4. E2EE chat: WebSocket relay, self-destruct worker, key management
5. Asset pipeline: libvips + ffmpeg worker, AVIF/WebP output, CID assignment
6. File sharing: upload flow, seeder announcement, signed URLs
