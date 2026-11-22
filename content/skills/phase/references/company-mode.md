# Company mode and the veto window

Pal Center's Company mode lets a spec-review or findings stop pass on the owner's veto window
instead of a reply in the session. This file says how a session reads the mode and acts on a
decision. It applies to `/pm`, `/pm pickup`, every phase and every quick fix. The `Work CLI:` line
gives the path of `pal-work`.

## 1. Mode first

1. Before anything else, every `/pm`, `/pm pickup`, phase start, phase resume and quick-fix start
   runs `pal-work mode --project <key>` (`<key>` is `project.yaml` `key`) and keeps its output for
   the session.
2. The output is the only truth about Company mode. Never trust a note, a memory, a spec or a seed
   that says the mode is on or off.
3. Exit 5, no `Work CLI:` line, or a `Pal Center version:` older than 0.21.0: the mode is unknown.
   Every stop then waits for the owner, as written in SKILL.md, **Owner stops**.

## 2. At a spec-review or findings stop

Do the usual stop steps first (SKILL.md, **Owner stops** steps 1–4: publish, set
`waiting_on_owner`, push, notify). Then follow the last line of the `mode` output:

- `owner stops: run wait-decision --phase NNN …`: run `pal-work wait-decision --phase NNN` in the
  foreground with a Bash timeout of 600000 ms. On exit 3 (still waiting) run it again. Then act on
  what it prints:

  | Result | What the session does |
  |---|---|
  | `approved` or `defaulted` | Use every recommended answer in the published page |
  | `changed: <text>` | Use the owner's text; a recommended answer it does not change stands |
  | `vetoed` or `none` | Wait for the owner in the session, as **Owner stops** step 4 says |
  | exit 4 (failed) or 5 (not running) | Wait for the owner in the session |

- `owner stops: wait for the owner`: wait as **Owner stops** says.

The sign-off PR and the staging test are always the owner's: they never pass on a decision, in any
mode. Other stops (`questions`, a stuck build) also wait for the owner.

## 3. Never by default

A stop whose questions involve money, a password or key, or anything public (a store listing, a
public post, a domain, anything a stranger could see) is held before waiting:
`pal-work hold --phase NNN --reason "<why, in one line>"`. A held decision never passes by default;
the session waits for the owner's own answer.

## 4. Recording

1. Each answer goes in `doc/decisions.md` as usual: `N. YYYY-MM-DD (phase NNN): <question> → <answer>`.
2. An answer that came from the veto window's default ends with `(used by default)`. An answer from
   `changed:` quotes the owner's text.
3. **The trial measure.** Every answered owner stop, in any mode, adds one entry to `owner_stops` in
   `phase-state.yaml`, in the commit that clears `waiting_on_owner`:
   `- {kind: <waiting kind>, since: <its since>, answered: <RFC 3339 now>, could_decide: true|false, via: <how it ended>}`.
   - `could_decide` is `true` for a spec-review or findings stop with recommended answers and no
     money, key or public question (section 3), whether Company mode was on or not.
   - `via:` is the `wait-decision` result (`approved`, `defaulted`, `changed`, `vetoed`) when one came;
     otherwise `owner`. A `via` other than `owner` needs `could_decide: true`.
   - Leave the entry out when the **Reader check** (SKILL.md, **Until Pal Center and CI catch up**)
     warned that the installed Pal Center cannot read `owner_stops`; the stop then stays in the log
     only.
4. Clear `waiting_on_owner`, update `last_progress_at`, commit and push, then set the work state
   back to `working` (**Owner stops** step 6).

## 5. Resume

A resumed session whose seed says a decision ended (approved, defaulted, changed or vetoed) runs
`wait-decision` first (`pal-work wait-decision --phase NNN`): it returns the ended decision at once. Act on it as section 2
says before anything else in **Resume**.

## 6. The bundle line

Whenever `/pm` recommends work or a pickup claims a card, the lead Roadmap card's first line is
`Bundle: … · vehicle: … — <reason>` (pm skill, **Bundle and vehicle** step 3). Pal Center reads the
vehicle from it, so the line is written in the same write as the move, never later.
