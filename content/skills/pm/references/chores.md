# Chores

A chore is repeatable work with no spec: a dependency bump, a check every retro asks for, a weekly
clean-up. Chores live on the Roadmap board as cards; Pal Center's own pickup runs them first, and a
hand-run `/pm pickup` does the same.

## 1. When a chore is due

A chore runs from the Roadmap board's `Chores (Active)` column. Its card's body holds a timing line
and, after each run, a `Last run: YYYY-MM-DD` line.

| Timing line | Due when |
|---|---|
| `Every: day`, `Every: week` or `Every: month` | That long has passed since the card's `Last run: YYYY-MM-DD` line (no line: due now) |
| `When: before phase` | First thing in a session that starts a phase card |
| `When: after phase` | After a phase reaches `done` |
| No timing line | Once; after the run the card moves back to `Chores` |

- A chore may run beside a phase, in the other session folder (`<KEY>-1` or `<KEY>-2`). It never
  touches the phase's branches.
- A card with `Action: harness notes` is Pal Center's alone (its harness-notes batch): the skill
  leaves it, due or not.
- Cards in the `Chores` column (not Active) never run.

## 2. Running a chore

The session follows the card's description, then ends with a final review:

1. **Code changes** run as a quick fix (`.pal-harness/skills/quick-fix/SKILL.md`): the one review,
   the PR into `dev`, and the merge on green. A chore needs no owner test step unless it changes
   something the owner sees.
2. **A check-only chore** files what it finds as backlog issues (`pm` skill, **Reviewer findings**
   step 1) and writes a short report as a comment on its card: what it checked, what it found, the
   issue links.
3. Then the session writes `Last run: YYYY-MM-DD` on the card (replacing the old line), with
   `gh project item-edit --id <draft-id> --body "<text>"`. It never edits the board's columns or
   their options; only the once-only card moves back to `Chores`.

## 3. Registering a chore

1. When the owner asks for repeatable work, or a session sees work that repeats (the same fix three
   phases running, or a check every retro asks for), the session proposes a card in one message:
   - the title;
   - the `Every:` or `When:` line;
   - a description to start from (what to run, what counts as done).
2. It creates the card only after the owner's yes in the session, in the column the owner names
   (`Chores (Active)` by default):
   `gh project item-create <n> --owner <o> --title "<title>" --body "<timing line>\n\n<description>"`,
   then moves it to that column.
3. A pickup or `/pm` run never creates a chore on its own, and never registers a card with
   `Action: harness notes`.
