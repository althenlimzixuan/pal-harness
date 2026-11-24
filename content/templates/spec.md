# <NNN> — <Title>

- **Status:** Draft
- **Roadmap card:**
- **Vehicle:** phase

> Reads in about 10 minutes: about 1,500 words plus diagrams. Detail that does not fit goes in the
> Appendix. If it still does not fit, split the phase. Written in plain English for you: no jargon,
> and the test steps and diagrams use the words you would use.

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

## Rulings

(Only while you change the spec in the session: your rulings, numbered and dated, one line each.
Delete the section when there are none.)

1. YYYY-MM-DD: (the ruling)

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
3. (only when a result may surprise you: Do: … → Expect: … by design: why; delete it otherwise)
4. (only when it needs hardware you can't reach before sign-off: Do: … → Expect: …
   `unverified: <reason>`; the sign-off PR lists it)

(A step whose result is right but may surprise you says why, in the step: "→ Expect: a new code
arrives (by design: the app never says whether an account exists)".)

(A destructive or optional step, such as removing the phone's screen lock, goes in their own feature
block with its own script, so you can skip it on purpose.)

(With the `no-owner-test` override, replace the test script with **What to expect now:** a list of
what behaves differently and where to see it, one line each, with nothing to run.)

## UX outline

(Screens or steps in plain words, one short line each. Delete this section when nothing on screen
changes.)

**QA seeds:** (the records each changed screen needs to look filled in the QA environment, such as
"3 audit lines, a paused project, a held decision"; or "none")

## Data sources

Delete this section when the phase builds every figure it shows. Otherwise one row per figure, list
or badge the screens show from data the API already serves.

| Figure | Endpoint and field | Checked in |
|---|---|---|
| (a count, a list, a badge) | `GET /api/v1/(path)` → `(field)` | `doc/openapi.yaml` line N, or (handler file:line) |

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

## Trust boundaries

Delete this section when no code runs as another account or elevated, and nothing reads a file,
config, environment, pipe or registry value that another account or a less-trusted process can write.

| Input | Who can write it | Who reads or runs it, as which account | Session or desktop | What stops misuse | What it does not stop |
|---|---|---|---|---|---|
| (a file, config, environment, pipe or registry value) | (account or process) | (process, account) | (your desktop, session 0, a hidden desktop) | (an ACL, a check in code and its test) | (what is left open) |

Processes that run as another account or elevated, one line each: who starts it, which binary (and
who can write that path), and where its environment comes from.

## Out of scope

- (Something a reader might expect, and is not in this phase.)

## Appendix

<details>
<summary>Detail for those who want it</summary>

(Edge cases, data shapes, alternatives considered.)

(A rule that follows another feature's rule or a decision, such as a relay of its cases, lists
every case of the original rule, taken from that rule's own list (its spec, the decision or its
code), each with its wording. It ends with `cases: N of N`. A case left out on purpose stays listed
with `not covered: <why>`.)

</details>
