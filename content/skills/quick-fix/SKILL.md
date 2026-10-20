---
name: quick-fix
description: Use to run a card whose vehicle is quick-fix — a small change with no spec or plan, built in the session, one review, one test step for the owner, then sign-off, the owner's test and ship as in the phase skill. Use when /pm routes a quick-fix card or its pickup takes a safe backlog item. Run from the project's SSOT root.
---

# /quick-fix — a small fix, still reviewed and tested

A quick fix skips the spec and the plan. It is still reviewed, and it adds one test step for the
owner. Everything else (saved state, the worktree set, owner stops, the interim rules) works as in
`.pal-harness/skills/phase/SKILL.md`; read that file's **Saved state**, **The worktree set** and
**Owner stops** sections first.

## When it fits

`/pm`'s bar: a small change in one repo, no API change, no product decision. If, while working, it
needs an API change, a product decision, a second repo or more than a few tasks, stop: tell the
owner (an owner stop, `kind: questions`) and recommend running it as a phase.

## 1. Start

1. Create the worktree set and the phase folder exactly as the phase skill's **Start** steps 1–2
   (`pal-harness spec new <slug>` makes `specs/NNN-slug/` on `docs/NNN-slug`). Delete the generated
   `spec.md`: a quick fix has none.
2. Write `phase-state.yaml` from the template: `vehicle: quick-fix`, `approach: inline`,
   `stage: building`, the bundle, `session:`, `tasks: {done: 0, total: 1}`, `last_progress_at` now.
   Start `phase-log.md`. Commit and push `docs/NNN-slug`.
3. The feature the fix touches (if any) becomes `building` in `doc/features.yaml`; move its card.

## 2. The test step

Write the owner's one test step for what the fix changes (do → expect), concrete and visible:

- **The fix touches a feature:** add the step to that feature's `test_script` in
  `doc/features.yaml`, numbered after its last step. For the fix of a `deferred` feature, the failed
  step is the test step; add nothing.
- **No feature fits:** write the step in the sign-off PR's "Your test scripts", and record its result
  as a comment on the issue.

## 3. Build

1. One `fix/NNN-slug` branch in the one repo, a worktree in the set from `origin/dev`
   (as the phase's `references/build.md` section 2 shows, with `fix/` for `feature/`).
2. Test first from a JSON fixture (`testing.md`), then the code, by the build rules in
   `workflow.md`. Commit, push, update `tasks` and `last_progress_at`, push `docs/NNN-slug`.
3. Stuck (fix rounds used up, CI red, no progress for 1 hour): as the phase's `references/build.md`
   section 5.

## 4. One review

1. Make the reports as the phase's `references/review.md` section 1 says.
2. Run one review over the whole branch: `pal-review run --kind code` (and `--kind uiux` when
   something on screen changed) plus the `pal-review-code` agent; add `pal-review-security` when the
   fix touches login, user data, payments or credentials. Save each report under
   `specs/NNN-slug/reviews/`.
   - `pal-review` exit codes (Bash timeout 600000 ms): 0 the report is printed; 3 still running →
     `pal-review wait <id>` (same timeout) until it ends; 4 failed → record it, retry once, then
     add `<kind>: failed (<reason>)` to `codex.skipped` and go on with a note; 5 skipped → add
     `<kind>: skipped (<reason>)` to `codex.skipped`.
   - With Pal Center older than 0.4.0 (the `Pal Center version:` line), which cannot resolve a
     worktree set to its project, add
     `<kind>: not available (worktree set)` to `codex.skipped` and go on. Never run `pal-review`
     from the owner's checkout for a quick fix.
3. Severity, 3 fix rounds and leftover findings as in `references/review.md` sections 3–5.
4. Open the PR into `dev` with the Codex summary in its body; it merges on its own once CI is green, with `gh pr merge --merge` (never squash).
   `stage: code-review` while reviewing, then `stage: signoff`.

## 5. Sign-off, the owner's test, ship

Follow the phase skill's references in order:

1. `references/signoff.md`: the `[HUMAN GATE]` sign-off PR, shorter (the fix, its review, its test
   step).
2. `references/staging-test.md`: the owner runs the one test step; fix now or defer as there.
3. `references/ship.md`, without the retro unless something went badly; then file that one idea as
   ship.md section 5 says.

The owner stops for a quick fix are the sign-off and the test step, plus any findings or stuck stop.
