# Stage: staging test (`staging-test`)

The owner runs each bundle feature's test script. A feature is done only when it passes here.

Under `no-owner-test` (SKILL.md, **Flow overrides**) this stage is skipped: sign-off goes straight
to `.pal-harness/skills/phase/references/ship.md`, and nothing here runs.

## 1. The environment

1. **The project has staging:** staging is built from nothing from the `staging` branches, with
   fresh seed data: the owner's test accounts (the same ones Codex's UI review uses, from
   `doc/review-brief.md`) plus sample data. Never production data.
2. **No staging yet:** start the local stack from `origin/staging` of each repo (in the set: `git -C <set>/<component> fetch origin`, then
   `git -C <set>/<component> switch --detach origin/staging`), with fresh seed data the same way.
   Results are recorded with `env: local`.
3. **Workflow Overrides say "test on production":** run the promotion and deploy of
   `.pal-harness/skills/phase/references/ship.md` sections 1–2 first, then this test on production
   (`env: prod`).
4. Mobile: the owner uses Expo Go pointed at the staging API (on the local stack, the API's LAN
   address).

## 2. Regression first

Once the project has staging, run regression (and the load test when `load_test: true`) against
staging now; write the reports into `_test_reports/<component>/<YYYY-MM-DD-HH>-<NNN-slug>/` on
`docs/NNN-slug` before the owner starts. A blocking result: fix it as in section 4 (up to 3 rounds),
then "needs you".

## 3. Ready for the owner

1. Set the bundle features `not-tested`.
2. `waiting_on_owner` with `kind: staging-test`, the published spec page as the link (its Features
   hold the test scripts) and `questions: 0`. Commit and push.
3. PushNotification, kept short: "needs you: phase NNN is ready for your test — N steps". The
   in-session message lists each bundle feature's numbered steps inline (do → expect), then the
   link, and says where to test: the URL, the test account for each role (names, never passwords),
   Expo Go. With a `Pal Center version:` line of 0.4.0 or later, it adds: the steps are also there
   in Pal Center's Work tab, each feature's **Test** button under Waiting on you opens its steps.
4. Collect the results:
   - With a `Work CLI: "<path>"` line and `Pal Center version:` 0.4.0 or later, arm
     `"<path>" wait-results --phase NNN --timeout 9m` through Monitor and re-arm it on exit 3 (still
     waiting). Exit 0 prints the runs as JSON: record them in step 5. After a resume or compaction
     run `"<path>" results --phase NNN` first. Exit 5 (Pal Center not running) or no line: the
     walk-through below.
   - Otherwise walk the owner through each test script step by step through Remote Control: show
     one step (do, expect), wait for pass or fail with an optional note, then the next.
5. Record every step in the feature's `results` in `doc/features.yaml`:
   `{date, env, step, result, note, issue}`. Commit and push after each feature.

## 4. A failed step: fix now or defer

Ask, with a recommended answer and its reason:

- **fix now** — a session fixes it now:
  1. The feature becomes `fixing`.
  2. In each repo involved, the fix branches from staging: `fix/NNN-<short>`, a worktree in the
     set from `origin/staging` (run `git -C <set>/<component> fetch origin` first). The next phase's work on `dev` never reaches staging this way.
  3. Test first, then the fix; one review (`pal-review-code` and `pal-review run --kind code`;
     exit codes, the version rule, severity and 3 fix rounds as in `references/review.md`).
  4. Its PR into `staging` merges on its own after the review and green CI, with
     `gh pr merge <n> -R <owner/repo> --merge` (never squash); staging (or the local
     stack) redeploys. Then PR the same branch into `dev` too, merged with `--merge`.
  5. The step reopens: the feature becomes `retest`, and the owner runs that step again.
- **defer** — only a backlog issue is filed:
  1. `gh issue create -R <owner/repo>` with label `bug` and a body starting
     `Feature: F-012. Why: step N failed on <env>: <note>`; add it to the backlog board.
  2. Put the issue (`owner/repo#n`) in that result's `issue`. The feature becomes `deferred`: not
     done. `/pm` sets it `retest` when the issue closes.

## 5. Anything else the owner raises

Something that is not a failed step ("make that button bigger") gets the same choice:

- **fix now** when it is small and within this phase's features, handled like a failed step;
- otherwise **defer**: an issue labelled `feedback` (the owner's own feedback) on the backlog, with
  the feature it concerns.

## 6. When every step is answered

1. Each bundle feature is `passed` when every step passed, `deferred` when any step was deferred.
2. `gates.staging_test_done` the date; clear `waiting_on_owner`; `stage: shipping`. Commit and push;
   compact (SKILL.md, **Owner stops** step 6).
3. Go to `.pal-harness/skills/phase/references/ship.md`.
