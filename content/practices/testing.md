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

- Mocks generated via `mockery --all --keeptree`. Never hand-write mocks.
- Integration tests use `testcontainers-go` for MySQL + Redis — no shared test DB.
- Table-driven test pattern is the default: `[]struct{ name, input, expected }`.
- `//go:build integration` tag keeps unit test runs fast.

## Next.js / React-Specific

- React Testing Library: query by role and label, not by CSS class or test ID.
- MSW (Mock Service Worker) for API mocking — no `jest.mock('fetch')` or axios mock adapters.
- Playwright for E2E: test critical user journeys, not every UI state.
- Snapshot tests are discouraged — they break on styling changes and don't verify behaviour.

## What Not to Test

- Framework internals (Gin routing, GORM query building, Next.js rendering)
- Generated code (wire_gen.go, swaggo docs, shadcn/ui components)
- One-line utility functions with no branching logic
- Third-party library behaviour
