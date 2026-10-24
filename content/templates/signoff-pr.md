[HUMAN GATE] Phase <NNN> — <Title>

(With the `auto-promote` override, drop the `[HUMAN GATE] ` prefix from the title above, and this
PR merges on its own once CI and the reviews pass.)

> Reads in about 10 minutes. Merging this PR is your approval for the whole phase.

## Summary

1. (What was built, in one sentence.)
2. (Features covered, by id and name.)
3. (Which of your answers shaped it.)
4. (Review and test verdict in one line: clean, or what is left.)
5. (What you do next: merge, then run the test scripts on staging. With `auto-promote`: nothing,
   it merges on its own. With `no-owner-test`: read the What to expect now list; there is no test.)
6. Plan: [plan.md](link)

## Calls I made

1. (A small call made without asking, and why.)

## Review results

| Reviewer | Found | Fixed | Left over, and where it went |
|---|---|---|---|
| (name) | 0 | 0 | none |

## Test reports

1. Unit tests: (passed / total). [report](link)
2. Coverage: (percent), minimum 80. [report](link)
3. Speed: (slowest call, change since the last report). [report](link)
4. Regression: (result). [report](link)

(With the `ci-test-summary` override, replace the list with the final commit's CI run: tests passed
and failed and coverage per component, the contract drift check, and the run's link.)

## When you merge

1. Staging is built fresh, with test accounts and sample data (never production data).
2. Regression runs (and the load test, if you asked for it).
3. You get a "ready for your test" message.

(With `auto-promote` the PR merges on its own once CI passes. With `no-owner-test` there is no staging
test: say what ships next and where the What to expect now list is.)

## Your test scripts

1. F-000 — (name): [test script](link)

(With the `no-owner-test` override, link each feature's What to expect now list instead.)
