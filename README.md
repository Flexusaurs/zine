# Zine Project Structure
Zine is a tiny social network for the DIY scene, utilizing p2p technologies to make media sharing easier on the user and server altogether

## Directories

```
zine/
├── cmd/              # Main entry points
│   └── api/          # API server
├── internal/         # Private application code
│   ├── api/          # API handlers and middleware
│   ├── feed/         # Feed service logic
│   ├── auth/         # Auth service (not yet implemented)
│   └── middleware/   # Shared middleware
├── pkg/              # Public libraries (if any)
├── scripts/          # Dev scripts
├── tests/            # E2E and integration tests
├── docker/           # Docker configs
└── docs/             # Documentation
```
