# DevOps / CI/CD Practices

## GitHub Actions Pipeline

Every repository must pass all stages on a PR's final commit before it merges:

| Stage | Go API | Next.js Web | React Native |
|---|---|---|---|
| Lint | `golangci-lint` | ESLint + Prettier check | ESLint |
| Type check | `go build ./...` | `tsc --noEmit` | `tsc --noEmit` |
| Unit tests + coverage | `make test` (80% gate) | `jest` (80% gate) | `jest` (80% gate) |
| Integration tests | `make test-int` | — | — |
| Build | `make build` | `next build` | `eas build --platform all` |
| Docker build | Yes (API only) | — (Vercel) | — (EAS) |

No stage may be skipped; the one exception is the tests on a promotion whose tree hash already passed (`testing.md`, Test reports). Flaky tests must be fixed, not retried.

## Branch & Merge Strategy

- `main` is always deployable to production; production deploys run only from `main`
- Work branches from `dev`: `feature/`, `fix/`, `chore/`, `docs/` prefixes; `hotfix/` branches from `main`, PRs to `main`, then merges back into `staging` and `dev`
- **Branch model before production:** a phase PR into `dev` (one the owner merges by hand may be squash-merged; a session merges its own with `gh pr merge --merge`) merges on its own once CI and the reviews pass. The owner merges the `[HUMAN GATE]` sign-off PR `dev → staging`. `staging → main` merges on its own after the owner's staging test (passed or deferred). Promotions use merge commits.
- **In production:** the project adds the owner's gate on `main` as a Workflow Override.
- A "fix now" during a staging test branches from `staging`, PRs into `staging` (staging redeploys), and is merged back into `dev`
- No direct commits to `main`
- Require passing CI before merge; the reviews are in `workflow.md`
- CI runs the tests once per PR's final commit; a promotion whose tree hash already passed skips them (`testing.md`)

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
- `npx audit-ci --high --config audit-ci.jsonc` runs in Node.js lint stage, with `audit-ci` pinned
  as a devDependency
- `audit-ci.jsonc` holds an `allowlist`. Each entry names the advisory (GHSA id), its reason
  (`notes`) and an `expiry` no more than 90 days out. An expired entry fails CI, so the reason is
  looked at again.
- A new high advisory fails CI → try, in order:
  1. `npm audit fix` without `--force`, which fixes most;
  2. if the package only runs in dev tooling (CLI, build, tests) and nothing shipped to users loads it,
     an allowlist entry with its reason;
  3. otherwise it is a real security bug: fix it, or ask the owner if the fix needs a breaking upgrade.
- Allowlisting shipped code is never allowed
- Docker image scanned with Trivy before push (add to pipeline when registry is set up)
- No secrets in CI logs — mask all sensitive outputs
