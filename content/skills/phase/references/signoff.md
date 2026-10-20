# Stage: sign-off (`signoff`)

One sign-off PR per phase, in the SSOT repo, `dev → staging`, titled `[HUMAN GATE]`. Merging it is
the owner's approval for the whole phase. Only the owner merges it.

## 1. Before opening it

1. **One staging at a time.** Read the `phase-state.yaml` of every other live phase of this project
   (`origin/docs/[0-9]*` branches). One in `staging-test` or `shipping` holds staging, so pause:
   1. Set `waiting_on_owner` to `{kind: staging-test, since: <now>, link: <that phase's test link
      from its waiting_on_owner>, questions: 0}`, add a note `waiting for phase MMM's staging
      test`, write a log entry, commit and push.
   2. Tell the owner in the session; do not poll. `/pm`'s start-of-session note shows the wait.
   3. The owner (or a later session) continues with `/phase resume NNN` once phase MMM is `done`;
      the resume clears `waiting_on_owner` and starts this section again.
2. Check: every component PR is merged into `dev`; the reports come from the final commits and pass
   (or are recorded "n/a" or skips); no major or critical finding is open; regression passed (until
   the project has staging).

## 2. Bring `docs/NNN-slug` into `dev`

1. The branch carries the spec, the plan, the reviews, the reports, `doc/features.yaml` statuses,
   `doc/decisions.md`, `doc/later.md` and `phase-state.yaml` (`stage: signoff`).
2. Open its PR into `dev` just before the sign-off PR. It merges on its own once CI is green, with a
   merge commit so the same branch can PR again at ship: `gh pr checks <n> --watch`, then
   `gh pr merge <n> --merge`. Keep the branch: Pal Center reads `phase-state.yaml` from it, and the
   close-out PR at ship comes from it.

## 3. Write the sign-off PR

Fill `.pal-harness/templates/signoff-pr.md` (the template's first line is the title), sized for a
10-minute read:

1. **Summary**, numbered: what was built, which features, which of the owner's answers shaped it
   (from `doc/decisions.md`), the review and test verdict in one line, and what the owner does next.
2. **Calls I made:** from the spec and the build.
3. **Review results:** one row per reviewer and per Codex kind: found, fixed, left over and where it
   went (an issue link or `doc/later.md`). Recorded Codex skips go here.
4. **Test reports:** counts, coverage and speed changes, each with its link under `_test_reports/`
   on `dev`.
5. **When you merge:** what happens next (staging, regression, the load test if asked, "ready for
   your test").
6. **Your test scripts:** one link per feature, to its block in `spec.md`.

## 4. Open it and wait

1. Write the body to a file, then in the SSOT repo:
   `gh pr create --base staging --head dev --title "[HUMAN GATE] Phase NNN — <Title>" --body-file <file>`.
2. Never merge it yourself and never turn on auto-merge for it.
3. Owner stop (SKILL.md, **Owner stops**): `waiting_on_owner` with `kind: signoff` and the PR URL as
   the link. Push `docs/NNN-slug`. PushNotification: "needs you: sign-off for phase NNN — <PR URL>,
   about 10 minutes. Reply 'merged' when done."
4. On any reply, read the PR: `gh pr view <n> --json state,mergedAt`.
   - Changes asked: fix them on the branch they belong to (a component fix by the usual PR into
     `dev`; a docs fix on `docs/NNN-slug`, merged into `dev` the same way), update the PR body and
     notify again.
   - Merged: go on.

## 5. After the owner merges

1. `gates.signoff_merged` the date; clear `waiting_on_owner`; commit and push.
2. Promote each component repo the phase touched, `dev → staging`, yourself: open
   `gh pr create -R <owner/repo> --base staging --head dev --title "Promote dev to staging (phase NNN)"`,
   watch CI, and merge with a merge commit (`gh pr merge <n> -R <owner/repo> --merge`). These PRs
   carry no `[HUMAN GATE]`: the owner's merge above approved them.
3. `stage: staging-test`. Commit and push; compact (SKILL.md, **Owner stops** step 6).
4. Go to `.pal-harness/skills/phase/references/staging-test.md`.
