# Linux checks on a Windows host

Read this when a phase or quick fix on a Windows PC changes code that only builds or runs on Linux:
`!windows` Go files, `#[cfg(unix)]` Rust, or anything a Linux CI job runs. It is the recipe Pal
Center phase 020 had to work out on its own, so no phase works it out again. CI's Linux job stays
the final proof; these local runs catch the failures before a 25-minute CI run does.

## 1. Cheap first

From Windows, before anything heavier, lint and vet as Linux. A Windows lint skips `!windows` files:

- `GOOS=linux golangci-lint run ./...` (the version CI pins)
- `GOOS=linux go vet ./...`, and `GOOS=darwin go vet ./...` for macOS-only files
- a Rust crate with no system libraries in its build scripts: `rustup target add x86_64-unknown-linux-gnu`,
  then `cargo clippy --target x86_64-unknown-linux-gnu --all-targets -- -D warnings` (macOS: the
  same with `aarch64-apple-darwin`). A crate whose build needs the target's libraries (Tauri, GTK,
  WebKit) fails here: use section 3 for Linux

None of these needs WSL or Docker. macOS code that only a Mac can build has no local proof: the
task's log line says `macOS proof: CI's macOS job`, handled like section 5.

## 2. WSL for Go tests

1. **Is it there?** `wsl.exe -l -v` lists the distributions. None → section 5.
2. **git 2.28 or later.** Tests that run `git init -b` or `git init --initial-branch` need it, and
   Ubuntu 20.04 ships git 2.25. Check with `wsl.exe -d Ubuntu -- git --version`. A newer git from the
   git-core PPA needs `sudo`, so it is an owner-run command (the plan's **Owner-run commands**), run
   once:
   `sudo add-apt-repository -y ppa:git-core/ppa && sudo apt-get update && sudo apt-get install -y git`.
   Until it is done, list the tests that fail only for the old git in the phase log as known WSL
   failures; CI's Linux job is their proof.
3. **Go in WSL** is a Linux build of the same version as `go.mod`'s `go` line, on the PATH of a login
   shell (`bash -lc`). Run with `GOTOOLCHAIN=local` so it never downloads another.
4. **The run.** Run from the worktree, with TMPDIR on a Linux-side tmpfs (`mktemp -d -p /dev/shm`). Never leave temp files
   under `/mnt/<drive>`: Windows drives in WSL do not keep Unix permissions or sockets, and are slow.

   ```bash
   wsl.exe -d Ubuntu -- bash -lc 'export TMPDIR=$(mktemp -d -p /dev/shm) && cd "$(wslpath -a "<worktree>")/<dir>" && GOTOOLCHAIN=local GOFLAGS=-p=1 GOMEMLIMIT=2GiB go test <packages>; rc=$?; rm -rf "$TMPDIR"; exit $rc'
   ```

   Read the result from the command's own exit code. Run the task's own packages with `-run` while
   building, and the full Linux suite once, in the task that owns it.

## 3. Docker for Rust and Tauri

WSL has no Rust toolchain by default; use Docker Desktop with Linux containers. Keep the target
folder inside the container (`/tmp/target`), never on the mounted Windows drive:

```bash
docker run --rm -m 6g -v "<worktree>:/src" -w /src/<dir> rust:1-bookworm bash -c "apt-get update -qq && apt-get install -y -qq libwebkit2gtk-4.1-dev libayatana-appindicator3-dev librsvg2-dev libxdo-dev libssl-dev && rustup component add clippy && CARGO_TARGET_DIR=/tmp/target cargo clippy --manifest-path src-tauri/Cargo.toml -- -D warnings && CARGO_TARGET_DIR=/tmp/target cargo test --manifest-path src-tauri/Cargo.toml"
```

The `apt-get` list is Tauri 2's; a project that needs others lists them in its plan. `docker info`
first: an error means Docker Desktop is not running (section 5).

## 4. Memory

- Before WSL tests: `pal-harness mem --min 2GiB`. Before a Docker Rust build:
  `pal-harness mem --min 4GiB` (a Tauri build with WebKit needs 4 to 6 GiB; `-m 6g` caps it).
- Run both in the foreground with a Bash timeout of 600000 ms (`devops_cicd.md`, **Free memory
  before heavy runs**). Claude Code stops background watchers when memory runs low.
- WSL's own limit is `memory=` under `[wsl2]` in `%UserProfile%\.wslconfig`. That is the owner's
  file: suggest a change in one line, never edit it.
- Still short after `--wait`: skip the heavy local run and say so in the log (section 5).

## 5. CI as the proof

When WSL or Docker is missing, or memory stays short, the task's log line says
`Linux proof: CI's Linux job`, with the reason. The first red Linux run in CI is then fixed in the
same task, by its normal rules; it is a fix round only if the code was wrong, not because the local
run was skipped.
