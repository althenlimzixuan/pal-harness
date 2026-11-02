# Changelog

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
