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
- Branch per change from `dev`: `feature/ fix/ chore/ docs/`; `hotfix/` branches from `main`, PRs to `main`, then merges back into `staging` and `dev`. `main` always deployable; no direct commits to `main`.
- **TDD**: write the failing test from a JSON fixture first (`testing.md`), then implement.
- Follow the stack practice (`go_api` / `nextjs_web` / `react_native` / `django_admin`).
- Keep the API contract in sync by the split rule in `product_docs.md`: the provider's PR regenerates `openapi.yaml` and serves Swagger UI; in the same piece of work, the SSOT repo's PR updates its `doc/openapi.yaml` copy and `doc/API.md`.
- **Review**: Before opening a PR that changes UI or code, run `pal-review` (`uiux` for UI changes, `code` always), handle the findings by the severity table, and put the report in the PR body: the verdict, every major and critical finding with the owner's decision, and the minor ones with what Claude did. When the work has a spec folder, also save the full report as `specs/NNN-slug/reviews/codex-<kind>-YYYY-MM-DD-HHmm.md` in the SSOT repo.
  - `pal-review.exe` is not on PATH: call it by the absolute path in the session-start line `Codex review CLI: "<path>\pal-review.exe"` (written `pal-review` here). Run each command with a Bash timeout of 600000 ms (the 120 s default would cut it off). `run` prints `Review <id> queued` first.
  - Run from the SSOT root: `pal-review run --kind uiux` for UI changes and `pal-review run --kind code` always (`--repo <path>` takes a project-root-relative repo path such as `api`).
  - Exit codes: 0 report printed; 3 still running (run `pal-review wait <id>` with the id from `Review <id> queued`, same Bash timeout, until it ends); 4 failed; 5 skipped (put the printed reason in the PR body as the recorded skip: "Codex review: skipped (<reason>)").
  - Severity:

    | Level | Meaning | What the session does |
    |---|---|---|
    | none | Nothing worth changing | Review passes |
    | minor | Polish or a local fix; no change to a journey, a contract or data | Claude fixes it, or dismisses it with a one-line reason in the report |
    | major | Changes a journey, a shared pattern, the design system, or more than one screen or module | Claude stops and asks the owner: the finding, why it matters, and the impact |
    | critical | Broken or unsafe: data loss, security, a journey that cannot be completed, an accessibility blocker | As major, listed first |

  - Claude may argue a finding up or down a level, but a downgrade from major or critical is never silent: it is written in the report with the reason, and the owner still sees it.
- **Gate**: PR with tests + lint + typecheck green.

## 3. Test
- Coverage gate **80% minimum**, enforced in CI.
- Layers: unit → integration → E2E (`testing.md`).
- Contract tests validate the API against `openapi.yaml`.
- **Gate**: all layers + coverage pass → merge the feature PR into `dev` (it may be squash-merged); promote `dev → staging → main` with merge commits.

## 4. Deliver
- CI/CD per surface (`devops_cicd.md`): build → scan → deploy.
- Targets: Cloud Run / ECS (api, admin), Vercel (web), EAS Build + Submit (mobile).
- Tag the release; update ADRs / changelog as needed.
- **Gate**: deployed, smoke-checked, handed to client / end user.

## Workflow Overrides

Per-project deviations live in that project's CLAUDE.md. If none, the four phases above apply as-is.
