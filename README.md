# pal-harness

The shared AI harness for AI Workshop projects. It holds the project manifests,
generated tables, the pinned-copy writer/checker, numbering commands, and the
practices and templates content.

Projects consume it as a pinned `.pal-harness/` copy committed in their own repo
(see ADR-0004 in the pal-center repo), not as a live dependency.

## Commands (to come)

Run as `go run github.com/althenlimzixuan/pal-harness/cmd/pal-harness@v<version> <command>`:

- `check` - verify the pinned copy and generated tables are current
- `generate` - regenerate the tables
- `spec new` / `adr new` - create the next numbered spec or ADR
- `pin` - write the pinned `.pal-harness/` copy
- `version` - print the harness version
