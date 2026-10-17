# SSOT Standard

Every project is one **SSOT repo** (the project root; sessions start here) plus one or more
**component repos**. This file says where a fact lives and which source wins. Section numbers refer
to the SSOT standard design spec.

Tools: `pal-harness <command>` runs as
`go run github.com/althenlimzixuan/pal-harness/cmd/pal-harness@v<VERSION> <command>`, where
`<VERSION>` is the content of `.pal-harness/VERSION`. Commands: `check`, `generate`,
`spec new <slug>`, `adr new <slug> [--conflict]`.

## 1. Tiers and the precedence ladder (§2)

| Tier | Holds | Rank |
|---|---|---|
| Workspace | `ai-workshop` (project registry), `pal-harness` (this harness), `pal-center` (the app) | cross-project standards only |
| Project SSOT repo | the recorded intent: docs, manifest, specs, ADRs | 0 (wins) |
| Provider and platform components | API, `<key>-infra` | 1 |
| Consumer components | web, mobile_app, admin, api-client, discord | 2 |

- The SSOT outranks any component; the provider outranks consumers; `api-client` is a consumer.
- The workspace tier does not outrank a project's SSOT about that project's own facts.
- The ladder says which source to **show first**, never which to trust silently. When two sources
  disagree about a fact, follow `conflict_protocol.md`.
- Legacy material in `context/legacy/` is reference only, for behavior parity. Where it disagrees
  with the SSOT, the SSOT wins without a conflict ADR.

## 2. Layout: kind of fact, its one home (§4.1)

| Kind of fact | Home |
|---|---|
| Project facts (components, roles, boards, harness pin) | `project.yaml` |
| Who, what, why | `doc/PRD.md`, `doc/SRS.md` |
| Standing principles | `doc/PRINCIPLES.md` |
| Entities, states, glossary | `doc/DOMAIN.md` |
| API contract | `doc/API.md` (human) + `doc/openapi.yaml` (copy of the provider's generated file) |
| Decisions and conflict resolutions | `doc/ADR/NNNN-slug.md` |
| UML (use case, activity, sequence, state, ERD, deployment) | `doc/diagrams/*.mmd` |
| Environments, deploy, cost (facts only) | `doc/infra/` |
| External API facts (auth, limits, pricing) | `doc/third-party/` |
| Runbooks, release notes | `doc/runbooks/`, `doc/releases/` |
| Designed work | `specs/NNN-slug/` |
| Raw feedback notes | `_feedbacks/YYYY-MM-DD-HHmm.md` |
| Read-only reference material | `context/` |
| Postman collections and environments | `postman/` (no secret values) |
| Pinned harness copy | `.pal-harness/` |
| Project-specific agents and marked overrides | `.claude/` |
| Component code, code-level detail, its bugs and issues | the component repo |
| Backlog and phase status | the boards, not a file |

- **Placement rule:** folders a tool owns stay where the tool expects them (`.github/`, `.claude/`,
  `.postman/`, `postman/`, and in components `.vercel/`, `eas.json`). Everything people write goes
  under `doc/`. Config only one component uses stays in that component; the SSOT records its
  facts in `doc/infra/`.
- **Deliberately absent:** `docs/`, `.specify/`, `feedbacks/`, `doc/feedback/`, loose files at the
  root, markdown backlog or roadmap ledgers, and anything holding credentials. Do not create them.
- Tech-stack decisions are recorded in an SSOT ADR. A component may mirror one, citing the ADR
  number; the SSOT wins.
- `.superpowers/` is gitignored scratch. Never cite it as truth.

## 3. Manifest and generated blocks (§4.2, §5.1)

- `project.yaml` is the truth for project facts. Prose (why, context, decisions) never goes in it.
- The component tables in the SSOT `README.md` and `CLAUDE.md` are **generated** from it. They sit
  between `<!-- pal:begin ... -->` and `<!-- pal:end ... -->` markers. Never hand-edit between
  them. Change `project.yaml`, then run `pal-harness generate`. `pal-harness check` fails on drift.
- Each component root holds `ssot.yaml` (project, SSOT URL, role, rank) pointing up to the SSOT.
  Its `CLAUDE.md` names the SSOT URL and imports `@../CLAUDE.md`.

## 4. Branching (§2.1)

- `feature/*`, `fix/*`, `chore/*`, `docs/*` branch from `dev` and return to `dev` by PR.
- Promote `dev` to `staging`, then `staging` to `main`, with merge commits (not squash).
- Feature PRs into `dev` may be squash-merged.
- `hotfix/*` branches from `main`, PRs to `main`, then is merged back into `staging` and `dev`.
- No direct commits to `main`. Branch protection is not enforced; the rule is by convention.

## 5. Specs and plans (§4.4)

- `specs/NNN-slug/` is the only home for designed work. Start one with `pal-harness spec new <slug>`
  (creates `specs/NNN-slug/spec.md` on branch `docs/NNN-slug` from `dev`).
- superpowers writes its design and plan into that folder (`design.md`, `plan.md`), never into
  `docs/superpowers/`.
- `spec.md` follows the template: user stories P1-P3, UX Outline when a surface is active,
  `FR-NNN` requirements, `SC-NNN` success criteria naming their deciding predicate, Key Entities,
  Assumptions. `[NEEDS CLARIFICATION: ...]` markers are allowed.
- Each Roadmap card names its `specs/NNN-slug/`, and the folder names the card (§4.6).

## 6. Boards (§5.4, §4.6)

- The Backlog, Roadmap and Feedback boards are the truth for backlog and phase status. Never keep
  a backlog or roadmap ledger in a repo; a markdown roadmap is at most a generated snapshot.
- Product-level feedback files to the SSOT repo; a bug or issue for one component files to that
  component's repo.

## 7. Generated artifacts flow down the ladder (§4.5)

- A generated file has one origin and flows **provider, then SSOT copy, then consumers**.
- Consumers read the SSOT copy or the published client, never the provider's working tree.
- Never edit a copy by hand. Regenerate at the origin; each hop's drift check fails on mismatch.
- The Postman collection in `postman/` is regenerated from the SSOT `openapi.yaml` whenever the
  contract changes (§4.7).

## 8. Feedback notes (§4.7)

- Raw notes are `_feedbacks/YYYY-MM-DD-HHmm.md`, committed to the SSOT repo.
- `_feedbacks/` holds raw notes; the Feedback board holds triaged items.
- Once a note is filed, its header lists the card links so it is never triaged twice.

## 9. Secrets

No secret is ever committed to any repo: no tokens, keys, `.env` files, recovery codes or
credentials, in any file, including docs and Postman environments. Record facts about where a
secret lives, never the value.
