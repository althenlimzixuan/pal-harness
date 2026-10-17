# Conflict Protocol

Applies in every session. The precedence ladder is in `ssot_standard.md` §1.

## Trigger

Two sources disagree about a fact (a doc and code, a component and the SSOT, a mirror and its
ADR, two docs).

## Steps

1. **Stop.** Do not continue work that depends on the disputed fact.
2. **List the sources in ladder order** (SSOT repo, then provider and platform components, then
   consumers). For each give the file path and a short quote.
3. **Ask the owner which is right.** Never pick silently, even when the ladder suggests an answer.
4. **After the answer,** run `pal-harness adr new <slug> --conflict` and fill in the headings:
   - *Finding*: which sources disagreed, quoted with paths
   - *Decision point*: the question that was open
   - *Decision*: the owner's answer
   - *Reason*: why
   - *Changes*: what was updated in the same piece of work
5. **Update the losing source** in the same piece of work, even when it is the SSOT, so the sources
   agree again.

The ADR lands in `doc/ADR/` of the SSOT repo, numbered with the other ADRs.

## Exception

`context/legacy/` is reference only. Where it disagrees with the SSOT, the SSOT wins and no ADR is
written. Do not ask.

## Not a conflict

- A generated block that no longer matches `project.yaml`: run `pal-harness generate`.
- A generated copy (such as `openapi.yaml`) that differs from its origin: regenerate it.

If the origin itself looks wrong (e.g. `project.yaml` or the provider is stale), that is a conflict:
follow the steps above.
