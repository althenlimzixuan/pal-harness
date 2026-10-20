# Development Workflow Practices

Every project runs one phase flow. Detail lives in the skills: `.pal-harness/skills/phase/SKILL.md`
(spec to ship) and `.pal-harness/skills/pm/SKILL.md` (choosing the work). A project's CLAUDE.md may
override specifics under a **Workflow Overrides** heading (a client UAT gate, a production gate on
`main`, where the owner tests). State the scope at kickoff.

## The flow

North star, then Pal picks a bundle, then spec, plan, build, review, sign-off, the owner's staging
test, ship, retro. The four planned owner stops are the **spec**, the **plan**, the **sign-off** PR
and the **staging test**. Besides those, a session asks the owner only for a major or critical
finding, an exhausted fix-round limit, or a stuck build, always with a "needs you" notification.
Everything sent to the owner opens with numbered points, reads in about 10 minutes and works on a phone.

1. **Choose.** `/pm` proposes a bundle (one roadmap feature plus nearby fixes, about a working day) and
   a vehicle: `phase`, `quick-fix` (no spec or plan, still reviewed, adds a test step) or `owner-task`.
   Reviewer findings enter the backlog only as a real bug or tied to a "need" feature with the reason;
   everything else goes to `doc/later.md`.
2. **Spec.** Draft it alone from the north star, the card and `doc/decisions.md`. Small choices go in
   "Calls I made". Ask only what changes the build: numbered, at the top, each with a recommended
   answer, so one reply works: `accept all` takes every recommended answer; `accept all except 2: B`
   changes one. Each feature carries a numbered test script (step: what you do, what you should see).
   Codex reviews the spec first; after the owner asks for changes, Codex re-checks unless the reply
   says `skip Codex` (recorded). Codex never reviews the plan.
3. **Plan.** After the spec is approved. Every task names its repo, files, the failing test and JSON
   fixture to write first, the code, the commands that prove it, and the commit. Claude's plan reviewer
   checks it; build starts only after the owner approves. "include load test" in the approval adds the
   staging load test.
4. **Build.** Quick fix: in the session, one review. Small (up to ~4 tasks, one repo): in the session,
   one whole-branch review. Larger (5+ tasks, several repos or an API change): one fresh agent per
   task, a review after each, a whole-branch review at the end. The owner does not watch the build;
   notify only for needs you, done and stuck.
5. **Review.** Codex reviews code every phase and the UI when screens changed. Claude's six reviewers
   (code, tests, errors, security, contract, UI) run when their trigger applies. Fix rounds: 3, or 5
   for per-task reviews in a larger build; then ask the owner ("needs you").
6. **Sign-off.** Tests and reports (`testing.md`) pass (regression too, until the project has staging), then one `[HUMAN GATE]` sign-off PR per phase in
   the SSOT repo, readable in 10 minutes. Merging it approves the whole phase.
7. **Staging test.** Staging is built fresh with seed data; regression (and an opted-in load test) runs
   there and its report joins the phase folder before the owner starts. The owner runs each feature's test script:
   pass or fail per step. A failure is **fix now** or **defer** (a backlog issue; the feature is
   `deferred`, not done). Anything else the owner raises: fix now when small and within the phase,
   otherwise defer as `feedback`. A feature is done when it passes on staging.
8. **Ship and retro.** After the owner passes or defers, promote to `main`, deploy, smoke-check and tear
   staging down. Write `retro.md` (went well, went badly); file each improvement idea as a harness
   capture, or in `_feedbacks/` until the capture command exists.

Saved state lives in `specs/NNN-slug/phase-state.yaml`; `ssot_standard.md` says where each file goes.
Every 5 phases Pal walks the north star, the feature list and the later list with the owner.

## Branches and merges (before production)

- Work branches (`feature/ fix/ chore/ docs/`) branch from `dev` and return to `dev` by PR (a feature
  PR may be squash-merged). It merges on its own once CI and the reviews pass.
