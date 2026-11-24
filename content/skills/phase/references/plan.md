# Stage: plan (`planning`)

The plan turns the approved spec into tasks. No owner stop: once Claude's plan review passes, building starts.

## 1. Write it

1. Use `superpowers:writing-plans` with the approved `spec.md`, and save into
   `specs/NNN-slug/plan.md` (never `docs/superpowers/`), in the format of
   `.pal-harness/templates/plan.md`. Without superpowers, write it by the template yourself and say
   so in the summary.
2. Every task names:
   1. its repo and the files it creates or changes;
   2. **the test to write first:** the test file and its JSON fixture
      `tests/{test_type}/{use_case}.json` (`testing.md`); its RED evidence is the failing test
      names from the run's own output, never a count;
   3. the code that makes the test pass, in a line or two;
   4. the commands that prove it, and the result to expect;
   5. the commit message.
   A task without a test file and a fixture is not finished.
3. When the API changes, follow the split rule in `product_docs.md`: the provider repo's task
   regenerates its `openapi.yaml` and serves Swagger UI; an SSOT task on `docs/NNN-slug` updates the
   `doc/openapi.yaml` copy and `doc/API.md`. Contract tasks come before the tasks that call them.
4. Cover every spec feature, every test-script step, every endpoint in the spec's `## Contract` and
   every behaviour in its Appendix with at least one task.
5. Each task's prove-it commands include the repo's lint, through its own script (`workflow.md`).
6. A task that hands over a key, a secret, a token or an invitation says who may receive it and how
   the sender verifies the recipient.
7. Under `installer-ready` (SKILL.md, **Flow overrides**), the last task bumps the version in the
   project's version files, in the phase's PR; ship tags that version.
8. A phase **touches Windows security** when it adds or changes code or tests that use ACLs,
   accounts, tokens or privileges, window stations or desktops, services or named pipes. The summary
   says so, and the last task before the PR is the session-0 run (`references/build.md`, section 6
   step 2).
9. Every task builds and tests what the spec's **Trust boundaries** table says stops misuse.
10. **Gate dry-runs.** A gate is a grep, lint rule or script in a task's prove-it commands whose
    result decides that the task is done, beyond the repo's normal lint and tests. Before the plan
    review, run each gate on the current tree and fill the plan's **Gate dry-runs** table: the gate,
    what it matches today (each match marked a real hit or a false hit), and its result on a planted
    bad sample (a scratch file with one violation, deleted after). A gate passes on code that follows
    its rule and fails on the sample. A gate with a false hit, or one that misses the sample, is fixed
    before the review. With no gate, write "none" under the heading.
11. **A restyle keeps what the screen did.** A task that restyles or re-lays-out an existing screen
    carries an **Existing behaviour** list: every badge, figure, link, control and state (empty,
    loading, error) the screen shows today, each with the test that covers it. An item with no test
    gets one in that task, written first and passing on the old code, before the restyle.
12. **Deletes and locks.** While writing the tasks, fill the plan's **Deletes and locks** section
    ("none" when no task has either):
    - each helper that deletes or cleans up files, folders, rows or keys: the task, what it deletes,
      and the **read-only check** that runs first and proves the target is one it created (a marker
      file it wrote, a path under its own root, a row it owns). A path the user chose is never deleted
      without such a check;
    - each callback, event or hook fired while a lock is held: the task, the lock, and what the
      callback may take. It must not take the same lock, or a lock taken before it elsewhere;
    - a helper that deletes files or folders lists its link and junction test cases: a symlink or
      junction inside the target that points out of it, and a target that is itself a link. The test
      proves the delete removes the link, never what it points to.
