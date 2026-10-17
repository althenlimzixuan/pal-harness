# Product Documentation Practices

Covers the project's SSOT repo — the **management layer** above all code components. It is the
source of truth for **what** is being built and **why**. Where each kind of document lives is fixed
by the layout in `ssot_standard.md` (spec §4.1); this file does not redefine it. Code practices
(`go_api.md`, `nextjs_web.md`, …) cover *how*; this file covers *what/why* and how docs stay in sync.

## doc/ Structure (the documents this file covers)

```
doc/
  PRD.md              # product requirements — the full brief
  API.md              # canonical human-readable API contract
  openapi.yaml        # generated OpenAPI 3.1 spec (from the Go API)
  ADR/
    0001-record-architecture-decisions.md
    0002-....md
  wireframes/         # exported screens / links to Figma
```

`doc/` lives in the project's SSOT repo, never in a component. Designed work lives in
`specs/NNN-slug/` and raw feedback notes in `_feedbacks/`, both at the SSOT root, not under `doc/`.

## PRD.md — required sections

1. **Problem** — the pain, who feels it, why now
2. **Target users / personas**
3. **Value proposition** — one paragraph
4. **Goals & non-goals**
5. **User journeys** — the critical end-to-end paths
6. **Functional requirements** — numbered (`FR-1`), each testable
7. **Business rules** — non-obvious constraints Claude must not violate
8. **Success metrics / KPIs**
9. **Out of scope / future**
10. **Open questions**

The product `CLAUDE.md` is the **condensed** brief (tagline, users, domain, journeys, key rules).
`PRD.md` is the full detail. Business rules in the PRD **must** also appear in the product
`CLAUDE.md` so every surface inherits them through the `@../CLAUDE.md` import.

## ADR — one decision per file

Numbered, immutable once **Accepted** (supersede with a new ADR rather than editing).

```markdown
# ADR-0002: Use MySQL as primary datastore

## Status
Accepted — 2026-06-10

## Context
[forces at play, constraints]

## Decision
[what we chose]

## Consequences
[trade-offs accepted, follow-ups]

## Alternatives considered
[options rejected and why]
```

Statuses: `Proposed` → `Accepted` → `Superseded by ADR-XXXX`.
Write an ADR for any choice that is expensive to reverse (datastore, auth model, cloud target,
admin stack, API versioning strategy).

## API contract — dual source of truth

A RESTful API has **two** synchronized artifacts:

| Artifact | Role | Produced |
|---|---|---|
| `doc/API.md` | Human-readable contract — endpoints, auth, error model, examples | Hand-written, surface-agnostic |
| `doc/openapi.yaml` | Machine spec for codegen, Swagger UI, contract tests | **Generated from the provider's code**; the SSOT file is a copy of it |
| Swagger UI | Live, browsable docs for consumers (web/mobile/B2B) | Served by the API from the spec |

**Universal rule — any RESTful API change must, in the same PR:**
1. Regenerate the OpenAPI spec (`make docs`) and commit it
2. Update `doc/API.md` to match
3. Keep Swagger UI served (e.g. `/swagger/index.html` in non-prod)

**Generated artifacts flow down the ladder (§4.5):** provider → SSOT copy → consumers. The provider
generates `openapi.yaml` and is its only origin; the SSOT keeps a copy; consumers (`api-client`, web,
mobile) read the SSOT copy or the published client, never the provider's working tree. Each hop has a
drift check that fails when the copy no longer matches its origin.

The Postman collection in the SSOT's `postman/` is the truth for the collection (§4.7). Regenerate it
from the SSOT copy of `openapi.yaml` whenever the contract changes; a cloud Postman workspace is only a
mirror.

CI fails if the committed spec differs from a fresh generation (spec must never drift from code).
`doc/API.md` and `openapi.yaml` must describe the same endpoints — the human doc is the contract
consumers read, the spec is what tools consume.

## Lifecycle (how the docs connect to delivery)

```
discovery → PRD.md → ADRs for expensive choices → API.md + openapi.yaml (contract)
  → build per stack practice → test (80% gate, testing.md) → deliver (devops_cicd.md)
```

Every functional requirement in the PRD should trace to a test; every API endpoint to a spec entry.
