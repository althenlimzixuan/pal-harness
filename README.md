# pal-harness

The shared AI harness for AI Workshop projects. It holds the project manifests,
generated tables, the pinned-copy writer/checker, numbering commands, and the
practices and templates content.

Projects consume it as a pinned `.pal-harness/` copy committed in their own repo
(see ADR-0004 in the pal-center repo), not as a live dependency.

## Commands

Run as `go run github.com/althenlimzixuan/pal-harness/cmd/pal-harness@v<version> <command>`,
where `<version>` is the project's `harness.version`. `dir` defaults to `.`.

| Command | What it does |
|---|---|
| `check [dir]` | In an SSOT root (`project.yaml`): validates the manifest, checks that the generated blocks in `README.md` and `CLAUDE.md` match it, that `.pal-harness/` matches this release (overridden items skipped), and that `.github/workflows/ssot.yml` calls `check.yml@v<version>` with that version. In the workspace root (`workspace.yaml`): checks the registry block in `CLAUDE.md`. On a `harness.version` mismatch it reports only that and the command to run instead. |
| `generate [dir]` | Rewrites the generated blocks from the manifest. Writes nothing if any block is missing or broken. |
| `pin [dir]` | Writes this release's `.pal-harness/` copy, leaving overridden items untouched. SSOT root only. |
| `spec new <slug> [dir]` | On a clean tree, creates branch `docs/NNN-slug` from `dev` with `specs/NNN-slug/spec.md`. `NNN` is one past the highest number in `specs/`, on `dev` and `origin/dev`, and in any `docs/NNN-*` branch, local or remote. SSOT root only. |
| `adr new <slug> [--conflict] [dir]` | Creates `doc/ADR/NNNN-slug.md` from the ADR (or conflict ADR) template. `NNNN` is one past the highest number in `doc/ADR/`, on `dev` and on `origin/dev`. SSOT root only. |
| `version` | Prints the harness version. |

Exit codes: `0` clean or done, `1` findings (or a block `generate` cannot write), `2` usage,
I/O or precondition error.
