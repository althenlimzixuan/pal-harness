# Changelog

## 1.10.0 — 2026-11-24

Ten harness notes raised in Pal Center (ai-workshop#58 to #67), from Pal Center phase 022's and 023's retros. The design is `doc/specs/2026-11-24-harness-1.10.0-design.md`. No state field changes, so the installed Pal Center reads 1.10.0 phase state as it reads 1.9.0. The planned `upgrade` command and automatic re-pin move to 1.11.0.

- **New `pal-harness minors <ledger> [--tag <tag>] [--task <N>]`** (#60). It lists sdd's deferred minors grouped by the tag set at triage (`Task K: minor (deferred) [carry]: …`):
  - `carry`, `final`, `later`, `issue`, then sdd's `parked` lines, then `untagged`, then a count of `fixed`;
  - exit 1 when a minor has no tag or an unknown one.

  The controller tags at triage (`subagents.md` section 2 step 5) and builds each Carry-over list with `--task K --tag carry`. `build.md` **Deferred minors** and `pal-review-code` check 8 read the command's list instead of the ledger.
- **New `pal-harness merge-check [--base origin/dev] [dir]`** (#63). It fetches, then tries the merge of `HEAD` in a throwaway worktree that is always removed; the caller's tree is never touched. It works on git older than 2.38. It runs at the start of the review stage (`review.md` **Merge check first**) and before each PR into `dev` (build section 6 step 3, quick fix). A conflict is merged in before CI, keeping both sides of a bookkeeping file.
- **`spec new` branches from `origin/dev`** (#67). With an `origin` remote it fetches `dev` and creates the branch from `origin/dev` with `--no-track`. It stops when the fetch fails, origin has no `dev`, or local `dev` has commits not on `origin/dev`. With no remote it uses local `dev`, as before.
- **QA seeds** (#58): the spec's UX outline ends with **QA seeds**, the records each changed screen needs to look filled. The plan adds missing seeds to the project's QA sample data, the plan review checks them (check 13), and the QA pass screenshots each changed screen in its seeded state.
- **RED by name and guard tests** (#59):
  - the controller matches each new test by name to its RED line before the task reviewer is dispatched;
  - a RED recovered afterwards by a mutation run is logged `RED by mutation: <test>` and is not a fix round;
  - the plan's optional **Guard tests:** pass on their first run, and their RED is a mutation run.
- **Mutation check** (#66): the task reviewer breaks the covered code once per new test (at most 5 per task), confirms the test fails, restores with `git restore -- <files>`, and checks `git status`. A test that stays green is an Important finding.
- **Plan review** (#61, #64):
  - check 19 **Commands and paths** looks up every path, make target, npm script and generator flag a task names;
  - check 20 **Other operating systems** needs the other OSes' checks in **Prove it**: `GOOS=linux` and `GOOS=darwin go vet`, `cargo clippy --target <triple>`, or the Docker run and `macOS proof: CI's macOS job`.

  `linux-check.md` **Cheap first** gains the `cargo clippy --target` and macOS lines.
- **A fix stays inside its finding** (#62): a fixer changes owner-facing behaviour only as far as its finding says, and reports anything more under **New question**. That question joins a coming findings stop, or becomes a backlog issue (`Why: a fix needs an owner-facing change`) listed in the sign-off PR.
- **Done work drops out** (#65): `/pm` **Bundle and vehicle** step 1 and the spec stage check each backlog issue against `doc/features.yaml`, merged PRs and closed issues, and close the done ones with `Done in <link>`. Pal Center's own Inbox triage gets the same check through F-036.

## 1.9.0 — 2026-11-22

Ten harness notes raised in Pal Center (ai-workshop#48 to #57), mostly from Pal Center phase 020's retro. The design is `doc/specs/2026-11-22-harness-1.9.0-design.md`. The planned `upgrade` command and automatic re-pin move to 1.10.0.

**Release order.** 1.9.0 adds the `owner_stops` field to `phase-state.yaml`, and Pal Center reads phase state with the harness version in its own `daemon/go.mod`. Move Pal Center's dependency to 1.9.0 and ship it **before** any project re-pins to 1.9.0. Until then, the reader check keeps sessions from writing `owner_stops`.

- **Trial measure** (#48): optional `owner_stops:` in `phase-state.yaml`, one entry per answered owner stop: `{kind, since, answered, could_decide, via}`, where `via` is `owner` or the `wait-decision` result. `check` validates it. The phase skill's **Owner stops** step 5 and `company-mode.md` section 4 write it, and the retro's **Owner stops** table reads it, falling back to the log for older phases.
- **Pal Center and the pin agree** (#55, #56):
  - the phase-state template leaves out optional keys whose default applies (`qa`, `owner_stops` are comments), and **Saved state** writes them only with a non-default value;
  - new `pal-harness check --reader <version>` prints a `warning:` line for each field a state file uses that is newer than that reader (`state.FieldSince`: `qa` and `results.check` 1.8.0, `owner_stops` 1.9.0). Warnings don't change the exit code;
  - **Reader check** at phase start, resume and ship: read the harness version in Pal Center's `daemon/go.mod` at tag `v<Pal Center version>` and run `check --reader` with it. On a warning, the session leaves the newer keys out and tells the owner to update Pal Center.
- **Plans** (#49): a plan over 8 tasks or about 30 KB is written in parts (`plan.md`, `plan-2.md` …), each reviewed on its own (`pal-review-plan-partN-…`). Round 1 of the plan review is the whole review; a later round checks only the fixes, and a new finding on unchanged text is raised only when critical.
- **QA warm-up** (#50): the QA pass loads the first page once before the first step, and a first-load timeout is retried once (`cold start: retried`). `templates/screens.spec.ts` warms up in `beforeAll`.
- **Subagents** (#51, #54):
  - implementers finish or stop every background command, Monitor and watcher before the final report;
  - a report from a finished task is ignored, and a second gets TaskStop;
  - should-fix minors from task K's review go on task K+1's **Carry-over** list and are fixed first (`Fix Task K review minors`). The last task's go to one fixer before the whole-branch review.
- **CI outages** (#52): `devops_cicd.md` **Outage re-run**. A job that failed with a mirror or network sign (`Failed to fetch`, `Hash Sum mismatch`, `Could not resolve host`, connection reset or timed out, a registry 5xx) is re-run once with `gh run rerun <run-id> --failed` and logged as `infra re-run`, not a fix round. If it fails the same way again, a green local CI run counts.
- **Linux checks on Windows** (#53): new `skills/phase/references/linux-check.md` covers:
  - `GOOS=linux` lint and vet first;
  - WSL with git 2.28 or later (the git-core PPA as an owner-run command);
  - TMPDIR on `/dev/shm`;
  - Docker `rust:1-bookworm` for Tauri;
  - `pal-harness mem` minimums and `.wslconfig`;
  - `Linux proof: CI's Linux job` when none of these is available.

  It is linked from `build.md` and the quick-fix skill.
- **Every ending tested** (#57): the task reviewer's checks and `pal-review-tests` check 10 list every way an outcome feature's run can end (start failed, owner stop, restart mid-run, idle, waiting, cancelled, timed out, crashed or killed, plus the feature's own). A missing test for any of them is a major finding.

## 1.8.0 — 2026-11-19

The 61 harness notes left in Pal Center's inbox after the 2026-11-19 clean-up (Pal Center phases 005 to 019, LFPV phases 006 to 011), released as one version. The design is `doc/specs/2026-11-19-harness-1.8.0-design.md`, and its notes table maps each note to its change. Note #64 (trial measures) stays open. The planned `upgrade` command moves to 1.9.0.

- **Company mode** (new `skills/phase/references/company-mode.md`):
  - every `/pm`, pickup, phase and quick fix runs `pal-work mode --project <key>` first;
  - at a spec-review or findings stop with Company mode on, the session runs `pal-work wait-decision --phase NNN` and acts on approved, defaulted, changed, vetoed or none;
  - anything about money, a key or anything public is held with `pal-work hold`;
  - an answer that came from the default is recorded `(used by default)`;
  - the sign-off PR and the staging test stay the owner's;
  - `/pm` writes the Bundle line in the same write as the move.
- **Chores and pickup** (new `skills/pm/references/chores.md`):
  - the pickup ladder gains a **Chores** rung for due cards in `Chores (Active)` (`Every:` / `When:` / `Last run:`);
  - `Action: harness notes` cards stay Pal Center's;
  - a chore card is created only after the owner's yes;
  - the pickup report gives the rung and reason for each skipped item, and the `/pm` note lists "Needs Approved to run".
- **QA pass** (new `skills/phase/references/qa.md`):
  - optional `harness.qa_command` in `project.yaml`, and `qa: none|running|passed|failed` in `phase-state.yaml` (missing reads as `none`);
  - it runs after the sign-off merge, before the staging test (or before ship under `no-owner-test`);
  - results go in `specs/NNN-slug/qa/`, and a failure is a fix round.
- **New flow override `uat-queue`:**
  - an unanswered staging test parks after 20 minutes;
  - up to 3 phases wait on `staging` for one combined test;
  - the oldest phase's session promotes them together, and the others skip promotion;
  - a failure that later phases build on holds the batch, and steps that publish are never deferred.
- **CI waits.** `devops_cicd.md` **Waiting on CI and merging**:
  - **Hold pushes while CI runs** (the one exception to **Saved state** step 3);
  - every watch runs as `timeout 520 gh pr checks … --watch --interval 30` (exit 124: run it again).
- **Memory.** **Free memory before heavy runs** covers `make lint` and `make mock`, and every heavy command runs in the foreground with 600000 ms. A run cut short is run again, never pushed over:
  - `make test` is retried per package group with `GOFLAGS=-p=1`;
  - mocks fall back to `mockery --name <Interface>` with one full `make mock` before the PR;
  - Jest uses `--maxWorkers=2`;
  - startup waits read one CI-scaled constant.
- **Subagents** (new `skills/phase/references/subagents.md`; `build.md`'s **Larger phase** steps 1, 3 and 4 move there):
  - TaskStop on a second repeated report;
  - a reply must quote the latest fix message's first line;
  - no waiting on own background work, and a timeout is a failure;
  - a **Known findings** list in every brief;
  - an owner-facing wording check in the task review;
  - stale language-server diagnostics are ignored when build, vet and tests pass.
- **`workflow.md`:**
  - **Plain English** for everything the owner reads;
  - **Session habits**: the branch check before and after a push, retrying a temporary auto-mode check failure, `--body-file` for PRs, `npm ci` when `node_modules` is missing or stale, and restating a split plan before acting on it.
- **Plan review:**
  - a minors-only round ends the review;
  - check 17 **File lists** and check 18 **Mobile release** (an app version bump; the resolved versions of a native dependency's shared libraries);
  - link and junction test cases for delete helpers;
  - an **Owner-run commands** section in the plan.
- **The spec:**
  - a new feature asked for during review becomes its own card;
  - a running **Rulings** list, with one Codex doc check per batch;
  - `unverified: <reason>` steps, which the sign-off PR lists;
  - a store identity check;
  - destructive or optional steps in their own feature;
  - the two-role seed.
- **Reviews:**
  - `review-code`: client timeout against the server's chained budget; concurrency and lifecycle, applied early; kill switch;
  - `review-errors`: programming errors not swallowed; failure paths of multi-step operations;
  - after each fix round the failure path and the retry are re-checked;
  - error-path fixtures are copied from a real response.
- **Tests:**
  - the speed check takes 10 pairs, and a slowdown needs both >20% and >2 ms on the p95;
  - Jest shims export only the native module's exports, with a parity test;
  - `screens.spec.ts` grows inner scroll containers before capturing.
- **New `pal-harness utf8 [dir]`:** every tracked text file must be UTF-8 (`<path>:<line>: not UTF-8`, exit 1). The reusable `ci-gate.yml` runs it before fingerprinting, so every repo on the CI gate gets it.
- **`features.yaml` results** take `check: <name>` in place of `step:` for an extra check (exactly one of the two). An unknown field in any `features.yaml` or `phase-state.yaml` type names the allowed fields.
- **Staging test:**
  - the two-role seed;
  - a pre-flight that checks Docker (`docker info`) and the adb server, and resets the app with `adb shell pm clear <package>` (allowed in the project's `.claude/settings.json`, or a debug-only reset);
  - steps that name exactly what to stop or start;
  - the QA summary link.
- **Release:**
  - new `templates/runbook-android-release.md` (EAS keystore, the upload key lock, the Play Console draft);
  - `ship.md` section 0 does a runbook dry-read and hands over the Owner-run commands;
  - `devops_cicd.md` rules for **Vercel**, **Cloud Run** and **Self-hosted runners** (`WINDOWS_RUNNER`, shown in `templates/ci.yml`);
  - `go_api.md` logs a hashed or shortened client address in production.
- **Retro:** an **Owner stops** section (kind, wait, and whether it could have been a decision under Company mode) with a one-line count.

## 1.7.0 — 2026-11-19

Pal Center's phase 003 and 005 and quick fix 002 notes, and LFPV's mock note (althenlimzixuan/ai-workshop#39–#47, althenlimzixuan/lfpv-docs#20). The design is `doc/specs/2026-11-19-harness-1.7.0-design.md`. The planned `upgrade` command moves to 1.8.0.

- New CI gate: reusable workflows `.github/workflows/ci-gate.yml` (fingerprints each job's paths and skips a job whose fingerprint already passed) and `ci-mark.yml` (records a passed job as `refs/ci-passed/<job>/<fingerprint>`), a new `pal-harness fingerprint [--rev HEAD] [--with <path>]... [--dir .] <spec | ->` command, and `templates/ci.yml`, the caller with a concurrency group that cancels the older run. `devops_cicd.md` gains **The CI gate**; the pipeline's one skip and `testing.md`'s Test reports move from "promotion tree hash" to "job fingerprint". Projects adopt it by one quick fix each.
- `devops_cicd.md`: **Watching CI** becomes **Waiting on CI and merging**: a watch killed for low memory is run again (with `pal-harness mem --wait 9m` from the second kill), never parked for the owner; before every `gh pr merge`, the unpiped `gh pr checks` exit code (`8` still running: do not merge); `ci:` is only `green`, `red`, `running` or `none`. Every merge in the phase, quick-fix and pm skills points there.
- `devops_cicd.md` and `go_api.md`: a **Generated code** stage, `make wire` then `git diff --exit-code`, so a stale `wire_gen.go` fails CI.
- `go_api.md` and `testing.md`: mocks are `go run github.com/vektra/mockery/v2@v2.53.7 --all --output mocks` into a flat top-level `mocks/` (package `mocks`), as LFPV runs it; `--keeptree` is gone; interface names are unique across the module.
- `plan.md`, the plan template and `review-plan`: a **Deletes and locks** section (each delete helper's read-only check that its target is its own; each callback fired under a lock, with the lock); check 16 and two `harness list` items make a missing one a major finding.
- `pal-work wait-ci` and the CI nudge (ai-workshop#41, #42) went to Pal Center as a product capture; the RED-per-test rule (ai-workshop#44) already shipped in 1.6.0.

## 1.6.0 — 2026-11-09

Pal Center's phase 009 notes and LFPV's phase 003 notes (althenlimzixuan/ai-workshop#29–#38). The design is `doc/specs/2026-11-09-harness-1.6.0-design.md`. The planned `upgrade` command moves to 1.7.0.

- New `pal-harness mem [--min 2GiB] [--wait 0] [--every 30s]`: the available physical memory against a minimum, re-checked while short for up to `--wait` (Windows and Linux). `devops_cicd.md` gains **Free memory before heavy runs**: before reports, full test runs, native builds and phone installs, run it; under 2 GiB stop the session's leftovers and wait up to 9 minutes, then run anyway with a `low memory` log line. `build.md`, `review.md`, `staging-test.md` and quick-fix point there.
- New `pal-harness adb [dir]`: reconnects each phone in `doc/devices.yaml` (a `second-user` role for a second test account's phone) at its saved address or the one mDNS finds, sets the reverse ports, and prints the pairing request for a phone still missing; it never disconnects or restarts adb. `react_native.md` describes the file and the command, and says Metro's file watching holds in every fix round; the staging-test pre-flight runs it.
- New template `templates/screens.spec.ts`: a Playwright test that saves every listed page in light and dark at the default and narrowest width. `testing.md` (**Screenshots**), `review.md` (run it before the reviewers and on the final commit when screens changed), `review-ui` (reads the PNGs), `signoff.md` and `signoff-pr.md` (a **Screens** section with the images).
- `spec.md` and the spec template: a **Data sources** table naming the endpoint and field behind each figure a spec says already exists; `review-plan` flags a figure in neither the table nor a task (major).
- `plan.md` and the plan template: a **Gate dry-runs** table (each grep or lint gate run on the current tree and on a planted bad sample before the plan review) and an **Existing behaviour** list with a test per item for restyle tasks; `review-plan` checks both and repeats each grep.
- `build.md`: the implementer reports each new test's RED output before its GREEN; the dispatch to sdd's task reviewer rejects a task without it, and adds the amber (warning colour only where the user must act), window-edge and existing-behaviour checks; **Deferred minors** from sdd's ledger go to `doc/later.md` (or a backlog issue) before the PRs, and `review-code` checks none is missing.
- `review.md`, `signoff.md`, quick-fix: a Codex skip because auto review is off says "turn it on in Pal Center at Projects → <project> → Reviews".

## 1.5.1 — 2026-11-09

Pal Center's phase 010 and 011 notes (althenlimzixuan/ai-workshop#19–#28). Rule text only. The design is `doc/specs/2026-11-09-harness-1.5.1-design.md`.

- `testing.md`: new **Windows security tests**: DACLs compared ACE by ACE (SID, access mask, flags), never by `String()`, SDDL or account name; every precondition set up and asserted in the test; a session-0 gap skips with its reason, listed in `summary.md`. Local Go test runs set `GOFLAGS=-p=1 GOMEMLIMIT=2GiB`; CI never does.
- `devops_cicd.md`: new **Watching CI**: `gh pr checks --watch --interval 30` (or `gh run watch`) in the foreground with a 9-minute timeout, repeated; never a background watcher. A self-hosted runner's Windows jobs may also run on `workflow_dispatch`. Every `--watch` line in the skills points there.
- `build.md`: the full check and the **Session-0 run** before the PR (the project's `Session-0 tests:` command, else `gh workflow run` on a self-hosted runner, else a recorded "not available"); **An owner request during the build** (one task added under **Added during build** with a decision, no plan re-review; bigger becomes a `todo` feature); per-task fix rounds run the repo-wide lint and typecheck.
- `review.md`, `staging-test.md`, quick-fix: a fix round is done only after the repo's full lint, typecheck and test commands; `pal-review-security` also runs when the spec has a Trust boundaries section.
- `spec.md` and the spec template: a **Trust boundaries** section when code runs as another account or elevated, or reads what another account can write; security claims name what enforces them.
- `plan.md`, `review-plan`: round 1 starts from the earlier majors (the last 5 phases' plan reviews and owner batches) and a fixed harness list, reported in an **Earlier majors** table; new checks for Trust boundaries, absolute security wording and the session-0 run.
- `review-security`: checks the code against Trust boundaries and flags absolute wording in guides, UI text and comments.
- `SKILL.md`: a Pal Center CLI blocked by Defender is treated as exit 5, with one message giving the exact-file exclusion command.

## 1.5.0 — 2026-11-02

The 21 open harness captures (Pal Center phase 001, LFPV phases 003 and 004, ETP phase 097). The design is `doc/specs/2026-11-02-harness-1.5.0-design.md`.

- New `pal-harness log <phase folder> <headline> [body file]`: appends a clock-stamped `phase-log.md` entry and sets `last_progress_at`; refuses a folder outside `.pal-worktrees/`. The phase and quick-fix skills write every log entry with it.
- New `pal-harness contrast <foreground> <background>`: the WCAG 2 ratio and AA verdicts for `#rgb`, `#rrggbb`, `rgb()` and `oklch()` (clipped to sRGB with a note; alpha refused). `review-ui` and contrast fixes measure with it.
- `devops_cicd.md`: new **CI runs and minutes** — `push` only on `dev`, `staging`, `main`; Windows and macOS jobs on `pull_request` only; a spending-limit block is run locally as evidence (`reviews/local-ci-*.md`), one notification, no fix round; bot PRs opened with `GITHUB_TOKEN` start CI by `workflow_dispatch`.
- `build.md`: bot PRs into `dev` get CI (close and reopen) and merge before a phase branches; sdd's scripts run in a subshell started in the set (`review-package` in the component worktree); implementer and fixer reports need a red run, or the task goes back; lint runs in each task.
- `SKILL.md`: a writing script's target is an absolute path inside the set, checked before it runs; a script that resolves its target from the current folder runs in a subshell started in the set.
- `review.md`, `signoff.md`, quick-fix, `signoff-pr.md`: a PR whose CI was blocked by the spending limit merges on its green local run, and the sign-off PR says CI did not run.
- `review-plan` and `plan.md`: every contract endpoint and appendix behaviour has a task; a key, secret, token or invitation handoff says who may receive it and how the sender verifies them (major finding otherwise); each test step can run where the owner tests; each task's prove-it includes lint.
- `spec.md`: each test step is checked against the local stack's or staging's limits and the real phone's.
- All seven reviewer agents gain the Write tool for their one report file; the session checks `git status` after each.
- `testing.md`, `review.md`, `review-tests`, `test-summary.md`: the speed check is paired, alternating old and new runs on one machine (at least 5 pairs, medians compared); the report command starts with a Windows-aware preflight and runs Jest reports one at a time.
- `workflow.md`, `react_native.md`: lint and format through the repo's own scripts, never a bare `npx prettier` or `npx eslint`.
- `staging-test.md`: new section 3 **Pre-flight** (health and accounts, a cold Metro start and a full reload, `adb devices` and `adb reverse`, a pairing code up front, free memory), run again after a gap; later sections renumber to 4–7.
- `review.md`, `staging-test.md`: a fix that changes owner-visible behaviour updates the affected test steps in the spec and `doc/features.yaml`.

## 1.4.2 — 2026-11-01

- `render`: the page's `body` no longer sets `max-width` (was `42rem`); pages use the full width, keeping their side padding (owner rule, 2026-11-01).

## 1.4.1 — 2026-10-25

LFPV phase 002 retro fixes.

- `devops_cicd.md`: the Node audit is `npx audit-ci --high --config audit-ci.jsonc`; dev-tooling advisories get an allowlist entry with a reason and an expiry (at most 90 days), shipped code never.
- `testing.md`, `review.md`: one report command (`make reports` or `npm run reports`) runs once, after the last review fix; a report with demo data is stale; quick-fix makes its reports after its fix rounds; `pal-review-tests` skips the report checks until the reports are made.
- `react_native.md`: new **Native builds on Windows** — local Android builds are arm64-only at Gradle `--priority=low`; all-ABI builds are EAS-only.
- `react_native.md`: new **Device testing** — Metro with file watching; never `adb disconnect`, `adb kill-server` or a wireless-debugging toggle; one reconnect, then ask the owner.
- `go_api.md`, `testing.md`: mocks with `mockery --all --keeptree --dir internal --output internal/mocks` (`make mock`, same settings in `.mockery.yaml`).
- `spec.md` reference and template: a surprising but correct step says why, "(by design: reason)".
- `staging-test.md`: the ready-for-test message lists each feature's steps inline; the notification stays short.

## 1.4.0 — 2026-10-24

Single-repo projects, four flow overrides, and no plan stop.

- `project.yaml` gains `harness.flow_overrides` with `auto-promote`, `no-owner-test`, `installer-ready` and `ci-test-summary`; an unknown or repeated name fails `check`, and `auto-promote` prints a warning.
- A project is single-repo when every component's repo is the SSOT repo; the phase, quick-fix and pm skills keep spec, plan, state and code on one `docs/NNN-slug` branch.
- New feature status `shipped` (only with `no-owner-test`).
- No plan stop in any project: the plan is still written, reviewed and linked from the sign-off PR; "include load test" moves to the spec reply.
- Every spec has an Infrastructure section built on `doc/infra.mmd` (new template), with names and kinds only.
- Reviewers check owner guides (`review-code`), the Infrastructure section (`review-contract`) and, under `ci-test-summary`, the CI test summary (`review-tests`).
- New Go API: `(*manifest.Project).HasFlowOverride`, `(*manifest.Project).SingleRepo` and `state.ValidateFeaturesFor`; `check` skips the `shipped` rule when `project.yaml` is invalid.
- `review-contract` also runs when the infrastructure changes, not only the API.
- In a single-repo project the final commit is the last commit that changes code; report, review and state commits after it do not make a report stale.
- Under `installer-ready` the version is bumped in the phase's PR (the plan's last task) or in the quick fix's commit; ship tags it and downloads the installer.

## 1.3.2 — 2026-10-21

A project that adopts the phase flow late can leave its old specs alone.

- Added the optional `harness.legacy_specs_before` field to `project.yaml` (three digits). `check` skips `phase-state.yaml` in `specs/NNN-*` below it; later specs are validated as before.
- `workflow.md`: check before every push and PR — the repo's own lint and tests at CI's pinned versions; never push unchecked. `go_api.md`: pin one golangci-lint version in CI and `make lint`, never `@latest`.

## 1.3.1 — 2026-10-20

The phase flow uses Pal Center 0.4.0.

- Phase, quick-fix and pm skills use the `Work CLI:`, `Capture CLI:` and `Codex review CLI:` session-start lines (`pal-work`, `pal-capture`, `pal-review`), and `wait-results` through Monitor for the staging test.
- Spec review runs `pal-review run --kind doc --spec specs/NNN-slug`; worktree-set reviews run with Pal Center 0.4.0 or later (`Pal Center version:` line).
- Sessions Pal Center starts (`<KEY>-1`, `-2`) count as worktree sets.
- Session PRs merge with `gh pr merge --merge`, never squash.

## 1.3.0 — 2026-10-20

The phase flow.

- Added the `pm`, `phase` and `quick-fix` skills.
- Added reviewer agents (`pal-review-code`, `-tests`, `-errors`, `-security`, `-contract`, `-ui`, `-plan`) and `pal-retro`.
- Added templates: spec, plan, phase state, features, north star, decisions, later, sign-off PR, retro and test summary.
- Rewrote the `workflow`, `testing`, `ssot_standard` and `devops_cicd` practices.
- Added the `state` package. `check` now validates `doc/features.yaml` and `specs/*/phase-state.yaml`.
- `pin` now writes the skill and agent loaders into `.claude/skills/` and `.claude/agents/`, and `check` reports loader drift.
- Added `pal-harness render <file.md>` for phone-readable pages.

## 1.2.0 — 2026-10-19

- Added the **Code patterns** rule to the Develop phase of `workflow.md`: clean layers (business logic only in the use-case layer), SOLID and DRY apply to every component on every stack, and a breach is a review finding.

## 1.1.0 — 2026-10-19

- Added the **Review** rule to the Develop phase of `workflow.md`: sessions run `pal-review` (Codex) before a PR that changes UI or code, handle findings by the severity table, and report the result in the PR.

## 1.0.0 — 2026-10-17

Final-review fixes.

- The `ssot.yml` check is strict.
- The practices agree on the split API rule and the override rule.
- `adr new` and `pin` require the SSOT root, and `pin` keeps overrides.
- Numbering works across in-flight branches.
- Manifest validation is tighter.
- Added a release guard and a Windows CI job.
- One finding per problem.

## 1.0.0-rc.1 — 2026-10-17

- Repo skeleton: Go module, embedded content, plugin manifest, CI.
