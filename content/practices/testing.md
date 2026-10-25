# Universal Testing Philosophy

## Principles

- **TDD-first**: Write the failing test before the implementation. Always.
- **Fixtures over inline data**: Test cases live in `tests/{test_type}/{use_case}.json`. Never hardcode inputs or expected values inline.
- **Test behaviour, not implementation**: If refactoring internals breaks tests, the tests are wrong.
- **Coverage gate**: 80% minimum, enforced in CI. Coverage is a floor, not a target.
- Tests are documentation. A test that requires comments to understand is a bad test.

## Test Fixture Format

```json
{
  "description": "creates a user with valid input",
  "input": { ... },
  "expected": { ... },
  "mocks": { ... }
}
```

Load fixtures via a shared helper:
- Go: `testhelper.LoadFixture(t, "tests/unit_test/create_user.json")`
- Next.js / RN: `loadFixture('tests/unit_test/create_user.json')`

## Test Layers

| Layer | Tag / Config | Scope | Dependencies |
|---|---|---|---|
| Unit | (default, no tag) | Single function or component | All I/O mocked |
| Integration | `//go:build integration` / `jest.config.integration.ts` | Multiple layers together | Real DB + cache via testcontainers or docker-compose |
| E2E | `//go:build e2e` / Playwright | Full user journey | Full stack running |

Run in CI in this order: unit → integration → E2E. Never skip unit if integration passes.

## Go-Specific

- Mocks generated via `mockery --all --keeptree --dir internal --output internal/mocks` (`make mock`,
  same settings in `.mockery.yaml`); inside `internal/`, the mirrored paths are importable. Never
  hand-write mocks.
- Integration tests use `testcontainers-go` for MySQL + Redis — no shared test DB.
- Table-driven test pattern is the default: `[]struct{ name, input, expected }`.
- `//go:build integration` tag keeps unit test runs fast.

## Next.js / React-Specific

- React Testing Library: query by role and label, not by CSS class or test ID.
- MSW (Mock Service Worker) for API mocking — no `jest.mock('fetch')` or axios mock adapters.
- Playwright for E2E: test critical user journeys, not every UI state.
- Snapshot tests are discouraged — they break on styling changes and don't verify behaviour.
- A new npm advisory in test tooling goes through `audit-ci.jsonc` as `devops_cicd.md` (Security in CI) says.

## What Not to Test

- Framework internals (Gin routing, GORM query building, Next.js rendering)
- Generated code (wire_gen.go, swaggo docs, shadcn/ui components)
- One-line utility functions with no branching logic
- Third-party library behaviour

## Test reports

- **One run, on the final commit.** CI runs the tests once on each PR's final commit; that run makes
  the reports, except the staging reports (regression, and the load test when requested), which come
  from the staging deploy of the merged sign-off. A promotion whose code fingerprint (git tree hash) already passed skips the tests and
  only builds and deploys. Reports live in the SSOT repo, so committing them never changes the code
  under test. In a single-repo project the final commit is the last commit that changes code;
  report, review and state commits after it do not make a report stale. Until CI has report jobs, the session runs the reports once, locally, on the final commit.
- **One report command.** Each project has one (`make reports`, or `npm run reports` in a Node-only
  repo) that writes every report the phase needs on the final commit into the phase's `_test_reports/`
  folder (regression and the load test come from staging once it exists). Until CI has report jobs
  (CI's report jobs run the same command), the session runs it once, after the last review fix, on
  the final commit: not before the fixes, and not per reviewer.
- **Location:** `_test_reports/<component>/<YYYY-MM-DD-HH>-<NNN-slug>/` in the SSOT repo, committed.
- **Every phase:** `unit_test.html`, `test_coverage.html`, `regression.html`,
  `profiling.html`, and a numbered `summary.md` (pass/fail counts, coverage %, slowest calls, change
  since the last report, verdict against the thresholds below).
- **Regression timing:** once the project has staging, regression (and the opted-in load test) runs on
  staging after the sign-off PR merges, and its report joins the phase folder before the owner's staging
  test. Until the project has staging, regression runs with the other reports on the final commit,
  before the sign-off PR.
- **On request only:** `load_test.html`, a load test against staging. The owner asks by adding
  "include load test" to the spec reply.
- **Blocking problems:** coverage under 80%; a failing test; an API call's p95 more than 20% slower,
  or a web page's Lighthouse performance score 5 points or more lower, than the previous phase's report
  (the first phase sets the baseline). Claude gets 3 fix rounds, then the session asks the owner (a "needs you" notification).
- A missing, empty or stale report is a review finding. Staging reports must match the tree hash of
  the staging deploy; every other report must come from the PR's final commit.
- **A report the project cannot make yet.** One that does not apply to the stack (Lighthouse with no web
  surface) is recorded as "n/a" with the reason. One that applies but has no tooling becomes a task in
  the project's first plan; until it ships, the report is a recorded skip, not a blocker.
- **Suggested tools:** gotestsum and Go's cover tool, `jest-html-reporter` (not `jest-html-reporters`)
  and Istanbul, pytest-html, Playwright's HTML report, pprof, Lighthouse, k6 (load).
- A report that shows results other than this run's (demo data, a template placeholder) is a
  stale-report finding. Its pass and fail counts must match `summary.md`.
