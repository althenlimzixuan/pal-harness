# Stage: ship (`shipping`, then `done`)

After the owner's staging test is passed or deferred (under `no-owner-test`, straight after
sign-off): promote to `main`, deploy, tear staging down, write the retro, close out. SKILL.md's
**Flow overrides** change sections 1, 2 and 4 as marked.

## 1. Promote to `main`

1. In each component repo the phase touched, and in the SSOT repo (single-repo: the SSOT repo
   only):
   `gh pr create -R <owner/repo> --base main --head staging --title "Promote staging to main (phase NNN)"`,
   watch CI, merge with a merge commit (`gh pr merge <n> -R <owner/repo> --merge`). It merges on its
   own until the product is in production.
2. A project in production has the owner's gate on `main` (its Workflow Overrides): title these PRs
   `[HUMAN GATE] …` and wait for the owner as an owner stop (`kind: signoff`); never merge them
   yourself. Under `auto-promote`, `staging` → `main` merges on its own once CI passes, with no
   `[HUMAN GATE]` prefix.
3. Already promoted (a "test on production" project): skip this section.

## 2. Deploy and smoke check

1. Production deploys only from `main`, by the project's own pipeline. Watch it to the end.
2. Run the smoke check: the project's runbook check if `doc/runbooks/` has one, otherwise the health
   endpoint and the main journey's first screen. Write the result in the log.
3. A failed deploy or smoke check: "needs you" with what failed (`kind: questions`). Do not start
   another phase on top of it.
4. No production environment yet: write "no production deploy yet" in the log and go on.
5. **Under `installer-ready`** steps 1–4 do not apply (there is no deploy); instead:
   1. Once `main`'s CI is green, tag `vX.Y.Z` (the version the phase set) on the merge commit and
      push the tag (tagging is pre-authorized). Wait for the release workflow to publish the
      release (`gh run watch <id> -R <owner/repo>`, then `gh release view vX.Y.Z -R <owner/repo>`).
   2. Download its installer, through Bash:
      `gh release download --pattern '*-setup.exe' --dir "$USERPROFILE/Downloads" --clobber vX.Y.Z -R <owner/repo>`
      (in PowerShell the folder is `"$env:USERPROFILE\Downloads"`; never a bare `/Downloads`).
      Check the file exists at that path before giving it to the owner.
   3. A failed release workflow or download: "needs you" with what failed (`kind: questions`),
      as section 2 step 3 does for a deploy.
   4. Never stop or start Pal Center, and never run the installer: the owner installs when it
      suits them.
   5. The release notes and the session's final message carry the What to expect now list, the
      release link and the installer's full path (`%USERPROFILE%\Downloads\<file name>`). Add the
      list to the notes with `gh release edit vX.Y.Z -R <owner/repo> --notes-file <file>`, the file
      holding the release's notes plus the list.

## 3. Tear staging down

1. Staging is always torn down after the test (the project's teardown; on the local stack, stop it
   and remove its containers and seed data).
2. A failed teardown is a "stuck" alert: send "needs you" with the error (Pal Center alerts on it
   too, once it exists). Staging also tears itself down after 48 hours idle; never rely on that.
3. Under `no-owner-test`, skip this section when nothing was started for a test.

## 4. Statuses and cards

1. The bundle features keep their final status (`passed` or `deferred`); move each card to its
   column (pm skill mapping). Under `no-owner-test` they become `shipped`, never `passed`.
2. Close each bundle issue the phase fixed and the owner passed (or shipped, under
   `no-owner-test`), if a PR did not close it:
   `gh issue close <n> -R <owner/repo> --comment "Shipped in phase NNN"`.

## 5. Retro

1. Dispatch the `pal-retro` agent with the phase folder; the session files nothing itself and
   reports what the agent filed. The agent writes `specs/NNN-slug/retro.md` from
   `.pal-harness/templates/retro.md`: what went well (keep), what went badly (fix), and the ideas.
   Keep it short.
2. File each idea as a harness capture in Pal Center's inbox with the session-start line
   `Capture CLI: "<path>"`: `"<path>" --kind harness --project <key> "<idea>"` (exit 4 rejected, 5
   Pal Center not running). With no line or exit 5, write the ideas to
   `_feedbacks/YYYY-MM-DD-HHmm.md` in the SSOT repo, one line each, marked `harness`.
3. The good points stay in `retro.md`.

## 6. Close out

1. `stage: done`; clear `waiting_on_owner`; update `last_progress_at`.
2. Open the close-out PR (`gh pr create -R <owner/repo> --base dev --head docs/NNN-slug`) from
   `docs/NNN-slug` into `dev` (the branch's second PR): the retro, the
   results and statuses in `doc/features.yaml`, the final `phase-state.yaml`, any `_feedbacks/`
   note. It merges on its own once CI is green (`gh pr merge <n> -R <owner/repo> --merge`). Pal Center reads
   the phase from `dev` from now on.
3. Once deployed (or, with no production yet, once staging → main merges; under `installer-ready`,
   once the installer is downloaded) and the close-out PR merged, write **recommended next**: run the
   pm skill's **Recommended next** (`.pal-harness/skills/pm/SKILL.md`), which writes through its own
   worktree.
4. PushNotification: "done: phase NNN shipped — <one line>. Retest: N. Deferred: N." Under
   `installer-ready`, add the installer's full path.

## 7. Clean up

1. In every worktree of the set: `git status` clean and `git log '@{u}..'` empty (run git through Bash; on a detached HEAD use
   `git -C <path> log origin/<branch>..<branch>`). Anything unpushed →
   push it or say what it is; never remove a worktree that holds unpushed work.
2. `git -C <project root>/<component> worktree remove <set>/<component>` for each component, then
   `git -C <project root> worktree remove <set>`. Delete the phase's merged local branches.
