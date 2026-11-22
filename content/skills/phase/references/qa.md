# The QA pass

Before the owner is asked to test, the session checks the build itself in the project's QA
environment, so the owner only tests a build that already passed. It runs only when `project.yaml`
sets `harness.qa_command`; with it unset, nothing here runs and `qa:` stays `none`.

## 1. When it runs

- After the sign-off merge and the promotion to `staging` (`references/signoff.md`, section 5),
  before the staging test.
- Under `no-owner-test`: after the sign-off merge, before ship.
- Under `uat-queue`: once per phase, before that phase joins the queue.
- A quick fix runs it too, for its one test step.

## 2. Run it

1. Set `qa: running` in `phase-state.yaml`, write the log entry, commit and push.
2. Free memory first (`devops_cicd.md`, **Free memory before heavy runs**). Start the environment
   with `harness.qa_command` from the SSOT worktree, from `origin/staging` of each repo (as
   `references/staging-test.md` section 1 step 2 checks it out). It runs with fakes, so it never
   needs the owner's accounts or real data.
3. **The warm-up.** Before the first step, load the app's first page once and wait for it to render,
   with no screenshot: a dev server (Vite, Next.js, Metro) compiles on its first load, and that cold
   start can outlast a step's timeout. A step whose first page load still times out is retried once,
   and `summary.md` notes `cold start: retried` on that row; a second timeout fails the step.
4. Follow each bundle feature's test script in the spec (under `no-owner-test`, each What to expect
   now line). For each step: do it, compare with its expected result, and save a screenshot.
5. Record the run in `specs/NNN-slug/qa/`:
   - `summary.md`: one row per step (feature, step, pass or fail, a one-line note, the screenshot);
     the date, the commit of each repo and the command used at the top;
   - `screens/`: one PNG per step, named `F-012-step-3.png`.
6. Stop the environment and everything it started.

## 3. The result

- **Every step passed:** `qa: passed`. Commit and push the `qa/` folder on `docs/NNN-slug`, then go
  on to the staging test (or ship). The staging-test message links the QA summary.
- **A step failed:** `qa: failed`. It is a fix round, handled like a failed staging step under fix
  now (`references/staging-test.md`, section 5): a `fix/NNN-<short>` branch from `origin/staging`,
  test first, one review, the PR into `staging`, then into `dev`. Then run the QA pass again from
  section 2.
- **3 fix rounds and still failing:** an owner stop with `kind: findings`: the note
  (`reviews/owner-YYYY-MM-DD-HHmm.md`) gives the failing steps, their screenshots, what was tried,
  and options with a recommended one.

A QA failure never reaches the owner's test: the owner is asked to test only after `qa: passed`, or
after the findings stop's answer.
