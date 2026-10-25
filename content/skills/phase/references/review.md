# Stage: review (`code-review`)

One whole-branch review over every `feature/NNN-slug` branch (base `origin/dev`) and the
`docs/NNN-slug` branch, then the test reports from the final commit.

## 1. The test reports

Make them once, after the last review fix (section 4 step 4), on the final commit: not before the
fixes, and not per reviewer. `pal-review-tests` then checks them (`testing.md` has the files, the
thresholds and the tools).

1. Until CI has report jobs, run the project's one report command (`make reports`, or
   `npm run reports` in a Node-only repo) once, locally, on each branch's head. It writes them to
   `_test_reports/<component>/<YYYY-MM-DD-HH>-<NNN-slug>/` on `docs/NNN-slug`, with a numbered
   `summary.md` from `.pal-harness/templates/test-summary.md`.
2. Every phase: `unit_test.html`, `test_coverage.html`, `profiling.html`, `summary.md`. Regression:
   here with the others until the project has staging; once it has staging, regression (and the
   load test when `load_test: true`) runs on staging after the sign-off PR merges
   (`.pal-harness/skills/phase/references/staging-test.md`).
3. A report that does not apply is "n/a" with the reason; one with no tooling yet is a recorded
   skip, as the plan says.
4. Blocking: coverage under 80%; a failing test; an API call's p95 more than 20% slower, or a web
   page's Lighthouse performance score 5 points or more lower, than the previous phase's report
   (the first phase sets the baseline). Fix it, up to 3 rounds, then ask the owner ("needs you").
   After each fix, make the reports again on the new final commit.
5. Under `ci-test-summary` (SKILL.md, **Flow overrides**) make no report files. Once CI is green
   on the final commit, take its test summary (tests passed and failed, and coverage, per
   component; the contract drift check) from the run (`gh run view <id> -R <owner/repo>`) and save
   it as `specs/NNN-slug/reviews/ci-summary-YYYY-MM-DD-HHmm.md`. The tests reviewer checks that
   summary instead of the report files. Coverage under 80% and a failing test still block.

## 2. The reviewers

Run them in parallel.

1. **Codex** (`pal-review` by the path in the session-start context, Bash timeout 600000 ms, from the
   set's root, `--repo <component path>` per repo): `pal-review run --kind code` every phase;
   `pal-review run --kind uiux` only when the phase changed something on screen. Exit codes:
   - 0: the report is printed. Save it as `specs/NNN-slug/reviews/codex-<kind>-YYYY-MM-DD-HHmm.md`;
     add 1 to `codex.code` or `codex.uiux`.
   - 3: still running. `pal-review wait <id>` (same Bash timeout) until it ends.
   - 4: failed. Record it in the log, retry once; failed again → add `<kind>: failed (<reason>)` to
     `codex.skipped` and go on with a note for the sign-off PR.
   - 5: skipped. Add `<kind>: skipped (<reason>)` to `codex.skipped`.

   Pal Center 0.4.0 or later (the `Pal Center version:` line) resolves a worktree set to its project:
   run from the set's root. With Pal Center older than 0.4.0, `pal-review` cannot run from the set:
   add `<kind>: not available (worktree set)` to `codex.skipped` and go on. Never run it from the
   owner's checkout for a phase; it would review the owner's code, not the phase's.
2. **Claude's reviewers**, each a subagent by its agent name, given the spec, the plan, the branch
   list with base `origin/dev`, and, once the reports exist (section 4 step 4), the report folder
   (under `ci-test-summary`, the CI summary file, with a line saying it replaces `_test_reports/`).
   Before that, tell `pal-review-tests` "reports not made yet: skip the report checks (3–7)":

   | Agent | When |
   |---|---|
   | `pal-review-code` | Every phase |
   | `pal-review-tests` | Every phase |
   | `pal-review-errors` | Every phase |
   | `pal-review-security` | When the phase touches login, user data, payments or credentials |
   | `pal-review-contract` | When the API or the infrastructure changed |
   | `pal-review-ui` | When screens changed |

3. The reviewers are read-only: save each returned report verbatim as
   `specs/NNN-slug/reviews/<agent>-YYYY-MM-DD-HHmm.md`.

## 3. Severity

The severity table in `workflow.md` applies to every finding, Codex's and Claude's alike:

- **none:** nothing to do.
- **minor:** Claude decides: fix it, or dismiss it with a one-line reason in the report.
- **major** and **critical:** one numbered batch to the owner (an owner stop, `kind: findings`,
  written to `reviews/owner-YYYY-MM-DD-HHmm.md`), critical first, each with its reason, impact and
  a recommended action, so "accept all" works.

Arguing a finding up or down a level is allowed; a downgrade from major or critical is written in the
report with the reason, and the owner still sees it.

## 4. Fix rounds

1. A round fixes every open finding, then re-runs only the reviewers that raised them.
2. This whole-branch review gets 3 rounds; the per-task reviews inside a larger build get 5 rounds.
3. Still unfixed after the limit → "needs you" with what was tried (`kind: findings`).
4. After the last code change, make the reports (section 1; under `ci-test-summary`, take the CI
   summary), so every report comes from the final commit (single-repo: the last commit that changes
   code; SKILL.md, **Single-repo projects**). Then run `pal-review-tests` on them.
   A missing, empty or stale report is a finding.

## 5. Leftover findings

Each finding not fixed goes to one place, recorded for the sign-off PR's review table:

1. **Backlog**, only when it is a real bug, or tied to a `need` feature: an issue in the repo that
   owns it, body starting `Feature: F-012. Why: …`, added to the backlog board.
2. Otherwise one line in `doc/later.md` on `docs/NNN-slug`:
   `- YYYY-MM-DD — <finding> (source: <reviewer>, phase NNN)`.

## 6. Merge the component PRs

1. Each `feature/NNN-slug` PR body carries the Codex report summary: the verdict, every major and
   critical finding with the owner's decision, the minor ones with what Claude did (`workflow.md`).
2. Each PR into `dev` merges on its own once CI is green and the reviews pass:
   `gh pr checks <n> -R <owner/repo> --watch`, then `gh pr merge <n> -R <owner/repo> --merge` (a merge commit, never squash). Do not pass
   `--delete-branch` from inside a worktree; branches go when the set is removed.
3. Single-repo: the code PR `docs/NNN-slug` → `dev` merges here the same way, with the reviews and
   the reports on it. Keep the branch.
4. `stage: signoff`. Commit and push `docs/NNN-slug`.
5. Go to `.pal-harness/skills/phase/references/signoff.md`.
