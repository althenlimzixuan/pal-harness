# <NNN> — <Title>: plan

- **Status:** Draft
- **Spec:** `spec.md` (approved on YYYY-MM-DD)

> Reads in about 10 minutes. Building starts only after you approve this plan.

## Summary

1. (The tasks, one line each: "Task 1 — …".)
2. (Repos touched.)
3. (Risks and choices flagged for you.)
4. (What you can test at the end.)
5. (Build approach and its reason: see below.)
6. (Reports this phase will produce: see below.)

## Build approach

**inline** or **subagent** — (the reason in one line: size, number of repos, API change.)

Reply "approved" to start. Add "include load test" to the reply to add a load test on staging.

## Reports

- Every phase: unit tests, coverage, regression, profiling, and a numbered summary.
- Load test: only when you say "include load test". Included: yes / no.
- Not applicable here: (report and reason, or "none".)

## Tasks

One block per task.

### Task 1 — (name)

- **Repo and files:** (repo; files to create or change)
- **Test first:** (test file) with fixture `tests/(test_type)/(use_case).json`
- **Code:** (what makes the test pass, in a line or two)
- **Prove it:** (the exact commands, and the result to expect)
- **Commit:** (the commit message)
