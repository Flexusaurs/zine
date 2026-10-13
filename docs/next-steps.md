# ZINE — Next Steps & Backlog

Last updated: 2026-10-13

---

## Immediate — Finish the Current Slice

- [ ] Verify `-race` flag works in local Windows environment (requires cgo on Windows/amd64)
  - If not: remove `-race` from `test-integration` in `make.ps1` only; keep it in CI (runs on Linux)
- [ ] Run `go mod download` once from any machine with unrestricted network access to populate the module cache
- [ ] Review and refine `make.ps1` end-to-end on a clean Windows clone

---

## Next Feature: Route Wiring

- [ ] `internal/post/handler.go` — wire `FanOuter` interface to call `feed.Service.FanOut` on post create
- [ ] `internal/feed/handler.go` — `GET /api/feed`, `POST /api/follows`, `DELETE /api/follows`
- [ ] Register all routes in `cmd/api/main.go` (currently stubbed)
- [ ] Smoke test end-to-end: register → login → post → follow → check feed

---

## P2P Distribution Layer

Code is scaffolded; needs wiring and testing.

- [ ] Register `internal/p2p/pow.go` HTTP handlers in `cmd/api/main.go`
- [ ] Register `internal/p2p/signaling.go` WebSocket route in `cmd/api/main.go`
- [ ] Wire CID assignment into the files upload handler
- [ ] End-to-end test: upload → CID assigned → `GET /api/p2p/resolve` returns seeder → browser fetches via `<zine-img>`
- [ ] Test `cmd/seeder/main.go`: solve PoW → register → serve content → heartbeat loop

---

## Asset Pipeline

- [ ] `cmd/worker/main.go` — Redis `BRPOP` loop consuming `jobs:asset`
- [ ] `pkg/imgproc/optimize.go` — libvips (govips) AVIF + WebP variants, BlurHash
- [ ] ffmpeg subprocess for video → WebM + MP4 thumbnail
- [ ] Update `files` table to store `s3_key_raw` and `s3_key_opt` separately
- [ ] Add `libvips-dev` + `ffmpeg` to GitHub Actions CI runner

---

## E2EE Chat Backend

- [ ] Add `003_chat.sql` migration:
  - `chat_rooms` (id, name, created_at)
  - `chat_members` (room_id, user_id, joined_at)
  - `chat_messages` (id, room_id, sender_id, ciphertext BYTEA, nonce BYTEA, self_destruct_at TIMESTAMPTZ, created_at)
- [ ] `internal/chat/hub.go` — WebSocket hub with room management
- [ ] `internal/chat/handler.go` — WS upgrade, relay encrypted blobs, never inspect plaintext
- [ ] Self-destruct worker in `cmd/worker` — runs every 30s, calls `delete_expired_messages()` SQL function
- [ ] Wire chat routes into `cmd/api/main.go`

---

## Frontend (Vanilla JS + Web Components)

- [ ] Profile layout engine — JSON layout config → CSS grid → rendered segment Web Components
- [ ] CSS sandbox — server-side sanitizer (allowlist properties, block XSS / injection vectors)
- [ ] Segment components:
  - `<zine-bio>` — bio text, pronouns, links
  - `<zine-posts>` — post feed for a profile
  - `<zine-media>` — image/video gallery
  - `<zine-music>` — audio player
  - `<zine-guestbook>` — comments / shout-outs
  - `<zine-links>` — link list
  - `<zine-stats>` — follower/following counts, join date
  - `<zine-topics>` — topic badges
  - `<zine-custom>` — freeform HTML widget (with CSS sandbox)
- [ ] Profile editor — drag-and-drop segment placement with live preview
- [ ] Feed page — home feed + topic feeds
- [ ] Chat UI — `ChatClient` integration + self-destruct toggle

---

## Open Decisions

| Question | Status |
|---|---|
| When to wire Typesense search into handlers? | Deferred — it's in `docker-compose.yml` but not integrated yet |
| Private posts + P2P: should they be origin-only (no seeder announcement)? | Likely yes — confirm |
| Seeder incentive structure: pure volunteer (current) sufficient, or add reputation? | Open |
| CSS sandbox allowlist: needs OWASP CSS injection test suite run before launch | Planned pre-launch |
| File size limit for uploads: 500 MB multipart cap — acceptable for zines/albums/patches? | Confirm with community |

---

## Milestone Summary

| # | Milestone | Status |
|---|---|---|
| 1 | Auth + posts + feed + follows (vertical slice) | ✅ Done |
| 2 | Route wiring + smoke tests | 🔲 Next |
| 3 | P2P distribution layer end-to-end | 🔲 |
| 4 | Asset pipeline (images + video) | 🔲 |
| 5 | E2EE chat backend | 🔲 |
| 6 | Frontend: profile engine + segments | 🔲 |
| 7 | Frontend: feed + chat UI | 🔲 |
| 8 | CSS sandbox + security review | 🔲 |
| 9 | Seeder binary production testing | 🔲 |
| 10 | Launch prep: Typesense, file limits, perf testing | 🔲 |
