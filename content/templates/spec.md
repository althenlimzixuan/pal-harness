# <NNN> — <Title>

- **Status:** Draft
- **Roadmap card:**
- **Vehicle:** phase

> Reads in about 10 minutes: about 1,500 words plus diagrams. Detail that does not fit goes in the
> Appendix. If it still does not fit, split the phase.

## Summary

1. (What this phase builds, in one sentence.)
2. (Which features it covers, by id and name.)
3. (What you will be able to do at the end.)
4. (The biggest choice made for you, and why.)
5. (Questions below that need your answer: how many.)

## Questions for the owner

Only questions that change what gets built. Reply in one line, like "accept all" or
"accept all except 2: B".

1. (The question.) Options: A) (…) B) (…). **Recommended: A** — (why, in one line).

## Calls I made

Small decisions already made. Tell me if you disagree with any.

1. (The call, and the reason in one line.)

## Features

One block per feature. Approving this spec approves how you will test it.

### F-000 — (name)

**Why:** (One or two lines: the need behind it.)

**Use case:** (One or two lines: who does what with it.)

**Test script:**

1. Do: (what you do) → Expect: (what you should see)
2. Do: (what you do) → Expect: (what you should see)

(With the `no-owner-test` override, replace the test script with **What to expect now:** a list of
what behaves differently and where to see it, one line each, with nothing to run.)

## UX outline

(Screens or steps in plain words, one short line each. Delete this section when nothing on screen
changes.)

## Contract

Delete this section when the phase does not change the API.

| Method and path | What it does | Errors |
|---|---|---|
| `POST /api/v1/(path)` | (one line) | 401, 403, 422 |

[diagram](diagrams/(name).mmd)

A breaking change goes in Questions for the owner, with its impact.

## Infrastructure

[Infrastructure](../../doc/infra.mmd)

Added or changed in this phase: (each part, one line), or "no change".

(`doc/infra.mmd` is the project's one diagram of its runtime parts. No file yet → draft it from
`.pal-harness/templates/infra.mmd`. Names and kinds only, like "API on Cloud Run" or "MySQL"; never
project ids, hostnames, client ids, keys or URLs.)

## Out of scope

- (Something a reader might expect, and is not in this phase.)

## Appendix

<details>
<summary>Detail for those who want it</summary>

(Edge cases, data shapes, alternatives considered.)

</details>
