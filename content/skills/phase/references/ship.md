# Stage: ship (`shipping`, then `done`)

After the owner's staging test is passed or deferred: promote to `main`, deploy, tear staging down,
write the retro, close out.

## 1. Promote to `main`

1. In each component repo the phase touched, and in the SSOT repo:
   `gh pr create -R <owner/repo> --base main --head staging --title "Promote staging to main (phase NNN)"`,
   watch CI, merge with a merge commit (`gh pr merge <n> -R <owner/repo> --merge`). It merges on its
   own until the product is in production.
2. A project in production has the owner's gate on `main` (its Workflow Overrides): title these PRs
   `[HUMAN GATE] …` and wait for the owner as an owner stop (`kind: signoff`); never merge them
   yourself.
3. Already promoted (a "test on production" project): skip this section.

## 2. Deploy and smoke check

1. Production deploys only from `main`, by the project's own pipeline. Watch it to the end.
2. Run the smoke check: the project's runbook check if `doc/runbooks/` has one, otherwise the health
   endpoint and the main journey's first screen. Write the result in the log.
3. A failed deploy or smoke check: "needs you" with what failed (`kind: questions`). Do not start
   another phase on top of it.
4. No production environment yet: write "no production deploy yet" in the log and go on.

## 3. Tear staging down

1. Staging is always torn down after the test (the project's teardown; on the local stack, stop it
   and remove its containers and seed data).
2. A failed teardown is a "stuck" alert: send "needs you" with the error (Pal Center alerts on it
   too, once it exists). Staging also tears itself down after 48 hours idle; never rely on that.

## 4. Statuses and cards

1. The bundle features keep their final status (`passed` or `deferred`); move each card to its
   column (pm skill mapping).
2. Close each bundle issue the phase fixed and the owner passed, if a PR did not close it:
   `gh issue close <n> -R <owner/repo> --comment "Shipped in phase NNN"`.

## 5. Retro

1. Dispatch the `pal-retro` agent with the phase folder; the session files nothing itself and
   reports what the agent filed. The agent writes `specs/NNN-slug/retro.md` from
   `.pal-harness/templates/retro.md`: what went well (keep), what went badly (fix), and the ideas.
   Keep it short.
2. File each idea as a harness capture in Pal Center's inbox with the session-start line
   `Capture CLI: "<path>"`: `"<path>" --kind harness --project <key> "<idea>"` (exit 4 rejected, 5
   Pal Center not running). With no line or exit 5, write the ideas to
   `_feedbacks/YYYY-MM-DD-HHmm.md` in the SSOT repo, one line each, marked `harness`.
3. The good points stay in `retro.md`.

## 6. Close out

1. `stage: done`; clear `waiting_on_owner`; update `last_progress_at`.
2. Open the close-out PR (`gh pr create -R <owner/repo> --base dev --head docs/NNN-slug`) from
   `docs/NNN-slug` into `dev` (the branch's second PR): the retro, the
   results and statuses in `doc/features.yaml`, the final `phase-state.yaml`, any `_feedbacks/`
   note. It merges on its own once CI is green (`gh pr merge <n> -R <owner/repo> --merge`). Pal Center reads
   the phase from `dev` from now on.
3. Once deployed (or, with no production yet, once staging → main merges) and the close-out PR merged, write **recommended next**: run the
   pm skill's **Recommended next** (`.pal-harness/skills/pm/SKILL.md`), which writes through its own
   worktree.
4. PushNotification: "done: phase NNN shipped — <one line>. Retest: N. Deferred: N."

## 7. Clean up

1. In every worktree of the set: `git status` clean and `git log '@{u}..'` empty (run git through Bash; on a detached HEAD use
   `git -C <path> log origin/<branch>..<branch>`). Anything unpushed →
   push it or say what it is; never remove a worktree that holds unpushed work.
2. `git -C <project root>/<component> worktree remove <set>/<component>` for each component, then
   `git -C <project root> worktree remove <set>`. Delete the phase's merged local branches.