13. **Mobile phases.** A phase that changes the mobile app has a task that bumps the app version
    (the runbook's release step), whatever the flow overrides. A task that adds or upgrades a native
    dependency records the resolved versions of the shared libraries it pulls in, against the
    existing native modules (`react_native.md`, **Native dependencies**). A mismatch is fixed in the
    plan, before the build.
14. **File lists.** Each task's file list holds every file its code, tests, fixture and prove-it
    commands name.
15. **Owner-run commands.** Every command only the owner can run (a production `terraform apply`,
    `gh workflow run` on a protected workflow, `eas submit`, a store or console step, anything auto
    mode refuses) is listed under the plan's **Owner-run commands**, with its exact text and where to
    run it, in the order ship needs them, and the exact format of every value it asks for, with an
    example (`user:<email>`, a project ID, not its name). A Terraform step lists the login, `init`
    and apply in order, with the GCP project ID (`devops_cicd.md`, **gcloud and Terraform**). A first
    store release copies `.pal-harness/templates/runbook-android-release.md` (or the project's
    runbook) in a task.
16. **Look up every command and path** a task names, once, before the review: each existing file or
    folder; the folder and naming of each new file (a fixture beside the real fixtures); each make
    target, npm script and generator flag; each file the spec names as a behaviour's origin. Fix the
    task when one is not there.
17. **QA seeds and checks.** A QA seed in the spec's UX outline that the project's QA sample data
    lacks gets a task that adds it, with a test that the seeder loads it. A task that changes QA sample
    data or a screen lists the QA checks in **Prove it** (the `CLAUDE.md` line `QA checks: <command>`).
    With `harness.qa_command` set and no line, find the command (the `qa` script beside its folder),
    add the line on `docs/NNN-slug` and list it in Calls I made.
18. **Other operating systems.** A task that touches code for one OS (an OS file suffix or build tag
    in Go; `cfg(target_os …)`, `cfg(unix)` or `cfg(windows)` in Rust) adds to **Prove it** the check
    for each other OS it builds on: `GOOS=linux go vet ./...` and `GOOS=darwin go vet ./...` for Go;
    `cargo clippy --target <triple> --all-targets -- -D warnings` (after `rustup target add`) for a
    Rust crate with no system libraries; for one with them (Tauri, GTK, WebKit) the Docker run in
    `references/linux-check.md`, and `macOS proof: CI's macOS job`.
19. **Pixel values and saves.** A UI task whose spec gives pixel values (a grid size, a row height, a
    width) has a **Render check:** (the template's task block): one screenshot with sample data that
    fills them, looked at before the task is done. A screen with several independent save buttons
    says "saves per item" in its **Code:** line, and its test saves two items close together.

## 2. The summary

Numbered, at the top, as the template lists: the tasks in one line each, the repos touched, the
risks and choices flagged for the owner, what the owner can test at the end, the build approach with
its reason, and the reports this phase will produce. The whole plan reads in about 10 minutes; detail
goes in the tasks, not the summary.

1. **Build approach** (sizes in `.pal-harness/skills/phase/references/build.md`): `inline` for a small
   phase (up to ~4 tasks, one repo), `subagent` for a larger one (5 or more tasks, several repos or an
   API change). Set `approach` and `tasks.total` in `phase-state.yaml`.
2. **Reports** (`testing.md`): unit tests, coverage, regression, profiling and a numbered summary
   every phase; a load test only when the owner adds "include load test" to the spec reply. A
   report that does not apply to the stack is listed as "n/a" with the reason. A report that
   applies but has no tooling yet becomes a task in this plan when no earlier plan added it; until
   that task ships, the report is a recorded skip, not a blocker.
3. **Size:** about one working day. Bigger → propose a split as one owner question (section 4),
   with a recommended answer.
4. **Parts.** A plan over 8 tasks, or over about 30 KB, is written in parts of up to 8 tasks each, so
   each review reads a plan it can hold:
   - `plan.md` holds the summary, the coverage and the tables for the whole phase, plus Part 1's tasks;
   - `plan-2.md`, `plan-3.md` and so on hold the later parts' tasks, numbered on from the part before,
     each in the template's task format.

   It stays one phase, one branch per repo and one sign-off; `tasks.total` counts every part. Every
   part passes its review (section 3) before the build starts, and the parts are built in order.

## 3. Claude's plan review

1. Dispatch the `pal-review-plan` agent with the paths of `spec.md` and `plan.md` and its report
   path, `<set>/specs/NNN-slug/reviews/pal-review-plan-YYYY-MM-DD-HHmm.md`; it writes the report
   there (`references/review.md` section 2 step 3 has the check after it finishes).
2. It checks that every spec feature, test step, contract endpoint and appendix behaviour has a task;
   the code patterns (TDD, clean layers, SOLID, DRY); the approved contract, including the
   provider-repo / SSOT-repo split; trust handoffs; that each test step can run where the owner tests;
   the size of about one working day; and that every task names its test file and fixture. Its
   first round starts from the earlier majors: the major and critical findings of the project's
   earlier plan reviews and owner batches, plus the harness's fixed list (`pal-review-plan`).
3. **Round 1 is the whole review:** the reviewer runs every check and reports every finding it sees,
   however small, holding nothing back. A later round checks only the fixes and the text they
   changed. A plan in parts gets one review per part, its report named
   `pal-review-plan-partN-YYYY-MM-DD-HHmm.md`; Part 1's review also runs the whole-phase checks
   (coverage, size, the earlier majors). A round with only minor findings ends the review: apply
   them and write "minors applied, no re-review" in the log. A major or critical finding that does not change what gets built is fixed
   and the reviewer re-run, up to 3 rounds; it is listed in the summary's risks and choices.
4. Codex never reviews the plan.

## 4. No plan stop

The owner does not approve the plan. Write it, review it, fix it, save it, then build.

1. A major or critical finding that changes what gets built (a feature, the test script, the
   contract or the size) becomes one owner question: an owner stop with `kind: questions`
   (SKILL.md, **Owner stops**), with the reason, the impact and a recommended answer. Record the
   answer in `doc/decisions.md`, revise the plan and re-run `pal-review-plan`.
2. Otherwise set the plan's `Status:` to `Auto-approved (YYYY-MM-DD) — no plan stop`, with the date
   the plan was saved. The sign-off PR links the plan.
3. `gates.plan_approved` that date (`load_test` stays as the spec stage set it); `stage: building`;
   bundle features `building`. Commit and push.
4. Go to `.pal-harness/skills/phase/references/build.md`.
