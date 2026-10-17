# Development Workflow Practices

The standing lifecycle for any project. Four phases, each with a gate.
A project's CLAUDE.md may override specifics under a **Workflow Overrides** heading (e.g. add a
client UAT gate, a Figma sign-off, a staged rollout). State the scope at kickoff.

## 1. Design
- Confirm which components are in scope (SSOT repo + components).
- Designed work lives in `specs/NNN-slug/` (start one with `pal-harness spec new`).
- Capture the brief: problem, users, journeys, business rules → condensed in the project CLAUDE.md,
  full in `doc/PRD.md` (see `product_docs.md`).
- Record expensive, hard-to-reverse choices as ADRs (`doc/ADR/`): datastore, auth model, cloud
  target, admin stack, API versioning.
- **Contract-first** for any REST API: draft `doc/API.md` (endpoints, auth, error model) before code.
- **Gate**: brief + contract reviewed before implementation starts.

## 2. Develop
- Branch per change from `dev`: `feature/ fix/ chore/ docs/`; `hotfix/` branches from `main`. `main` always deployable; no direct commits to `main`.
- **TDD**: write the failing test from a JSON fixture first (`testing.md`), then implement.
- Follow the stack practice (`go_api` / `nextjs_web` / `react_native` / `django_admin`).
- Every PR keeps the API contract in sync: regenerate `openapi.yaml`, update `doc/API.md`, serve Swagger UI.
- **Gate**: PR with tests + lint + typecheck green.

## 3. Test
- Coverage gate **80% minimum**, enforced in CI.
- Layers: unit → integration → E2E (`testing.md`).
- Contract tests validate the API against `openapi.yaml`.
- **Gate**: all layers + coverage pass → squash-merge feature PRs into `dev`; promote `dev → staging → main` with merge commits.

## 4. Deliver
- CI/CD per surface (`devops_cicd.md`): build → scan → deploy.
- Targets: Cloud Run / ECS (api, admin), Vercel (web), EAS Build + Submit (mobile).
- Tag the release; update ADRs / changelog as needed.
- **Gate**: deployed, smoke-checked, handed to client / end user.

## Workflow Overrides

Per-project deviations live in that project's CLAUDE.md. If none, the four phases above apply as-is.
