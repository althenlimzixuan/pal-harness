---
name: phase
description: Use to run an approved card through the phase flow (spec, plan, build, review, sign-off, the owner's staging test, ship and retro), or `/phase resume NNN` to continue one from its saved state. Use when /pm routes a card with vehicle phase, when the owner says /phase, or when a session was started on a phase. Run from the project's SSOT root.
---

# /phase — running one phase

You run one phase: from a card the owner approved to features the owner tested and shipped. Pal's
`/pm` (`.pal-harness/skills/pm/SKILL.md`) chose the work, owns the pickup ladder and writes
`todo`, `recommended` and `picked`; every later feature status is yours. A quick fix runs by
`.pal-harness/skills/quick-fix/SKILL.md` instead. `pal-harness` runs as `ssot_standard.md` says.

## Invocations

- `/phase` (or a route from `/pm`) with an approved bundle: **Start**.
- `/phase resume NNN`: **Resume**.
- `/phase status`: report the stage, `waiting_on_owner`, tasks and CI from `phase-state.yaml`; do
  nothing else.

## Stages

On entering a stage, and on resuming into one, read its reference first; read only that one.

| `stage:` | What happens | Bundle feature status | Reference |
|---|---|---|---|
| `spec` | Draft the spec alone; Codex reviews it | `brainstorming` | `.pal-harness/skills/phase/references/spec.md` |
| `spec-review` | Owner stop: the owner answers and approves | `spec-review` | same |
| `planning` | Write the plan; `pal-review-plan` checks it | `planning` | `.pal-harness/skills/phase/references/plan.md` |
| `plan-review` | Owner stop: the owner approves the plan | `plan-review` | same |
| `building` | Build the tasks | `building` | `.pal-harness/skills/phase/references/build.md` |
| `code-review` | Codex and Claude's reviewers; the test reports | `code-review` | `.pal-harness/skills/phase/references/review.md` |
| `signoff` | Owner stop: the `[HUMAN GATE]` sign-off PR | `code-review` | `.pal-harness/skills/phase/references/signoff.md` |
| `staging-test` | Owner stop: the owner runs each test script | `not-tested`, then `passed`, `fixing`, `retest` or `deferred` | `.pal-harness/skills/phase/references/staging-test.md` |
| `shipping` | Promote to `main`, deploy, tear staging down, retro | final | `.pal-harness/skills/phase/references/ship.md` |
| `done` | Nothing left; the close-out PR merged | final | — |

Never fake a stage: a gate without its artifact (an approved spec, a merged PR, a recorded result)
is not passed.

## The worktree set

Sessions never switch branches in the owner's checkout. A phase works in its own git worktree set
under `<workspace>/.pal-worktrees/<KEY>-<NNN>/`: the SSOT repo at the set's root and each component
repo the phase touches at its project path inside it (`<KEY>-096/api`). `<workspace>` is the folder
holding `workspace.yaml` (a parent of the project); `<KEY>` is `project.yaml` `key`. A session Pal
Center started works in a set already: `<workspace>/.pal-worktrees/<KEY>-1` or `-2`. Any session
whose folder is under `.pal-worktrees/` is in a set; a session the owner starts by hand uses
`<KEY>-<NNN>`. Use absolute paths or `git -C <path>` for
every command and never `cd` into the set, so the owner's checkouts never move. `/pm` uses `<KEY>-pm-<YYYYMMDD-HHmm>` for its own
bookkeeping, never a phase's set.

## Start

1. Read the bundle and vehicle from the lead card's first line (`Bundle: … · vehicle: phase`), and
   the bundle features from `doc/features.yaml` on `origin/dev` after `git fetch origin`.
2. Unless started in a set already (any folder under `.pal-worktrees/`: `<KEY>-1`, `-2` or `<KEY>-<NNN>`),
   create the SSOT worktree:
   1. `git -C <project root> worktree add --detach <workspace>/.pal-worktrees/<KEY>-new-<slug> origin/dev`
   2. `pal-harness spec new <slug> <workspace>/.pal-worktrees/<KEY>-new-<slug>` (the folder
      argument; never `cd` into the set, since Windows cannot move a folder that is a process's
      current folder). It creates `specs/NNN-slug/spec.md` on the new branch `docs/NNN-slug` from
      the local `dev` (none yet → `git -C <project root> branch dev origin/dev` first).
      If `git -C <project root> rev-parse dev` differs from `origin/dev`, run
      `git -C <workspace>/.pal-worktrees/<KEY>-new-<slug> reset --keep origin/dev` next: this moves
      only `docs/NNN-slug` (the new folder is untracked and stays); never reset `dev`.
   3. `git -C <project root> worktree move <workspace>/.pal-worktrees/<KEY>-new-<slug> <workspace>/.pal-worktrees/<KEY>-<NNN>`
   Component worktrees join the set when building starts (`references/build.md`).
3. Write `specs/NNN-slug/phase-state.yaml` from `.pal-harness/templates/phase-state.yaml`: phase,
   slug, title, `vehicle: phase`, the bundle, `stage: spec`, `session:` (this session's name, if
   known; otherwise `""`), `last_progress_at` now, every count zero, every gate `null`.
4. Start `specs/NNN-slug/phase-log.md` with its first entry. Set the bundle features' `phase:` to
   `"NNN"` and their status to `brainstorming`. Commit, then `git push -u origin docs/NNN-slug`.
5. Add the line `Spec: specs/NNN-slug/` to the lead card's body, under the bundle line, and move the
   feature cards to `In flight`.
6. Go to `references/spec.md`.

## Resume (`/phase resume NNN`)

1. `git -C <project root> fetch origin`; find `origin/docs/NNN-*`. No such branch → say so and stop.
2. Use the set at `<workspace>/.pal-worktrees/<KEY>-<NNN>/` if it exists; check `git status` and
   unpushed commits in each worktree before anything else. Missing → recreate it:
   `git -C <project root> worktree add <set> docs/NNN-slug` (git makes the local branch from the
   remote one), then each component's `feature/NNN-slug` the same way. A local branch that already exists
   (`docs/NNN-slug`, `feature/NNN-slug`) is reattached, then `git -C <path> pull --ff-only` in that
   worktree.
3. Read `phase-state.yaml`, the last entries of `phase-log.md` and the stage's reference. They and
   `git log` win over anything you remember.
4. Set `session:` to this session's name if known, write a "resumed" log entry, push.
5. `waiting_on_owner` set → send its link again and wait (**Owner stops** step 4). Exception: at
   `stage: signoff` waiting for another phase's staging test, check that phase; once it is `done`,
   clear `waiting_on_owner`, push and carry on (`references/signoff.md` section 1). Otherwise carry
   on with the stage.

## Saved state

`specs/NNN-slug/phase-state.yaml` is the phase's single recovery map: after a compaction, a crash or
a new session, the phase continues from it and `git log` alone.

1. **The session is its only writer.** Pal Center and `/pm` read it from the pushed `docs/NNN-slug`
   branch (from `dev` once the phase is `done`). Write only the template's keys; `pal-harness check`
   rejects unknown ones. Quote `phase: "NNN"` and gate dates (`"YYYY-MM-DD"`).
2. Update `last_progress_at` (RFC 3339 with offset) at every task and every stage change; `tasks`
   as tasks finish; `ci` as CI changes; `stage` on entering a stage; a gate's date when it passes.
3. **Commit and push `docs/NNN-slug` after every `phase-state.yaml` write.** Pal Center reads the
   pushed branch, and its stuck detection reads `last_progress_at`: an unpushed write makes a busy
   phase look stuck after 1 hour.
4. Each write gets a `phase-log.md` entry (`## YYYY-MM-DD HH:MM — <headline>`, a few lines); `notes`
   holds one-line breadcrumbs only.
5. **Feature statuses** for the bundle's features go in `doc/features.yaml` on the same
   `docs/NNN-slug` branch (`/pm` reads them there), in the same commit as the stage change. Move
   each feature's Roadmap card to the column its status maps to (the table and `gh` commands are in
   the pm skill's **Reconcile** step 6 and **How you write**).
6. Run `pal-harness check` in the SSOT worktree before pushing a stage change; fix every finding.
7. Write YAML and Markdown with the Write or Edit tool and commit messages from a file
   (`git commit -F`), never through a shell argument or heredoc that may hold a backtick.

## Owner stops

The four planned stops are the **spec**, the **plan**, the **sign-off** PR and the **staging test**.
Besides those, ask the owner only for a major or critical finding, an exhausted fix-round limit (3;
5 for per-task reviews in a larger build), a stuck build, or a question that would change the approved spec or plan — always with a "needs you"
notification. Everything else is decided by you and listed under "Calls I made".

| Stop | `waiting_on_owner.kind` | `link` |
|---|---|---|
| Spec | `spec-review` | the spec page |
| Plan | `plan-review` | the plan page |
| Sign-off | `signoff` | the PR URL |
| Staging test | `staging-test` | the spec page (its Features hold the test scripts) |
| Major or critical findings, an exhausted fix-round limit | `findings` | the page of `reviews/owner-YYYY-MM-DD-HHmm.md` |
| A stuck build or another blocking question | `questions` | the page of `reviews/owner-YYYY-MM-DD-HHmm.md` |

For the last two, write the batch or the note to `specs/NNN-slug/reviews/owner-YYYY-MM-DD-HHmm.md`
first: the problem, what was tried, and numbered options with a recommended one.

At every stop:

1. Everything sent reads in about 10 minutes, opens with numbered points and works on a phone.
   Questions are numbered, each with a recommended answer, so one line ("accept all") replies.
2. Publish the page: `pal-harness render <file.md>` writes `<file>.html` beside it (diagrams
   drawn). Publish that file as a private claude.ai page with the Artifact tool when this session
   has it; otherwise send the file. A PR stop links the PR.
3. Set `waiting_on_owner: {kind, since, link, questions}` (`questions` counts the numbered
   questions, 0 if none) and the bundle features' statuses; commit and push.
4. Send a PushNotification that starts "needs you": the phase, what to read, how long it takes, the
   link. If the PushNotification tool is not available, write the "needs you" text in the session,
   say it could not notify, and keep waiting. Keep the session open and waiting; the owner reads on the phone and answers in this
   session through Remote Control. Do not end the session and do not poll.
5. When the owner replies: record the answer (owner answers to questions go in
   `doc/decisions.md`), clear `waiting_on_owner`, update `last_progress_at`, commit and push.
6. Work state: at a stop run `pal-work status --session <name> --state waiting` (the `Work CLI:`
   line gives the path; `<name>` is the session name in phase-state `session:`; skip this when the
   line is missing or the CLI exits 5), after the owner's reply `--state working`, at `done` `--state idle`.
7. Then compact before continuing: run `/compact` when this session can, or reload from
   `phase-state.yaml`, the stage's reference and the approved files, never from memory of the
   earlier stage.

The owner does not watch the build. Notifications go out only for **needs you** and **done** (from
the session) and **stuck** (from Pal Center over Slack, once it exists).

## Plans and builds use superpowers

- Plans are written with `superpowers:writing-plans`, saved into `specs/NNN-slug/plan.md` (never
  `docs/superpowers/`).
- A larger build runs with `superpowers:subagent-driven-development`.
- The spec is drafted by this skill, not by `superpowers:brainstorming`, which asks one question at
  a time.
- Without superpowers installed, say so in the plan and the log, write the plan by
  `.pal-harness/templates/plan.md` yourself, and build inline.

## Until Pal Center and CI catch up

The session-start line `Pal Center version: X.Y.Z` is the installed Pal Center. The `doc` Codex
kind, reviews from a worktree set and the Features tab need 0.4.0 or later. The lines
`Work CLI: "<path>"`, `Capture CLI: "<path>"` and `Codex review CLI: "<path>"` give the paths of
`pal-work`, `pal-capture` and `pal-review`. A CLI that exits 5 means Pal Center is not running. With
no version line, an older version or exit 5, use the interim rows.
With Pal Center 0.4.0 or later the staging test collects results with `wait-results --phase NNN`
armed through Monitor, and `results --phase NNN` after a resume (`references/staging-test.md` section 3).

| Missing piece | Do this |
|---|---|
| Pal Center older than 0.4.0 (no `doc` kind, no worktree-set reviews) | Record "Codex spec review: not available" and `<kind>: not available (worktree set)` in `codex.skipped` and go on; never run `pal-review` from the owner's checkout for a phase |
| Features tab and `pal-work` (older than 0.4.0, or exit 5) | Walk the owner through the test script step by step through Remote Control and record the results in `doc/features.yaml` (`env: local` until the project has staging) |
| Capture command (no `Capture CLI:` line, or exit 5) | Retro ideas go to `_feedbacks/YYYY-MM-DD-HHmm.md` |
| Scheduled pickup, stuck alerts | `/pm pickup` runs the ladder by hand; there is no stuck alert, so send "needs you" yourself when stuck |
| CI report jobs, fingerprint skip | Run the reports once, locally, on the final commit |
| On-demand staging | The owner tests on the local stack before `staging` → `main`, so the order of sign-off, test and ship holds. A project whose Workflow Overrides say "test on production" merges to `main` first and tests after |

Session PRs (phase build PRs, fix-now PRs, quick-fix PRs) merge with `gh pr merge --merge`, never
squash: Pal Center tells a session folder's work is safely on GitHub from the merge commit after
the branch is deleted.

## Rules for every stage

- One card at a time per project (pm skill, **One card at a time**). The next phase may build while
  this one waits for the staging test; its sign-off PR waits until this test is done.
- Never force-push. On a rejected push, fetch and merge (or `git -C <path> pull --ff-only`), then push.
- Never merge a `[HUMAN GATE]` PR and never turn on auto-merge for one.
- Secrets come from the environment only. Never open `.env` or credential files.
- Read a build or test result from the command's own exit code, unpiped.
- The process itself changes only through retro ideas, never by editing `.pal-harness/`.
- Remove the set when the phase is `done` and nothing in it is unpushed (`git status` clean and
  `git log '@{u}..'` empty in every worktree; run git through Bash, and on a detached HEAD compare
  with `git -C <path> log origin/<branch>..<branch>`): `git worktree remove` each component worktree, then the
  SSOT one (`references/ship.md`).
