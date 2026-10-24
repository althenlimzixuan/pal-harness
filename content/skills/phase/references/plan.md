# Stage: plan (`planning`)

The plan turns the approved spec into tasks. There is no owner stop: once Claude's plan review
passes, building starts.

## 1. Write it

1. Use `superpowers:writing-plans` with the approved `spec.md`, and save into
   `specs/NNN-slug/plan.md` (never `docs/superpowers/`), in the format of
   `.pal-harness/templates/plan.md`. Without superpowers, write it by the template yourself and say
   so in the summary.
2. Every task names:
   1. its repo and the files it creates or changes;
   2. **the test to write first:** the test file and its JSON fixture
      `tests/{test_type}/{use_case}.json` (`testing.md`);
   3. the code that makes the test pass, in a line or two;
   4. the commands that prove it, and the result to expect;
   5. the commit message.
   A task without a test file and a fixture is not finished.
3. When the API changes, follow the split rule in `product_docs.md`: the provider repo's task
   regenerates its `openapi.yaml` and serves Swagger UI; an SSOT task on `docs/NNN-slug` updates the
   `doc/openapi.yaml` copy and `doc/API.md`. Contract tasks come before the tasks that call them.
4. Cover every spec feature and every test-script step with at least one task.

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

## 3. Claude's plan review

1. Dispatch the `pal-review-plan` agent with the paths of `spec.md` and `plan.md`. Save its report
   verbatim as `specs/NNN-slug/reviews/pal-review-plan-YYYY-MM-DD-HHmm.md`.
2. It checks that every spec feature and test step has a task; the code patterns (TDD, clean layers,
   SOLID, DRY); the approved contract, including the provider-repo / SSOT-repo split; the size of
   about one working day; and that every task names its test file and fixture.
3. Minor findings: fix them and re-run the reviewer, up to 3 rounds. A major or critical finding
   that does not change what gets built is fixed the same way and listed in the summary's risks
   and choices.
4. Codex never reviews the plan.

## 4. No plan stop

The owner does not approve the plan. Write it, review it, fix it, save it, then build.

1. A major or critical finding that changes what gets built (a feature, the test script, the
   contract or the size) becomes one owner question: an owner stop with `kind: questions`
   (SKILL.md, **Owner stops**), with the reason, the impact and a recommended answer. Record the
   answer in `doc/decisions.md`, revise the plan and re-run `pal-review-plan`.
2. Otherwise set the plan's `Status:` to `Auto-approved (YYYY-MM-DD) — no plan stop`, with the date
   the plan was saved. The sign-off PR links the plan.
3. `gates.plan_approved` that date; `load_test: true` when the owner's spec reply said "include load
   test"; `stage: building`; bundle features `building`. Commit and push.
4. Go to `.pal-harness/skills/phase/references/build.md`.
