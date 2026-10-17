# DevOps / CI/CD Practices

## GitHub Actions Pipeline

Every repository must pass all stages before a PR merges (and before each promotion toward `main`):

| Stage | Go API | Next.js Web | React Native |
|---|---|---|---|
| Lint | `golangci-lint` | ESLint + Prettier check | ESLint |
| Type check | `go build ./...` | `tsc --noEmit` | `tsc --noEmit` |
| Unit tests + coverage | `make test` (80% gate) | `jest` (80% gate) | `jest` (80% gate) |
| Integration tests | `make test-int` | — | — |
| Build | `make build` | `next build` | `eas build --platform all` |
| Docker build | Yes (API only) | — (Vercel) | — (EAS) |

No stage may be skipped. Flaky tests must be fixed, not retried.

## Branch & Merge Strategy

- `main` is always deployable to production
- Feature branches from `dev`: `feature/`, `fix/`, `chore/`, `docs/` prefixes; `hotfix/` branches from `main`, PRs to `main`, then merges back into `staging` and `dev`
- Feature PR → squash-merge into `dev`; promote `dev → staging → main` with merge commits
- No direct commits to `main`
- Require passing CI + 1 review before merge

## Docker Conventions (Go API)

- Multi-stage Dockerfile: `builder` stage (Go compile) → `runner` stage (distroless or alpine)
- Final image must not contain build tools or source code
- Image tagged with Git SHA: `registry/api:${GITHUB_SHA}`
- `docker-compose.yml` for local dev: API + MySQL + Redis + migration runner
- `docker/compose/observability.yml` for local Grafana stack (optional, not required for dev)

## Environment Variables

- All secrets in GitHub Actions secrets — never in repo or workflow files
- `.env.example` committed with all required keys, no values
- `.env` is always in `.gitignore`
- Environment-specific values: `.env.development`, `.env.staging`, `.env.production`
- Never log environment variable values at startup

## Deployment Targets

**Go API**:
- GCP path: Cloud Run (primary recommendation — scales to zero, managed)
- AWS path: ECS Fargate (more control, more ops)
- Migrations run manually before deploy: `make migrate-up`

**Next.js Web + Admin**:
- Vercel (default, regardless of backend cloud)
- Environment variables set in Vercel dashboard, not committed

**React Native Mobile**:
- EAS Build for binary compilation
- EAS Submit for App Store + Play Store submission
- OTA updates via Expo Updates for JS-only changes (no store review)

**Database / Cache**:
- GCP path: Cloud SQL (MySQL) + Memorystore (Redis)
- AWS path: RDS (MySQL) + ElastiCache (Redis)
- Never expose DB or Redis ports publicly — VPC/private network only

## Cloud Cost Hygiene

- Cloud Run / ECS: set memory/CPU limits explicitly — no unbounded containers
- Staging environment scales to zero when not in use
- Container registry: tag cleanup policy — keep last 10 images per service
- Database: automated backups enabled, point-in-time recovery on

## Security in CI

- `gosec` runs in Go lint stage
- `npm audit --audit-level=high` runs in Node.js lint stage
- Docker image scanned with Trivy before push (add to pipeline when registry is set up)
- No secrets in CI logs — mask all sensitive outputs
