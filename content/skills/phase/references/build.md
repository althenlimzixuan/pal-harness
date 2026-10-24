# Stage: build (`building`)

Build the approved plan's tasks, in order. The owner does not watch the build: send nothing until
the next owner stop unless the build needs them.

## 1. Approach by size

The plan states the approach with its reason; the owner may overrule it.

| Approach | When | How | `approach:` |
|---|---|---|---|
| quick fix | a small fix, its own vehicle | built in the session, then one review (`.pal-harness/skills/quick-fix/SKILL.md`) | `inline` |
| small phase | up to ~4 tasks, one repo | built in this session, then one whole-branch review | `inline` |
| larger phase | 5 or more tasks, or several repos, or an API change | `superpowers:subagent-driven-development`: one fresh agent per task, a review after each task, a whole-branch review at the end | `subagent` |

Without superpowers installed, say so in the log and build a larger phase inline, task by task.

## 2. Branches and worktrees

1. One `feature/NNN-slug` branch per component repo the plan touches, each a worktree in the set at
   its project path:
   `git -C <project root>/<component> fetch origin`, then
   `git -C <project root>/<component> worktree add -b feature/NNN-slug <set>/<component> origin/dev`.
2. SSOT work (the `doc/openapi.yaml` copy, `doc/API.md`, reports, reviews) goes on `docs/NNN-slug`
   in the set's root.
3. Never change a branch in the owner's checkouts.
4. Single-repo project (SKILL.md, **Single-repo projects**): no component worktrees; the code goes
   on `docs/NNN-slug` in the set's root.

## 3. Each task

1. Write the failing test from its fixture (`testing.md`); run it and see it fail for the right
   reason.
2. Write the code that makes it pass, by the build rules in `workflow.md` (TDD, the code patterns,
   the stack practice, contract-first). A change to owner-facing behaviour updates the project's
   owner guide in the same PR.
3. Run the plan's prove-it commands; read each result from the command's own exit code.
4. Commit with the plan's message; push the feature branch.
5. Add 1 to `tasks.done`, update `last_progress_at`, write a log entry; commit and push
   `docs/NNN-slug`.

**Larger phase:** follow `superpowers:subagent-driven-development`. Hand each fresh agent its task
text, its worktree path, and the paths of `testing.md`, `workflow.md` and the stack practice. Review
each task when its agent finishes; fix rounds for these per-task reviews: up to 5. Update
`phase-state.yaml` (step 5) as each task lands, not at the end.

**API change:** the provider's PR regenerates `openapi.yaml` and serves Swagger UI; in the same piece
of work the SSOT copy and `doc/API.md` are updated on `docs/NNN-slug` (the split rule in
`product_docs.md`). Build the contract task before the tasks that call it.

## 4. Questions during the build

Rare. Ask only what would change the approved spec or plan, as an owner stop with
`kind: questions` (SKILL.md, **Owner stops**). A smaller choice is yours: note it for the sign-off
PR's "Calls I made".

## 5. Stuck

The build is stuck when a task failed its fix rounds, CI stays red after 3 rounds of fixes, or no
task made progress (no commit) for 1 hour.

1. Stop that task. Write in the log what was tried and what failed.
2. Write the one-page note (`reviews/owner-YYYY-MM-DD-HHmm.md`: the problem, what was tried, the
   options with a recommended one) and publish it. Set `waiting_on_owner` (`kind: findings` for an
   exhausted fix-round limit, `questions` otherwise). Push.
3. PushNotification: "needs you: phase NNN is stuck — <one line>".
4. Tasks that do not depend on the stuck one may go on meanwhile.

Pal Center also sends its own "stuck" over Slack from `last_progress_at` (piece 3); pushing at every
task keeps that honest.

## 6. Done building

1. Open each `feature/NNN-slug` PR into `dev` (`gh pr create -R <owner/repo> --base dev --head feature/NNN-slug`); do
   not merge yet. Single-repo: open the code PR `docs/NNN-slug` → `dev` instead. Set
   `ci: running`; watch with `gh pr checks <n> -R <owner/repo> --watch`.
2. CI red → fix, push, watch again (each fix is a round; 3 rounds, then section 5). Green on every
   PR → `ci: green`.
3. `stage: code-review`; bundle features `code-review`. Commit and push.
4. Go to `.pal-harness/skills/phase/references/review.md`.
