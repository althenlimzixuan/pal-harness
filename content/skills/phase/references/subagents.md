# Subagents in a larger build

A larger phase builds with `superpowers:subagent-driven-development`: one fresh agent per task, a
review after each task (`references/build.md`, section 3). This file is what the session, as the
controller, tells and checks the agents.

## 1. The brief

Hand each fresh agent its task text, its worktree path, and the paths of `testing.md`, `workflow.md`
and the stack practice. The brief says TDD is required for the task: run the fixture's test before
any code, and report, for each new test, the RED output by name (the command and the failing lines
that name the test; one run may show several), then the GREEN output and a clean lint run. A test the
plan lists under **Guard tests:** pins behaviour already there and passes on its first run; its RED
is a mutation run (break the code it guards, run it, see it fail, restore), reported as
`RED by mutation: <test>`. Local Go test
runs set `GOFLAGS=-p=1 GOMEMLIMIT=2GiB` (`testing.md`, Go-Specific). The brief also tells
implementers:

1. to never wait on their own background work: every test, lint and build runs in the foreground
   (`devops_cicd.md`, **Free memory before heavy runs**), and a report is sent only when it ends;
2. to read the full test run's own exit code, and to report a timeout as a failure, never as a pass
   (a run that timed out hid a broken test in Pal Center phase 015);
3. to check the **Known findings** list below before reporting, and say in the report that it did;
4. before the final report, it finishes or stops every background command, Monitor and watcher it
   started (TaskStop for its own tasks), and says in the report that none is still running. A
   watcher left running keeps re-sending reports after the task is done;
5. to fix the brief's **Carry-over** list first, when it has one (section 2, step 4);
6. for a fixer (a task's fix rounds, the final fixer, the review's fix rounds): a fix changes
   behaviour the owner can see or reach (text, controls, defaults, what an answer or a click does)
   only as far as its finding says. When the finding cannot be fixed without more, leave that part
   and report it under **New question**: what it would change and why (`references/review.md`,
   section 4).

**Known findings.** Each one cost a fix round in more than one phase:

- no lock held across a network call, a process start or a wait on another lock;
- a retry re-reads its rule and its target before acting again, never the values it read before the
  failure;
- a missing source (a file, a service, a setting) shows as a visible state with its reason, never a
  silent pause or an empty list;
- `openapi` required fields match what the handler always sends, and what it validates as required.

## 2. The controller

1. **A repeating agent.** Stop an implementer with TaskStop the second time it repeats a finished
   report, or reports "waiting on background work". Give the task to a fresh agent with the last
   good report and what is left; that counts as one of the task's fix rounds.
2. **A stale reply.** Accept a reply only when it answers the latest fix message: the reply quotes
   that message's first line. A reply that does not is from an older message: send it back once
   with the latest message, and never act on it.
3. **The red run, by name.** List every new test by name from the task's diff (Go `func Test…`,
   Jest and Playwright `it(` / `test(`, Rust `#[test] fn …`) and find a RED line for each name in the
   report. A name with no RED line goes back to the implementer before the task reviewer is
   dispatched. Now that the code exists, its RED can only come from a mutation run, logged
   `RED by mutation: <test>`; that is not a fix round and costs no review round. A test that stays
   green under the mutation is rewritten, which is a fix round. The per-task review needs a clean
   lint run.
4. **Carry-over.** After task K's review, each minor the reviewer marked should-fix that its own
   rounds did not fix goes on a **Carry-over** list in task K+1's brief. Task K+1's implementer fixes
   them first, in their own commit (`Fix Task K review minors`), and task K+1's review checks them; it
   is not a fix round. The last task's carry-overs go to one fixer before the whole-branch review.
   This keeps the fix wave before the PR small.
5. **Tag deferred minors at triage.** Write each minor the task's review leaves open into sdd's
   ledger, at triage, with one tag: `Task K: minor (deferred) [carry]: <one line>`. The tags are
   `carry`, `final`, `later`, `issue`: `carry` for the next task's Carry-over (step 4), `final` for
   the one fixer before the whole-branch review, `later` for a `doc/later.md` line, `issue` for a real
   bug that becomes a backlog issue. Write `[fixed]` over the tag once it is fixed. Build task K+1's
   Carry-over list with `pal-harness minors <ledger> --task K --tag carry`. sdd's own
   `Task N: parked — … — Ruling: …` lines stay as they are.
6. **A finished task.** A report from a task already marked done is ignored: not acted on and not
   answered. A second one from the same agent gets TaskStop.

## 3. The task reviewer's checks

**The task reviewer's checks.** The dispatch to sdd's task reviewer adds:

1. match every new test by name to its RED line before reviewing any code; reject the task when the
   report lacks a RED run for any new test, or the RED came after the code without
   `RED by mutation: <test>` (a fix round);
2. for a screen, **Warning colour only where the user must act:** and **Window edges tested:**
   (`pal-review-ui` checks 8 and 9);
3. for a restyle, every item in its **Existing behaviour** list (`references/plan.md`, section 1) is
   still on the screen and its test passes;
4. owner-facing wording: every text the owner reads (labels, messages, notices, reasons) has no
   jargon, follows the order the owner sees things in, and uses the reason names the docs use;
5. the **Known findings** list (section 1) over the task's diff. The first task that touches shared
   state also gets `pal-review-code`'s concurrency and lifecycle checks in its own review;
6. the task's **Carry-over** list (section 2, step 4) is fixed;
7. **Every ending tested.** For a task that sorts runs, sessions or jobs into outcomes (a status, a
   result, an end state): list every way one can end (start failed, owner stop, restart mid-run,
   idle, waiting, cancelled, timed out, crashed or killed, plus any ending the feature defines
   itself) and match each to a test that drives it and checks the outcome. An ending with no test is
   a major finding;
8. **Mutation check.** For each new test (at most 5 per task, the tests of the task's main
   behaviours), make one small break in the code it covers (flip a condition, drop a call, return
   early), run that test and confirm it fails. Then restore every file you changed
   (`git restore -- <files>`) and check `git status --porcelain` shows nothing the review did not
   start with. A test that stays green is an Important finding: "`<test>` passes with `<break>`".
   The report names each break and its result;
9. a fixer's change to behaviour the owner sees goes no further than its finding (section 1, item 6).

## 4. Stale editor warnings

Language-server diagnostics shown after an edit can be stale. They are ignored when `go build`,
`go vet` and the tests pass (or the stack's own typecheck and tests); the command results win. A
diagnostic is acted on only when a command reproduces it.
