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
| `check [dir]` | In an SSOT root (`project.yaml`): validates the manifest, checks that the generated blocks in `README.md` and `CLAUDE.md` match it, that `.pal-harness/` matches this release (overridden items skipped), and that `.github/workflows/ssot.yml` calls `check.yml@v<version>` with that version. It also validates `doc/features.yaml` and `specs/*/phase-state.yaml`, and reports drift in the loaders under `.claude/`. In the workspace root (`workspace.yaml`): checks the registry block in `CLAUDE.md`. On a `harness.version` mismatch it reports only that and the command to run instead. |
| `generate [dir]` | Rewrites the generated blocks from the manifest. Writes nothing if any block is missing or broken. |
| `pin [dir]` | Writes this release's `.pal-harness/` copy, leaving overridden items untouched, and the skill and agent loaders into `.claude/skills/` and `.claude/agents/`. SSOT root only. |
| `spec new <slug> [dir]` | On a clean tree, creates branch `docs/NNN-slug` from `dev` with `specs/NNN-slug/spec.md`. `NNN` is one past the highest number in `specs/`, on `dev` and `origin/dev`, and in any `docs/NNN-*` branch, local or remote. SSOT root only. |
| `adr new <slug> [--conflict] [dir]` | Creates `doc/ADR/NNNN-slug.md` from the ADR (or conflict ADR) template. `NNNN` is one past the highest number in `doc/ADR/`, on `dev` and on `origin/dev`. SSOT root only. |
| `render <file.md>` | Renders a Markdown file as a phone-readable page. |
| `log <phase folder> <headline> [body file]` | Appends `## YYYY-MM-DD HH:MM — <headline>` (and the body file's text) to `phase-log.md` and sets `last_progress_at` in `phase-state.yaml`, both from the system clock. The folder must hold `phase-state.yaml` and sit inside a worktree set (`.pal-worktrees/`); otherwise it writes nothing. |
| `contrast <foreground> <background>` | Prints the WCAG 2 contrast ratio of two colours (`#rgb`, `#rrggbb`, `rgb()`, `oklch()`) and whether it passes AA for text (4.5:1) and for large text and UI parts (3:1). An `oklch()` colour outside sRGB is clipped, with a note; alpha is refused. |
| `mem [--min 2GiB] [--wait 0] [--every 30s]` | Prints the available physical memory against the minimum. With `--wait`, re-checks every `--every` until it is enough or the wait ends. Windows and Linux. Exit `0` enough, `1` still short, `2` usage or a platform it cannot read. |
| `adb [dir]` | Reads `doc/devices.yaml` (ports, and devices with a name, a role, a USB serial and a last wireless address), reconnects each device at its saved address or the one `adb mdns services` finds, and sets `adb reverse` for each port. Prints one line per device, a new address to save, and the pairing request for a device still missing. Never disconnects a device or restarts the adb server. Exit `0` every device ready, `1` one not ready, `2` usage, no `doc/devices.yaml` or no `adb`. |
| `fingerprint [--rev HEAD] [--with <path>]... [--dir .] <spec \| ->` | Reads a spec of one `<job>: <path> <path> ...` line per CI job and prints `<job>=<fingerprint>` for each: the SHA-1 of the job name and the git object id at `--rev` of each path it lists and each `--with` path (`missing` when absent). The CI gate skips a job whose fingerprint already passed. Exit `0` printed, `2` usage, a bad spec or not a git repository. |
| `utf8 [dir]` | Checks every file git tracks under `dir` (default `.`) and prints `<path>:<line>: not UTF-8` for each text file that is not valid UTF-8, with its first bad line. Binary files (a NUL in the first 8000 bytes) are skipped. Exit `0` clean, `1` a bad file, `2` usage or not a git repository. |
| `version` | Prints the harness version. |

Exit codes: `0` clean or done, `1` findings (or a block `generate` cannot write), `2` usage,
I/O or precondition error.

## Reusable workflows

Called from a project's `.github/workflows/` at `@v<version>`:

| Workflow | What it does |
|---|---|
| `check.yml` | Runs `pal-harness check` (the SSOT repo's `ssot.yml`). |
| `ci-gate.yml` | Runs `utf8` (a non-UTF-8 text file makes the gate red), then fingerprints each job's paths (`fingerprint`) and outputs `decisions` (`run` or `skip`) and `fingerprints` as JSON; a job whose fingerprint already passed is skipped. |
| `ci-mark.yml` | Records each job that passed as `refs/ci-passed/<job>/<fingerprint>`. |

A project's `ci.yml` starts from `.pal-harness/templates/ci.yml` (`devops_cicd.md`, **The CI gate**).
