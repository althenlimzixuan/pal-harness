# DevOps / CI/CD Practices

## GitHub Actions Pipeline

Every repository must pass all stages on a PR's final commit before it merges:

| Stage | Go API | Next.js Web | React Native |
|---|---|---|---|
| Lint | `golangci-lint` | ESLint + Prettier check | ESLint |
| Type check | `go build ./...` | `tsc --noEmit` | `tsc --noEmit` |
| Unit tests + coverage | `make test` (80% gate) | `jest` (80% gate) | `jest` (80% gate) |
| Integration tests | `make test-int` | — | — |
| Generated code | `make wire` then `git diff --exit-code` (a stale `wire_gen.go` fails) | — | — |
| Build | `make build` | `next build` | `eas build --platform all` |
| Docker build | Yes (API only) | — (Vercel) | — (EAS) |

No stage may be skipped; the one exception is a job whose fingerprint already passed (the CI gate): see **The CI gate** below and `testing.md`, Test reports. Flaky tests must be fixed, not retried.

## CI runs and minutes

- **Triggers:** `pull_request`, plus `push` only on `dev`, `staging` and `main`. Never a bare `push`:
  a work branch with a PR would run twice.
- **Costly runners:** jobs on Windows or macOS runners (billed at 2x and 10x) run on `pull_request`
  only (`if: github.event_name == 'pull_request'`); the `push` runs keep the Linux jobs. Every change
  reaches `dev`, `staging` and `main` through a PR, so the PR run is the gate. A self-hosted runner costs no minutes,
  so its Windows jobs may also run on `workflow_dispatch` (the session-0 run in the phase's
  `references/build.md`, section 6).
