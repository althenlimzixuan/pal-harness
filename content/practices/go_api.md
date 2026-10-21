# Go API Practices (Clean Architecture)

## Tech Stack

| Concern | Library |
|---|---|
| HTTP framework | `github.com/gin-gonic/gin` |
| Dependency injection | `github.com/google/wire` |
| Logging | `github.com/sirupsen/logrus` (JSON format enforced) |
| ORM | `gorm.io/gorm` + `gorm.io/driver/mysql` |
| Cache | `github.com/go-redis/redis/v9` |
| Config | `github.com/joho/godotenv` + `github.com/kelseyhightower/envconfig` |
| Mocks | `github.com/vektra/mockery/v2` |
| API docs | `github.com/swaggo/gin-swagger` |
| Migrations | `github.com/golang-migrate/migrate/v4` |

## Project Structure

```
cmd/api/
  main.go             # entrypoint — wire.Build call only
  wire.go             # wire providers
  wire_gen.go         # generated, never edit manually
internal/
  domain/             # entities, value objects, domain errors — zero framework imports
  usecase/            # business logic interfaces + implementations
  repository/         # repository interfaces (defined here, implemented in infrastructure)
  delivery/http/
    handler/          # gin handlers — thin, delegate to usecase
    middleware/       # auth, request-id, logging, recovery, rate-limit
    router/           # route registration + versioning (/api/v1/...)
  infrastructure/
    database/         # gorm repository implementations
    cache/            # redis implementations
    config/           # envconfig structs
    logger/           # logrus setup
pkg/
  apierror/           # RFC 7807 Problem Details type + central error mapper
  response/           # standard JSON response builders
  validator/          # input validation wrappers
migrations/           # golang-migrate SQL files (up + down)
docs/                 # swaggo generated — never edit manually
docker/               # Dockerfiles
```

**Dependency rule**: `domain` ← `usecase` ← `repository interface` ← `infrastructure`. Infrastructure implements domain interfaces; it never defines them.

## Make Targets

```bash
make dev           # hot reload (air)
make build         # compile binary
make test          # unit tests
make test-int      # integration tests (tag: integration)
make test-e2e      # e2e tests (tag: e2e)
make coverage      # HTML coverage report, gate: 80%
make lint          # golangci-lint run ./...
make mock          # mockery --all --keeptree
make migrate-up    # golang-migrate up
make migrate-down  # golang-migrate down 1
make migrate-new name=<name>
make docs          # swag init
make wire          # wire ./cmd/api/...
```

## Authentication

Multi-surface layered middleware (one chain, three mechanisms):

| Surface | Mechanism | Key Config |
|---|---|---|
| Web / Mobile | JWT Bearer | `JWT_ACCESS_TOKEN_TTL` (default `15m`, ENV-configurable) |
| Admin | OAuth2/OIDC via Keycloak | Self-hosted, do not build identity mgmt from scratch |
| B2B | API Key | Argon2 hash in DB, scoped, rate-limited per key |

RBAC across all surfaces. Roles in DB, cached in Redis with short TTL.
Refresh tokens stored in Redis, rotated on each use.

## Error Handling

All errors return `application/problem+json` (RFC 7807):

```json
{
  "type": "https://api.example.com/problems/validation-error",
  "title": "Validation Error",
  "status": 422,
  "detail": "email is required",
  "instance": "/api/v1/users"
}
```

- Sentinel domain errors in `internal/domain/errors.go`
- Central mapper in `pkg/apierror/` converts domain errors → HTTP status
- Never expose stack traces or internal messages to consumers
- `errors.Is` / `errors.As` only — never string-match errors

## API Design

- URL versioning: `/api/v1/`, `/api/v2/`
- Success: `{ "data": ..., "meta": { "page": ..., "total": ... } }`
- Pagination: cursor-based preferred over offset
- POST endpoints that create resources accept `Idempotency-Key` header
- Swagger annotations on all handlers — run `make docs` after any handler change
- **Contract sync rule** (the split rule in `product_docs.md`): the provider's PR regenerates `openapi.yaml` (`make docs`) and keeps Swagger UI served at `/swagger/index.html` (non-prod); CI fails if the committed spec drifts from a fresh `make docs`. In the same piece of work, the SSOT repo's PR updates its `doc/openapi.yaml` copy and `doc/API.md`.

## Wire (Dependency Injection)

- `wire.go` declares providers, `wire_gen.go` is generated — never edit manually
- Each layer has a `ProviderSet` in its package
- Run `make wire` after adding any new dependency
- All constructors accept interfaces, not concrete types (except infrastructure layer)

## DB Migrations

- One `{timestamp}_{description}.up.sql` + `.down.sql` per change
- Never modify a migration that has run in any environment
- Auto-run on startup in dev only. Manual (`make migrate-up`) in staging/prod
- Timestamp format: `20240101120000_add_users_table.up.sql`

## Code Conventions

- `context.Context` is always the first parameter of any I/O function
- Never `panic` outside of `main.go` startup checks
- Interfaces defined in the package that uses them (consumer-owned)
- No global variables except logger and config (init once in `main.go`)
- Config always via envconfig struct — never `os.Getenv` inline
- `context.WithTimeout` on all DB and external HTTP calls
- Exported functions: one-line godoc. Unexported: comment only if non-obvious

## golangci-lint

Required linters in `.golangci.yml`:
`errcheck`, `govet`, `staticcheck`, `revive`, `gosec`, `exhaustive`, `gofumpt`

Pin one version in CI and in `make lint` (`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@vX.Y.Z run ./...`),
so a clean local run means a clean CI run. Never `@latest`: a new release adds checks overnight.

## Observability (Phased)

- **Phase 1** (always): logrus JSON logs, `request_id` injected via middleware, propagated via context
- **Phase 2** (when app grows): `prometheus/client_golang`, expose `/metrics`
- **Phase 3** (deployment): Loki + Prometheus + Tempo + Grafana via `docker/compose/observability.yml`
