# SSOT Standard

Every project is one **SSOT repo** (the project root; sessions start here) plus one or more
**component repos**. A project may be **single-repo**: every component's `repo` in `project.yaml`
is the SSOT repo (the phase skill's **Single-repo projects**). This file says where a fact lives
and which source wins. Section numbers refer to the SSOT standard design spec.

Tools: `pal-harness <command>` runs as
`go run github.com/althenlimzixuan/pal-harness/cmd/pal-harness@v<VERSION> <command>`, where
`<VERSION>` is the content of `.pal-harness/VERSION`. Commands: `check`, `generate`,
`render <file.md>`, `spec new <slug>`, `adr new <slug> [--conflict]`.

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
| Infrastructure diagram (runtime parts) | `doc/infra.mmd` |
| External API facts (auth, limits, pricing) | `doc/third-party/` |
| Runbooks, release notes | `doc/runbooks/`, `doc/releases/` |
| Features: status, test script, results | `doc/features.yaml` |
| North star | `doc/north-star.md` |
| Owner answers to phase questions | `doc/decisions.md` |
| Parked reviewer findings | `doc/later.md` |
| Designed work and saved phase state | `specs/NNN-slug/` |
| Test reports | `_test_reports/<component>/<date-phase>/` |
| Raw feedback notes | `_feedbacks/YYYY-MM-DD-HHmm.md` |
| Read-only reference material | `context/` |
| Postman collections and environments | `postman/` (no secret values) |
| Pinned harness copy | `.pal-harness/` |
| Project-specific agents and marked overrides | `.claude/` |
| Component code, code-level detail, its bugs and issues | the component repo |
| Backlog (bugs, findings) | the backlog issues, not a file |

- **Placement rule:** folders a tool owns stay where the tool expects them (`.github/`, `.claude/`,
  `.postman/`, `postman/`, and in components `.vercel/`, `eas.json`). Everything people write goes
  under `doc/`. Config only one component uses stays in that component; the SSOT records its
  facts in `doc/infra/`.
- **Deliberately absent from the SSOT repo:** `docs/`, `.specify/`, `feedbacks/`, `doc/feedback/`,
  loose files at the root (`_test_reports/` and `_feedbacks/` are the allowed root folders), markdown backlog or roadmap ledgers, and anything holding credentials.
  Do not create them.
- Tech-stack decisions are recorded in an SSOT ADR. A component may mirror one, citing the ADR
  number; the SSOT wins.
- `.superpowers/` is gitignored scratch. Never cite it as truth.
- **Pinned copy (§6.2, §6.3):** never edit `.pal-harness/` by hand. An override needs the owner's
  decision and a `harness.overrides` entry (item, source, reason, since) in `project.yaml`. A patch
  goes back upstream only by a PR to `pal-harness`.
- **Flow overrides:** `harness.flow_overrides` in `project.yaml` lists the phase-flow changes a
  project turns on (`auto-promote`, `no-owner-test`, `installer-ready`, `ci-test-summary`), with
  the owner's reason in its `CLAUDE.md` Workflow Overrides. The phase skill's **Flow overrides**
  table says what each does; `check` fails on an unknown name.

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
- Feature PRs into `dev` that the owner merges by hand may be squash-merged. PRs a session merges itself (phase, fix-now, quick fix, pm) use `gh pr merge --merge`.
- Before production, which promotions merge on their own and which wait for the owner: `workflow.md`.
- `hotfix/*` branches from `main`, PRs to `main`, then is merged back into `staging` and `dev`.
- Each repo's GitHub default branch is `dev`, so new PRs target it.
- No direct commits to `main`. Branch protection is not enforced; the rule is by convention.

## 5. The phase folder (§12.3)

- `specs/NNN-slug/` is the only home for designed work. Start one with `pal-harness spec new <slug>`
  (creates the folder on branch `docs/NNN-slug` from `dev`). Never use `docs/superpowers/`.
- It holds `phase-state.yaml` (saved state, written only by the session), `phase-log.md` (append-only
  journal), `spec.md` and `plan.md` (each with a generated `.html`), `diagrams/*.mmd`, `reviews/`
  (Codex and Claude reviewer reports) and `retro.md`.
- `spec.md` follows the harness template: numbered summary, questions, calls I made, features with
  test scripts, UX outline, contract (only when the API changes), out of scope, appendix.
- Each Roadmap card names its `specs/NNN-slug/`, and the folder names the card (§4.6).
- A project that adopts the phase flow late sets `harness.legacy_specs_before` to its first phase-flow spec number; older specs keep their old format and are history.

## 6. Which source wins, and the boards (§12.8, §5.4)

- `doc/features.yaml` is the truth for features: status, test script, results. The Roadmap board is a
  projection of it (a draft card titled `F-012 — <name>`), and stays the truth only for the owner's
  intent on a card (approved or not). Backlog issues stay the truth for bugs and findings.
- `doc/decisions.md` holds owner answers to phase questions; expensive architecture choices still go
  to ADRs. `doc/later.md` holds parked findings, kept out of GitHub so they stop feeding the backlog.
- Never keep a backlog or roadmap ledger in a repo; a markdown roadmap is at most a generated snapshot.
- Product-level feedback files to the SSOT repo; a bug or issue for one component files to that
  component's repo. Issue references are repo-qualified (`owner/repo#431`).

## 7. Generated artifacts flow down the ladder (§4.5)

- A generated file has one origin and flows **provider, then SSOT copy, then consumers**.
- Consumers read the SSOT copy or the published client, never the provider's working tree.
- Never edit a copy by hand; regenerate at the origin.
  The §4.5 copy drift check is manual until a later harness release: before merging, compare
  each copy with its origin.
- The Postman collection in `postman/` is regenerated from the SSOT `openapi.yaml` whenever the
  contract changes (§4.7). Regenerating it is manual until a later harness release.

## 8. Feedback notes (§4.7)

- Raw notes are `_feedbacks/YYYY-MM-DD-HHmm.md`, committed to the SSOT repo.
- `_feedbacks/` holds raw notes; the Feedback board holds triaged items.
- Once a note is filed, its header lists the card links so it is never triaged twice.

## 9. Secrets

No secret is ever committed to any repo: no tokens, keys, `.env` files, recovery codes or
credentials, in any file, including docs and Postman environments. Record facts about where a
secret lives, never the value.
