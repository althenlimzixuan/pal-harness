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
are none. Every value you are asked for gives its exact format and an example: an IAM member is
`user:<email>` (`user:owner@example.com`), a project is its ID, not its name (`lfpv-prod-123`), a
region is its code (`asia-southeast1`).

1. (where to run it) `(the exact command)` (what it does, in one line)

A Terraform step lists, in its folder and in order: `gcloud auth application-default login` (once
per PC, or when it has expired); `terraform init -backend-config=backend.hcl` (or the project's
backend file) when the folder has no `.terraform/` yet; the `plan` and the `apply` with the exact
`-var-file`; and the GCP project ID they act on. Every `gcloud` command passes `--project <id>`.

## Tasks

One block per task.

### Task 1 — (name)

- **Repo and files:** (repo; files to create or change)
- **Test first:** (test file) with fixture `tests/(test_type)/(use_case).json`. RED evidence: the failing
  test names, from the run's own output (the command and the lines that name each test), never a count
- **Guard tests:** (optional: tests that pin behaviour already there and pass on their first run, each
  with its reason; each one's RED is a mutation run: break the code it guards, see it fail, restore)
- **Code:** (what makes the test pass, in a line or two. A screen with several independent save
  buttons, such as switches or forms, says "saves per item": each save sends only its own item, and a
  late answer never overwrites another item's value or a newer value of its own)
- **Values from data:** (a task that builds text from data, such as a message, a label, a path or a
  line naming a repo: each value it puts in, with its source field and type, for example
  "repo: `project.yaml components[].repo`, string `owner/name`"; delete this line otherwise)
- **Prove it:** (the exact commands, the repo's lint script among them, and the result to expect.
  Code for one OS adds the other OSes' checks, such as `GOOS=linux go vet ./...`,
  `GOOS=darwin go vet ./...` or `cargo clippy --target x86_64-unknown-linux-gnu`. A task that
  changes QA sample data or a screen adds the project's QA checks command, from its `CLAUDE.md`
  line `QA checks: <command>`, expected green)
- **Render check:** (UI tasks whose spec gives pixel values, such as a grid size, a row height or a
  width: the values; the sample data that fills them (the longest real label, the most rows); and
  the screenshot `specs/NNN-slug/screens/task-K-<name>.png`, taken once with that data and looked at
  before the task is done: nothing overlaps, is clipped or is cut off)
- **Existing behaviour:** (restyle tasks only: each badge, figure, link, control and state the
  screen shows today, and the test that covers it)
- **Native versions:** (a task that adds or upgrades a native dependency: the resolved versions of
  the shared libraries it pulls in, against the existing native modules)

A phase that changes the mobile app has a task that bumps the app version (the runbook's release
step).
- **Commit:** (the commit message)
