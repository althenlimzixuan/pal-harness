# Task K — <name>: report

<!-- The one report an implementer or fixer sends (`references/subagents.md` section 1). Write it
once, in this order, and send it once. At most 60 lines: save long output (a full test log) to a file
in the task's worktree and name it here. It ends with the line `End of report.`, and nothing follows it. -->

**Gate 1 — test first**

- Test commit: `<hash>` `Add failing test for <behaviour>`. It holds only tests and fixtures, and it
  was made before any code commit of this task (only a Carry-over commit may come earlier).
- RED: `<command>`, exit <N>. The lines that name each failing test:
  - `<test name>`: `<the failing line>`
- Guard tests: `RED by mutation: <test>` (the break, the failing line, restored), or `none`.

**Gate 2 — green and lint**

- GREEN: `<command>`, exit 0.
- Lint: `<the repo's lint script>`, exit 0.

**Gate 3 — full test run**

- `Full test run: <command>, foreground, exit <N>`. A timeout is a failure, never a pass.

**Checks**

- Known findings: checked (`references/subagents.md` section 1).
- Carry-over: fixed in `<hash>` `Fix Task K review minors`, or `none`.
- QA checks: `<command>`, exit <N>, or `QA checks: none named`.
- Background work: nothing I started is still running.

**Changes**

- `<hash>` `<message>`: `<files>`

**New question** (only a fixer whose finding cannot be fixed without more: what it would change and why)

End of report.
