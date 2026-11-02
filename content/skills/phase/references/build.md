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

1. **Bot PRs first.** Before cutting a branch in a repo, list its open bot PRs into `dev`
   (`gh pr list -R <owner/repo> --base dev --author app/github-actions`):
   1. no checks → `gh pr close <n> -R <owner/repo>`, then `gh pr reopen <n> -R <owner/repo>`: a
      reopen by the session fires `pull_request`, so the full CI runs, in every repo, with or
      without `workflow_dispatch:`. Watch it with `gh pr checks <n> -R <owner/repo> --watch`;
   2. green → `gh pr merge <n> -R <owner/repo> --merge`, then branch from the new `origin/dev`;
   3. red → leave it and note it in the log. If the phase's CI then fails on the same check, merge
      the bot's branch into the phase branch instead of fixing it by hand; that is not a fix round.
2. One `feature/NNN-slug` branch per component repo the plan touches, each a worktree in the set at
   its project path:
   `git -C <project root>/<component> fetch origin`, then
   `git -C <project root>/<component> worktree add -b feature/NNN-slug <set>/<component> origin/dev`.
3. SSOT work (the `doc/openapi.yaml` copy, `doc/API.md`, reports, reviews) goes on `docs/NNN-slug`
   in the set's root.
4. Never change a branch in the owner's checkouts.
5. Single-repo project (SKILL.md, **Single-repo projects**): no component worktrees; the code goes
   on `docs/NNN-slug` in the set's root.

## 3. Each task

1. Write the failing test from its fixture (`testing.md`); run it and see it fail for the right
   reason. Keep that red run (the command and its failing output) for the log.
2. Write the code that makes it pass, by the build rules in `workflow.md` (TDD, the code patterns,
   the stack practice, contract-first). A change to owner-facing behaviour updates the project's
   owner guide in the same PR.
3. Run the plan's prove-it commands, the repo's lint among them (through its own script, `workflow.md`);
   read each result from the command's own exit code. The task is not done until lint is clean.
4. Commit with the plan's message; push the feature branch.
5. Add 1 to `tasks.done` and write the log entry with `pal-harness log` (it stamps
   `last_progress_at`; SKILL.md, **Saved state**); commit and push `docs/NNN-slug`.

**Larger phase:** follow `superpowers:subagent-driven-development`.

1. Hand each fresh agent its task text, its worktree path, and the paths of `testing.md`,
   `workflow.md` and the stack practice. The brief says TDD is required for the task: run the
   fixture's test before any code, and put the red run (the command and its failing output) in the
   report, then the green run and a clean lint run.
2. sdd's scripts (`sdd-workspace`, `task-brief`, `review-package`) work in the git repo of the
   current folder, which may be the owner's checkout: never run them from the session's own folder.
   Run each in a subshell started in the set (SKILL.md, **The worktree set**; `<sdd>` is the
   folder of the `subagent-driven-development` skill):
   - `(cd <set> && bash <sdd>/scripts/sdd-workspace <set>/specs/NNN-slug/plan.md)` prints the
     workspace, `<set>/.superpowers/sdd/<plan>/` (it writes its own `.gitignore`); the ledger, the
     briefs and the agents' report files go there;
   - `(cd <set> && bash <sdd>/scripts/task-brief <set>/specs/NNN-slug/plan.md N)`;
   - `review-package` runs in the repo the task changed, with its output in that workspace:
     `(cd <set>/<component> && bash <sdd>/scripts/review-package <set>/specs/NNN-slug/plan.md <base> <head> <workspace>/review-<base7>..<head7>.diff)`.
3. Check every implementer's and fixer's report for the red run. If it is missing, send the task back
   to the agent; that counts as one of its fix rounds. The per-task review does not pass without a
   clean lint run.
4. Review each task when its agent finishes; fix rounds for these per-task reviews: up to 5. Update
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
task made progress (no commit) for 1 hour. A spending-limit block with a green local run is not
stuck (section 6 step 2).

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
2. CI red → read why first (`gh run view <id> -R <owner/repo>`). A job that never started because
   of the spending limit or billing is not a fix round and does not make the build stuck: run CI
   locally as `devops_cicd.md` (**CI runs and minutes**) says; a green local run counts as this
   PR's CI for the merge gates (`review.md` section 6, quick-fix, sign-off). Otherwise fix, push,
   watch again (each fix is a round; 3 rounds, then section 5). Green on every PR → `ci: green`.
3. `stage: code-review`; bundle features `code-review`. Commit and push.
4. Go to `.pal-harness/skills/phase/references/review.md`.
