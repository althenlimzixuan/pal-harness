# <NNN> — <Title>: plan

- **Status:** Draft
- **Spec:** `spec.md` (approved on YYYY-MM-DD)

> Reads in about 10 minutes. There is no plan stop: building starts once Claude's plan review passes.
> Over 8 tasks (or about 30 KB): this file holds the summary and Part 1's tasks (up to 8), and each
> later part goes in `plan-2.md`, `plan-3.md` …, reviewed one part at a time.
> Written in plain English for you: no jargon, and a diagram's labels are words you would use.

## Summary

1. (The tasks, one line each: "Task 1 — …".)
2. (Repos touched.)
3. (Risks and choices flagged for you.)
4. (What you can test at the end.)
5. (Build approach and its reason: see below.)
6. (Reports this phase will produce: see below.)

## Build approach

**inline** or **subagent** — (the reason in one line: size, number of repos, API change.)

## Reports

- Every phase: unit tests, coverage, regression, profiling, and a numbered summary.
- Load test: only when you said "include load test" in the spec reply. Included: yes / no.
- Not applicable here: (report and reason, or "none".)

## Gate dry-runs

Each gate (a grep, lint rule or script that decides a task is done), run before the plan review.
"none" when no task has one.

| Gate (task) | On the current tree | On a planted bad sample |
|---|---|---|
| (command) (Task N) | (N matches, all real hits; or the false hit's file:line) | fails, as it should |

## Deletes and locks

"none" when no task deletes anything or fires a callback under a lock. A helper that deletes files
or folders lists its link and junction test cases.

| Delete helper (task) | What it deletes | The read-only check first (proves the target is its own) | Link and junction test cases |
|---|---|---|---|
| (helper) (Task N) | (files, folders, rows or keys) | (a marker it wrote, a path under its own root, a row it owns) | (a link inside pointing out; the target itself a link) |

| Callback under a lock (task) | Lock held | What the callback may take (never the same lock) |
|---|---|---|
| (callback, event or hook) (Task N) | (lock) | (locks or calls it makes) |

## Owner-run commands

Commands only you can run, with their exact text, in the order ship needs them. "none" when there
are none.

1. (where to run it) `(the exact command)` (what it does, in one line)

## Tasks

One block per task.

### Task 1 — (name)

- **Repo and files:** (repo; files to create or change)
- **Test first:** (test file) with fixture `tests/(test_type)/(use_case).json`
- **Code:** (what makes the test pass, in a line or two)
- **Prove it:** (the exact commands, the repo's lint script among them, and the result to expect)
- **Existing behaviour:** (restyle tasks only: each badge, figure, link, control and state the
  screen shows today, and the test that covers it)
- **Native versions:** (a task that adds or upgrades a native dependency: the resolved versions of
  the shared libraries it pulls in, against the existing native modules)

A phase that changes the mobile app has a task that bumps the app version (the runbook's release
step).
- **Commit:** (the commit message)