- **Waiting on CI and merging:** watch with `gh pr checks <n> -R <owner/repo> --watch --interval 30`,
  always run as `timeout 520 gh pr checks <n> -R <owner/repo> --watch --interval 30` in the
  foreground with a Bash timeout of 540000 ms (9 minutes): the `timeout` ends it before the tool
  can, so Claude Code never moves it to the background. Exit 124 means checks are still running: run
  the same command again. Watch a workflow run started by hand the same way, with
  `gh run watch <id> -R <owner/repo> --interval 30`, run as `timeout 520 gh run watch <id> -R <owner/repo> --interval 30`. Never watch CI with `run_in_background`, Monitor or a background loop:
  Claude Code kills background watchers when memory runs low, as it does while a self-hosted runner
  builds on the same PC.
  0. **Hold pushes while CI runs.** The CI gate cancels the older run when a newer push reaches the
     same PR, so a state, log or docs push to a PR's branch while its CI runs costs a whole run.
     Commit those to the branch locally and push them in one push once the watch ends. This is the
     one exception to the phase skill's **Saved state** step 3; it is safe because a CI run is far
     shorter than the 1-hour stuck limit. Pushes to other branches go on as usual.
  1. **Wait** with that one foreground watch.
  2. **A watch that ends without its own exit code** (the tool says it was killed, interrupted or
     timed out) is run again. From the second such end in a row, run `pal-harness mem --wait 9m`
     (Bash timeout 600000 ms) first. It is never an owner stop: no `pal-work hold`, no `waiting`
     status, no limit on retries.
  3. **Merge guard.** Before every `gh pr merge`, run `gh pr checks <n> -R <owner/repo>` on its own,
     with no pipe (a pipe returns the last command's exit code and hides this one), and read its exit
     code: `0` every check passed or was skipped → merge; `8` still running → back to step 1; `1` a
     check failed → read the run (**Read a red run before fixing it**). The spending-limit case below
     keeps its local-evidence rule.
  4. **`ci:` values.** `phase-state.yaml` `ci:` is only `green`, `red`, `running` or `none`
     (`pal-harness check` rejects anything else). Exit 8 is written `running`; never `pending`.
  5. **One runner, one queue.** When a repo's heavy jobs run on the self-hosted runner
     (`gh variable get WINDOWS_RUNNER -R <owner/repo>` prints a label), its runs wait one after
     another, about 20 minutes each. So:
     1. before a push that starts CI (a push to a branch with an open PR), count the runs ahead in
        each repo the phase touches: `gh run list -R <owner/repo> --status queued` and
        `--status in_progress`, and the online runners with the label
        (`gh api repos/<owner/repo>/actions/runners`). Write, with `pal-harness log` (it stamps
        `last_progress_at`: a wait in the queue is not stuck), `CI queue: N runs ahead (about 20 min
        each)` with one runner, or `CI queue: N runs ahead, R runners (about M min)` with R of them,
        M = 20 × N / R;
     2. a branch with an open PR gets one push per round: all of the round's fixes, in every repo,
        committed locally and pushed together (step 0 already holds state and log pushes). Pushes
        before a PR opens start no CI (the `pull_request` trigger) and go on as usual.
- **The CI gate:** every repo with a code CI starts its `.github/workflows/ci.yml` from
  `.pal-harness/templates/ci.yml`. Its `gate` job calls the harness's reusable `ci-gate.yml`, which
  first runs `pal-harness utf8` (a tracked text file that is not UTF-8 makes the gate, and so the
  run, red; save the file as UTF-8), then runs `pal-harness fingerprint` over the paths each job reads; a job whose fingerprint already
  passed (recorded by `ci-mark.yml` as `refs/ci-passed/<job>/<fingerprint>`) is skipped. So
  docs-only and `specs/**` changes, the push after a merge and promotions do not run the suite
  again. The concurrency group (`cancel-in-progress: true`) cancels the older run when a newer push
  reaches the same PR or branch.
  - Each job's line in `gate.with.jobs` lists everything the job reads: code, build files, lint
    config, a contract copy its drift check compares. A job that starts reading a new path adds it
    in the same PR. `.github/workflows/ci.yml` is always included, so a change to it (a re-pin
    included) runs every job.
  - A gate that failed or never started (no GitHub minutes left) runs every job.
  - Locally, `pal-harness fingerprint --with .github/workflows/ci.yml <spec>` shows whether a change
    keeps a job's fingerprint.
  - The SSOT repo's `ssot` check runs on every push, `specs/**` included: it validates
    `phase-state.yaml` and runs on GitHub's machines, not the owner's PC.
- **Free memory before heavy runs:** before the test reports, a full test run (the full check before
  the PR, a fix round's check), `make lint`, `make mock`, a release or native build (Tauri, local Gradle or EAS, Docker) or an
  app install on a phone (`pal-harness` is run as `.pal-harness/practices/ssot_standard.md` says).
  Every heavy command (mock, lint, test, build) runs in the foreground with a Bash timeout of
  600000 ms, never `run_in_background`:
  1. run `pal-harness mem` (exit 0: at least 2 GiB available);
  2. under the minimum (exit 1): stop what the session started and no longer needs (an emulator, an
     old dev server, Metro from another checkout), then run `pal-harness mem --wait 9m` in the
     foreground (Bash timeout 600000 ms); it re-checks every 30 seconds;
  3. still under: run anyway and write `low memory: <free> free before <run>` in the log. It is not
     an owner stop. Exit 2 (a platform it cannot read) is noted in the log, and the run goes on.
  4. **A run cut short by low memory** (killed, or no exit code of its own) is run again, never pushed over:
     lint is run again; `make test` is re-run one package group at a time with `GOFLAGS=-p=1`,
     retrying only the group that was cut; a stopped `make mock` falls back to
     one mock per changed interface (`--name <Interface>` with `--all=false` and `--dir <its package folder>`),
     as `go_api.md` (**Mocks.**) says.
  5. **Jest exit `0xC0000142` on Windows** (exit code 3221225794 or -1073741502: a process that
     could not start) counts as a run cut short by low memory: run `pal-harness mem --wait 9m`, then
     re-run only the failed suites (`npx jest <their paths> --maxWorkers=2`). Log
     `process-start failure: re-ran N suites`; it is not a fix round. The same suites failing the
     same way again are re-run with `--maxWorkers=1`; a third time is a real failure.
- **Read a red run before fixing it** (`gh run view <id> -R <owner/repo>`). A job that never started
  because of the spending limit or a billing problem is not a code failure and not a fix round:
  1. run the repo's CI commands locally on the PR's head (the checks of `workflow.md`, **Check before
     every push**) and save the output as `specs/NNN-slug/reviews/local-ci-YYYY-MM-DD-HHmm.md`;
  2. send one "needs you: GitHub Actions spending limit reached on <repo> — CI ran locally instead"
     notification and keep working (it is not an owner stop);
  3. keep `ci: red` with the note `ci blocked: Actions spending limit; local CI green`. A PR into
     `dev` may merge on that local evidence; the sign-off PR says CI did not run.

  **Outage re-run.** A job that failed because a package mirror or the network was down is not a
  code failure either. Its failing step's log shows an outage sign: apt `Failed to fetch` or
  `Hash Sum mismatch`, `Could not resolve host`, `Connection timed out` or `Connection reset`,
  `ETIMEDOUT`, `ECONNRESET`, or an HTTP 5xx from a package registry, mirror or module proxy. Then:
  1. run `gh run rerun <run-id> -R <owner/repo> --failed` once and watch the run again (**Waiting on
     CI and merging**). Push nothing. Write `infra re-run: <job> (<sign>)` in the log; it is not a fix
     round and does not count toward the stuck limit;
  2. the re-run fails with an outage sign again: tell the owner in one line, then run the CI locally
     as steps 1 to 3 above say (note `ci blocked: <sign> outage; local CI green`). A green local run
     counts as the PR's CI for the merge gates.

  This holds for every CI watch: build, fix now, quick fix, sign-off and ship. A failure with no
  outage sign is read and fixed as usual.
- **Self-hosted runners:** a repo that runs out of free minutes moves its heavy jobs (Windows,
  long builds) to the owner's self-hosted runner. The job reads the runner from a repository
  variable, `runs-on: ${{ vars.WINDOWS_RUNNER || 'windows-latest' }}` (`templates/ci.yml` shows it),
  so setting `WINDOWS_RUNNER` to the runner's label moves it, and unsetting it moves it back.
  - **Two jobs at once (owner-run, once per repo).** One runner takes one job at a time, so a run
    with several Windows jobs waits on itself (about 35 minutes in Pal Center phase 023). A second
    runner on the same PC, in its own folder, named `pal-pc-2` with the same custom label, takes the
    next job beside the first; `WINDOWS_RUNNER` is unchanged. It installs a Windows service, so it is
    a handover, run by the owner in an admin PowerShell:

    ```powershell
    New-Item -ItemType Directory C:\actions-runner-2; Set-Location C:\actions-runner-2
    Expand-Archive C:\actions-runner\actions-runner-win-x64-<version>.zip .   # the zip already there
    $t = gh api -X POST repos/<owner/repo>/actions/runners/registration-token --jq .token
    .\config.cmd --unattended --url https://github.com/<owner/repo> --token $t --name pal-pc-2 --labels pal-pc --work _work --runasservice
    ```

    `gh api repos/<owner/repo>/actions/runners` then lists both online. Two jobs share the PC's
    memory: a run that fails on low memory while both runners are busy is re-run once; a second such
    failure means the owner stops the `pal-pc-2` service, and the queue goes back to one runner.
- **Bot PRs get CI:** a workflow that opens a PR with `GITHUB_TOKEN` gets no CI, since GitHub starts
  no workflow for it. So `ci.yml` has `workflow_dispatch:`, and the bot workflow (with
  `actions: write`) runs `gh workflow run ci.yml --ref <branch>` after opening the PR;
  `workflow_dispatch` is the one event `GITHUB_TOKEN` may start. A dispatched run skips the
  `pull_request`-only Windows and macOS jobs; a session that finds such a PR closes and reopens it
  (`build.md`, section 2), which runs them all.

## Branch & Merge Strategy

- `main` is always deployable to production; production deploys run only from `main`
- Work branches from `dev`: `feature/`, `fix/`, `chore/`, `docs/` prefixes; `hotfix/` branches from `main`, PRs to `main`, then merges back into `staging` and `dev`
- **Branch model before production:** a phase PR into `dev` (one the owner merges by hand may be squash-merged; a session merges its own with `gh pr merge --merge`) merges on its own once CI and the reviews pass. The owner merges the `[HUMAN GATE]` sign-off PR `dev → staging`. `staging → main` merges on its own after the owner's staging test (passed or deferred). Promotions use merge commits.
- **In production:** the project adds the owner's gate on `main` as a Workflow Override.
- A "fix now" during a staging test branches from `staging`, PRs into `staging` (staging redeploys), and is merged back into `dev`
- No direct commits to `main`
- Require passing CI before merge; the reviews are in `workflow.md`
- CI runs each job once per set of files it reads; a job whose fingerprint already passed is skipped (**The CI gate**, `testing.md`)

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
- **Cloud Run:** after `terraform apply`, traffic stays on the old revision until the deploy
  workflow runs, so the runbook runs the deploy workflow next and checks the serving revision
- **gcloud and Terraform:** every `gcloud` command, the session's own and the handover's, passes
  `--project <id>`: the PC's default project may belong to another product, so it is never
  trusted. An owner-run Terraform handover lists, in the folder and in order,
  `gcloud auth application-default login` (once per PC, or when it has expired),
  `terraform init -backend-config=backend.hcl` when the folder has no `.terraform/` yet, then the
  `plan` and `apply` with the exact `-var-file`, and names the GCP project ID

**Next.js Web + Admin**:
- Vercel (default, regardless of backend cloud)
- Environment variables set in Vercel dashboard, not committed
- **Vercel:** a project is set up with its root folder and production branch before the first
  deploy (a monorepo's root folder is the app's folder, the production branch is `main`). A
  production redeploy or a settings change is tried on a preview first.

**React Native Mobile**:
- EAS Build for binary compilation
- EAS Submit for App Store + Play Store submission
- OTA updates via Expo Updates for JS-only changes (no store review)
- Under Company mode a store submission or OTA update is held and goes out as one release (phase
  skill, `references/ship.md`, section 8)

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
  (`notes`) and an `expiry` no more than 90 days out;
  an expired entry stops suppressing, so CI fails while the advisory is still present and the reason
  is looked at again. A record needs `"active": true`:
  `"allowlist": [{ "GHSA-xxxx-xxxx-xxxx": { "active": true, "notes": "Expo CLI only; not shipped", "expiry": "YYYY-MM-DD" } }]`
- A new high advisory fails CI → try, in order:
  1. `npm audit fix` without `--force`, which fixes most;
  2. if the package only runs in dev tooling (CLI, build, tests) and nothing shipped to users loads it,
     an allowlist entry with its reason;
  3. otherwise it is a real security bug: fix it, or ask the owner if the fix needs a breaking upgrade.
- Allowlisting shipped code is never allowed
- Docker image scanned with Trivy before push (add to pipeline when registry is set up)
- No secrets in CI logs — mask all sensitive outputs
