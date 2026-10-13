# ZINE — Ways of Working

## Dev Environment

- **OS**: Windows, PowerShell is the primary shell
- **Docker Desktop**: installed and available
- **CI**: GitHub Actions (4 jobs: unit, integration, build, lint)
- `make.ps1` is the **primary** build script — all workflows default to PowerShell
- `Makefile` exists as a secondary fallback for WSL2, Git Bash, Linux, or macOS

---

## One-Time Setup (Windows)

```powershell
# Allow running local scripts
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# Prevent CRLF line endings in .go files (causes gofmt diffs)
git config core.autocrlf input
```

---

## make.ps1 Quick Reference

Run all targets from the repo root:

```powershell
.\make.ps1 <target>
```

| Target | What it does |
|---|---|
| `build` | Compiles to `bin\api.exe` |
| `run` | Build and run |
| `dev` | Sets `DATABASE_URL`, `REDIS_URL`, `JWT_SECRET`, `ENV` then `go run` |
| `test-unit` | Unit tests only (no DB / Redis) — runs `./internal/config/...` subset |
| `test-integration` | All packages with `TEST_DATABASE_URL` + `TEST_REDIS_URL` set |
| `test-cover` | Coverage report, opens `.\coverage\coverage.html` in your browser |
| `ci-test` | `docker-test-up` → tests → `docker-test-down` in a `try/finally` block |
| `lint` | golangci-lint |
| `fmt` | gofmt -w |
| `vet` | go vet |
| `tidy` | go mod tidy |
| `check` | fmt + vet + lint |
| `docker-up` | Start dev stack (Postgres, Redis, MinIO, Typesense) |
| `docker-down` | Stop dev stack |
| `docker-up-all` | Start dev stack + api + worker |
| `docker-test-up` | Start isolated test stack (Postgres :5433, Redis :6380) |
| `docker-test-down` | Stop test stack |
| `migrate` | Run migrations via `docker compose exec` (no local psql needed) |
| `db-shell` | `psql` shell inside the DB container |
| `redis-shell` | `redis-cli` shell inside the Redis container |
| `clean` | Remove `bin\` |
| `clean-all` | Remove `bin\` + `.coverage\` |
| `deps` | `go mod download` |

### Env vars set internally by make.ps1

`dev` and `ci-test` set these automatically using the `WithEnv` helper (restores originals on exit):

```
DATABASE_URL  = postgres://postgres:postgres@localhost:5432/zine?sslmode=disable
REDIS_URL     = redis://localhost:6379/0
JWT_SECRET    = dev-secret-change-in-production
ENV           = development
```

Test targets use:
```
TEST_DATABASE_URL = postgres://postgres:postgres@localhost:5433/zine?sslmode=disable
TEST_REDIS_URL    = redis://localhost:6380/1
```

---

## Docker Compose Setup

### Dev stack (docker-compose.yml)

| Service | Port | Notes |
|---|---|---|
| `db` (Postgres 16) | 5432 | Migrations at `/migrations` inside container |
| `redis` (Redis 7) | 6379 | |
| `minio` | 9000 / 9001 | S3-compatible object storage |
| `typesense` | 8108 | Full-text search |
| `api` | 8080 | Built from `cmd/api` |
| `worker` | — | Built from `cmd/worker` (planned) |

### Test stack (docker-compose.test.yml)

| Service | Port | Notes |
|---|---|---|
| `db` (Postgres 16) | 5433 | Isolated — different port from dev |
| `redis` (Redis 7) | 6380 | DB index 1 |

> Migrations are mounted at `/migrations` (not `/docker-entrypoint-initdb.d`) so they run explicitly via `.\make.ps1 migrate`.

---

## Running Tests

### Quick unit tests (no infrastructure needed)

```powershell
.\make.ps1 test-unit
```

### Full test suite with Docker

```powershell
.\make.ps1 docker-test-up
.\make.ps1 test-integration
.\make.ps1 docker-test-down
```

Or run everything atomically:
```powershell
.\make.ps1 ci-test
```

### Coverage

```powershell
.\make.ps1 test-cover
# opens .coverage\coverage.html in your default browser automatically
```

---

## Windows-Specific Notes

- Binary output is `bin\api.exe` (both `make.ps1` and `Makefile` detect Windows automatically)
- `-race` flag requires cgo on Windows/amd64. If it fails locally, remove `-race` from `test-integration` in `make.ps1` only — keep it in CI (which runs on Linux)
- `test-watch` (via `entr`) is not available on Windows. Use VS Code's built-in test runner or Air for live reload instead
- `migrate` uses `docker compose exec` to run psql inside the container — no local psql client needed

---

## Go Module Cache (Proxy Workaround)

The Go module proxy (`proxy.golang.org`) may be blocked in some network environments. If `go mod download` fails:

1. Download the required tarballs from `github.com` mirrors (e.g., `codeload.github.com/golang/crypto/...`)
2. Extract into `/root/go/pkg/mod/golang.org/x/` (or `%USERPROFILE%\go\pkg\mod\...` on Windows)
3. Create the module cache zip and metadata files manually
4. Set env vars before any `go` commands:
   ```powershell
   $env:GOPROXY    = "file:///C:/Users/you/go/pkg/mod/cache/download,off"
   $env:GONOSUMDB  = "*"
   $env:GONOSUMCHECK = "*"
   ```

If you have any machine with unrestricted network access, run `go mod download` there once and copy the entire `%USERPROFILE%\go\pkg\mod\cache\download` directory across.

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| Redis lists over NATS | Fewer moving parts; sufficient at DIY-scene scale |
| Race all content tiers concurrently | Sequential fallback wastes ~300ms even with 0 P2P peers |
| PoW as social mechanism, not economics | Sybil resistance without tokens or blockchain; fits DIY ethos |
| Per-test PostgreSQL schemas | Prevents flaky cross-test state contamination |
| Integration tests skip (not fail) without infra | CI unit job passes without Docker |
| `testing.TB` throughout testutil | `*testing.T` and `*testing.B` both use the same helpers |
| Vanilla JS + Web Components | No framework dependency; matches the project's DIY build-from-scratch ethos |
