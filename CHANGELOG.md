# Changelog

## 1.3.2 — 2026-10-21

A project that adopts the phase flow late can leave its old specs alone.

- Added the optional `harness.legacy_specs_before` field to `project.yaml` (three digits). `check` skips `phase-state.yaml` in `specs/NNN-*` below it; later specs are validated as before.

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
