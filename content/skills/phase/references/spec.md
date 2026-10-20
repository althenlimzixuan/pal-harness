# Stage: spec (`spec`, then `spec-review`)

The spec says what this phase builds and how the owner will test it. You draft it alone; the owner
reads it once, in about 10 minutes, and answers in one line.

## 1. Read

Read-only, from the SSOT worktree:

1. `doc/north-star.md`, the lead card, and each bundle feature's `why` and `use_case` in
   `doc/features.yaml`.
2. Each bundle issue (`gh issue view <n> -R <owner/repo>`).
3. `doc/decisions.md`: past owner answers. A question answered there is never asked again; it
   becomes a call.
4. When the API may change: `doc/API.md` and `doc/openapi.yaml`. Code in the component repos as
   needed, read-only.

## 2. Draft

Write `specs/NNN-slug/spec.md` alone, in the sections `spec new` copied from
`.pal-harness/templates/spec.md`. Ask the owner nothing while drafting.

1. **Summary:** the numbered points the template lists.
2. **Questions for the owner:** only questions that change what gets built. Numbered, at the top,
   each with its options and a recommended answer with a one-line reason, so the owner can reply
   `accept all` or `accept all except 2: B`.
3. **Calls I made:** every smaller choice, one line each with its reason.
4. **Features:** one block per bundle feature with its why, use case and numbered test script
   (step: what you do → what you should see). Write steps the owner can run with the seeded test
   accounts, each with one visible result, needing no code knowledge. A bundle fix the owner can
   see gets a step in the test script of the feature it touches; a fix for a `deferred` feature is
   tested by that feature's failed step. Approving the spec approves how the owner will test.
5. **UX outline:** screens or steps in plain words; delete it when nothing on screen changes.
6. **Contract:** only when the phase changes the API; the contract comes before any code. An
   endpoint table (method and path, what it does, errors) plus one sequence diagram per call in
   `specs/NNN-slug/diagrams/<name>.mmd`, linked from the table. REST under `/api/v1/`, RFC 7807 errors. A breaking change is a numbered question
   at the top, with its impact. Delete the section otherwise.
7. **Out of scope** and the collapsed **Appendix**, as the template shows.

## 3. Size

1. About 10 minutes: about 1,500 words plus diagrams. Move detail into the collapsed appendix.
2. Still too long, or more than about one working day of building → split the phase. Keep the
   part that fits, and make the split question 1: what stays, what becomes a new feature (recommended
   answer: split). On the owner's yes, add the new feature to `doc/features.yaml` as `todo` with its
   why and use case; `/pm` puts it on the board.

## 4. Codex reviews the spec

Run it before the owner sees the spec. `pal-review` is the absolute path from the session-start
line `Codex review CLI: "<path>\pal-review.exe"`; run each command with a Bash timeout of
600000 ms.

1. From the set's root: `pal-review run --kind doc`. Exit 3 → `pal-review wait <id>` until it ends.
2. Save the report as `specs/NNN-slug/reviews/codex-doc-YYYY-MM-DD-HHmm.md`; add 1 to `codex.spec`.
3. Minor findings: fix them in the spec first. Major and critical findings join the questions for
   the owner, each with its reason, impact and a recommended answer (critical first).
4. Exit 5 (skipped): add `spec: skipped (<reason>)` to `codex.skipped` and go on.
5. No CLI path in the session-start context, or `--kind doc` rejected (Pal Center before 0.4.0):
   add `spec: not available` to `codex.skipped`, write "Codex spec review: not available" in the
   log and in the spec's Calls I made, and go on.

Codex never reviews the plan.

## 5. Send it

1. `pal-harness render specs/NNN-slug/spec.md`, then publish `spec.html` (SKILL.md, **Owner stops**).
2. `stage: spec-review`; bundle features `spec-review`; `waiting_on_owner` with `kind: spec-review`,
   the page link and the number of questions. Commit and push.
3. PushNotification: "needs you: spec NNN — <title>, N questions, about 10 minutes". Wait.

## 6. The owner's reply

1. `accept all` takes every recommended answer. `accept all except 2: B` takes the rest and changes
   one. A reply that answers every question and asks for no change approves the spec.
2. A request for changes: revise the spec, then Codex checks the revision again (section 4) unless
   the reply says `skip Codex`; then add `spec rN: owner said skip Codex` to `codex.skipped`.
   Render, publish and send again (section 5).
3. An unclear reply: ask one short follow-up in the session.

## 7. On approval

1. Append each answered question to `doc/decisions.md`:
   `N. YYYY-MM-DD (phase NNN): <question> → <answer>`. Fold the answers into the spec's sections.
2. Set the spec's `Status:` to `Approved (YYYY-MM-DD)`. Re-render `spec.html`.
3. Copy each feature's test script into its `test_script` in `doc/features.yaml`
   (`{step, do, expect}`, numbered as in the spec).
4. `gates.spec_approved` the date; clear `waiting_on_owner`; `stage: planning`; bundle features
   `planning`. Commit and push; compact (SKILL.md, **Owner stops** step 6).
5. Go to `.pal-harness/skills/phase/references/plan.md`.
