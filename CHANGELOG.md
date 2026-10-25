# Changelog

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
