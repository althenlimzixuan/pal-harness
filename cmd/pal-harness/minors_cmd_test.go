package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const ledger = `# SDD ledger — plan: specs/022-x/plan.md
Task 1: complete (commits abc1234..def5678, 0 parked)
- Task 3: minor (deferred) [carry]: Cancel label reads "Close"
Task 3: minor (deferred) [carry]: stale comment in tray.rs
Task 2: minor (deferred) [later]: log line could name the port
Task 4: minor (deferred) [final]: three safety notes in exit_all
Task 5: minor (deferred) [issue]: e2e spec fails on dev
Task 2: minor (deferred) [fixed]: typo in notice
Task 4: parked — PID reuse window — Ruling: sub-second, command-only tool
Task 5: minor (deferred): no notice while Stop all runs
Task 6: minor (deferred) [soon]: wording of the hold reason
Task 6: Ruling: kept the old route — cheaper
`

func ledgerFile(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, "progress.md", text)
	return filepath.Join(dir, "progress.md")
}

func TestMinorsGroupsByTag(t *testing.T) {
	code, out, errs := do(t, "minors", ledgerFile(t, ledger))
	if code != 1 {
		t.Fatalf("code %d (want 1: untagged lines) err=%s", code, errs)
	}
	want := `carry (2)
  Task 3: Cancel label reads "Close"
  Task 3: stale comment in tray.rs
final (1)
  Task 4: three safety notes in exit_all
later (1)
  Task 2: log line could name the port
issue (1)
  Task 5: e2e spec fails on dev
parked (1)
  Task 4: PID reuse window — Ruling: sub-second, command-only tool
untagged (2)
  Task 5: minor (deferred): no notice while Stop all runs
  Task 6: minor (deferred) [soon]: wording of the hold reason
fixed (1)
`
	if out != want {
		t.Fatalf("output:\n%s\nwant:\n%s", out, want)
	}
	if !strings.Contains(errs, "2 minors have no known tag") {
		t.Fatalf("stderr %q", errs)
	}
}

func TestMinorsAllTagged(t *testing.T) {
	text := "Task 1: minor (deferred) [later]: a\nTask 2: minor (deferred) [fixed]: b\n"
	code, out, errs := do(t, "minors", ledgerFile(t, text))
	if code != 0 || out != "later (1)\n  Task 1: a\nfixed (1)\n" || errs != "" {
		t.Fatalf("code %d out %q err %q", code, out, errs)
	}
	code, out, _ = do(t, "minors", ledgerFile(t, "# SDD ledger\nTask 1: complete\n"))
	if code != 0 || out != "no deferred minors\n" {
		t.Fatalf("empty: code %d out %q", code, out)
	}
}

func TestMinorsFilters(t *testing.T) {
	f := ledgerFile(t, ledger)
	code, out, _ := do(t, "minors", "--task", "3", "--tag", "carry", f)
	if code != 0 || out != "carry (2)\n  Task 3: Cancel label reads \"Close\"\n  Task 3: stale comment in tray.rs\n" {
		t.Fatalf("task 3 carry: code %d out %q", code, out)
	}
	// Task 5 has an untagged line: exit 1 even though one is tagged.
	code, out, _ = do(t, "minors", "--task", "5", f)
	if code != 1 || !strings.Contains(out, "issue (1)") || !strings.Contains(out, "untagged (1)") {
		t.Fatalf("task 5: code %d out %q", code, out)
	}
	code, out, _ = do(t, "minors", "--tag", "later", f)
	if code != 0 || out != "later (1)\n  Task 2: log line could name the port\n" {
		t.Fatalf("tag later: code %d out %q", code, out)
	}
	code, out, _ = do(t, "minors", "--tag", "untagged", f)
	if code != 1 || !strings.HasPrefix(out, "untagged (2)") {
		t.Fatalf("tag untagged: code %d out %q", code, out)
	}
}

func TestMinorsUsage(t *testing.T) {
	f := ledgerFile(t, ledger)
	for name, args := range map[string][]string{
		"no file":      {"minors"},
		"two files":    {"minors", f, f},
		"bad tag":      {"minors", "--tag", "soon", f},
		"bad task":     {"minors", "--task", "x", f},
		"zero task":    {"minors", "--task", "0", f},
		"missing file": {"minors", filepath.Join(t.TempDir(), "none.md")},
	} {
		if code, _, errs := do(t, args...); code != 2 || errs == "" {
			t.Errorf("%s: code %d err %q", name, code, errs)
		}
	}
}
