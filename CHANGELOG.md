# Changelog

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