- The owner merges the `[HUMAN GATE]` sign-off PR `dev → staging`. `staging → main` merges on its own
  after the owner's staging test is passed or deferred. Promotions use merge commits.
- A project in production adds the owner's gate on `main` as a Workflow Override.
- A "fix now" during a staging test branches from `staging`, PRs into `staging`, and is merged back
  into `dev`.
- `hotfix/` branches from `main`, PRs to `main`, then merges back into `staging` and `dev`. No direct
  commits to `main`.

## Build rules

- **TDD**: write the failing test from a JSON fixture first (`testing.md`), then implement.
- **Code patterns** (every component, every stack):
  - **Clean layers**: business logic lives only in the use-case layer (Go: `usecase/`; other stacks: a use-case module per domain, kept out of components, screens, stores and views). Entities and models are plain data: no operations, no I/O, no framework calls. Handlers, screens and admin views stay thin: parse input, call a use case, render the result.
  - **SOLID**: one responsibility per type and file; small interfaces; depend on interfaces, not concrete implementations.
  - **DRY**: one home for each rule, value and query. Search for an existing helper before writing one; a second copy of the same logic is a defect.
  - A breach of these patterns is a review finding like any other.
- Follow the stack practice (`go_api` / `nextjs_web` / `react_native` / `django_admin`).
- **Contract-first** for any REST API: the spec's contract section comes before code. Keep it in sync by the split rule in `product_docs.md`: the provider's PR regenerates `openapi.yaml` and serves Swagger UI; in the same piece of work, the SSOT repo's PR updates its `doc/openapi.yaml` copy and `doc/API.md`.
- Record expensive, hard-to-reverse choices as ADRs (`doc/ADR/`); owner answers to phase questions go in `doc/decisions.md`.
- CI/CD per surface: `devops_cicd.md`.

## Codex review

Before opening a PR that changes UI or code, run `pal-review`. Kinds: `uiux` for UI changes, `code` always, `doc` for specs (arrives with Pal Center 0.4.0; until then record "Codex spec review: not available" and go on). Handle findings by the severity table and put the report in the PR body: the verdict, every major and critical finding with the owner's decision, and the minor ones with what Claude did. Also save the full report as `specs/NNN-slug/reviews/codex-<kind>-YYYY-MM-DD-HHmm.md` in the SSOT repo.

- `pal-review.exe` is not on PATH: call it by the absolute path in the session-start line `Codex review CLI: "<path>\pal-review.exe"` (written `pal-review` here). Run each command with a Bash timeout of 600000 ms (the 120 s default would cut it off). `run` prints `Review <id> queued` first.
- Run from the SSOT root (a phase runs it from its worktree set; see the phase skill): `pal-review run --kind uiux`, `pal-review run --kind code` (`--repo <path>` takes a project-root-relative repo path such as `api`).
- Exit codes: 0 report printed; 3 still running (run `pal-review wait <id>` with the id from `Review <id> queued`, same Bash timeout, until it ends); 4 failed; 5 skipped (put the printed reason in the PR body as the recorded skip: "Codex review: skipped (<reason>)").
- Severity:

  | Level | Meaning | What the session does |
  |---|---|---|
  | none | Nothing worth changing | Review passes |
  | minor | Polish or a local fix; no change to a journey, a contract or data | Claude fixes it, or dismisses it with a one-line reason in the report |
  | major | Changes a journey, a shared pattern, the design system, or more than one screen or module | Claude stops and asks the owner: the finding, why it matters, and the impact |
  | critical | Broken or unsafe: data loss, security, a journey that cannot be completed, an accessibility blocker | As major, listed first |

- Claude may argue a finding up or down a level, but a downgrade from major or critical is never silent: it is written in the report with the reason, and the owner still sees it.

## Workflow Overrides

Per-project deviations live in that project's CLAUDE.md. If none, the flow above applies as-is.
