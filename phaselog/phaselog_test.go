package phaselog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const stateYAML = `phase: "096"
slug: trip-budget
stage: building
last_progress_at: 2026-01-01T09:00:00+08:00
notes: []
`

// phaseDir makes <tmp>/.pal-worktrees/ETP-096/specs/096-trip-budget with a phase-state.yaml.
func phaseDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), ".pal-worktrees", "ETP-096", "specs", "096-trip-budget")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "phase-state.yaml"), []byte(stateYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func read(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

var at = time.Date(2026, 11, 2, 14, 5, 33, 0, time.FixedZone("MYT", 8*3600))

func TestAppendCreatesTheLogAndStampsTheState(t *testing.T) {
	dir := phaseDir(t)
	stamp, err := Append(dir, "Task 3 done", "", at)
	if err != nil {
		t.Fatal(err)
	}
	if stamp != "2026-11-02 14:05" {
		t.Errorf("stamp = %q", stamp)
	}
	if got := read(t, filepath.Join(dir, "phase-log.md")); got != "## 2026-11-02 14:05 — Task 3 done\n" {
		t.Errorf("phase-log.md = %q", got)
	}
	st := read(t, filepath.Join(dir, "phase-state.yaml"))
	if !strings.Contains(st, "\nlast_progress_at: 2026-11-02T14:05:33+08:00\n") {
		t.Errorf("last_progress_at not stamped:\n%s", st)
	}
	if strings.Count(st, "last_progress_at") != 1 || !strings.Contains(st, "stage: building\n") {
		t.Errorf("other lines changed:\n%s", st)
	}
}

func TestAppendAddsBodyAfterABlankLine(t *testing.T) {
	dir := phaseDir(t)
	log := filepath.Join(dir, "phase-log.md")
	if err := os.WriteFile(log, []byte("## 2026-11-02 09:00 — Started\n\nFirst entry."), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(dir, "  Resumed  ", "\nCI green on api.\nWaiting on review.\n\n", at); err != nil {
		t.Fatal(err)
	}
	want := "## 2026-11-02 09:00 — Started\n\nFirst entry.\n\n## 2026-11-02 14:05 — Resumed\n\nCI green on api.\nWaiting on review.\n"
	if got := read(t, log); got != want {
		t.Errorf("phase-log.md =\n%q\nwant\n%q", got, want)
	}
}

func TestAppendUTCKeepsANumericOffset(t *testing.T) {
	dir := phaseDir(t)
	if _, err := Append(dir, "x", "", time.Date(2026, 11, 2, 6, 5, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if st := read(t, filepath.Join(dir, "phase-state.yaml")); !strings.Contains(st, "last_progress_at: 2026-11-02T06:05:00+00:00\n") {
		t.Errorf("want +00:00, got:\n%s", st)
	}
}

func TestAppendRefusesAndWritesNothing(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "specs", "096-trip-budget")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "phase-state.yaml"), []byte(stateYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	noState := filepath.Join(t.TempDir(), ".pal-worktrees", "ETP-1", "specs", "096-x")
	if err := os.MkdirAll(noState, 0o755); err != nil {
		t.Fatal(err)
	}
	noStamp := phaseDir(t)
	if err := os.WriteFile(filepath.Join(noStamp, "phase-state.yaml"), []byte("phase: \"096\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok := phaseDir(t)

	cases := []struct {
		name, dir, headline, want string
	}{
		{"outside a set", outside, "x", "outside a worktree set"},
		{"no phase-state.yaml", noState, "x", "phase-state.yaml"},
		{"no last_progress_at line", noStamp, "x", "last_progress_at"},
		{"empty headline", ok, "  ", "headline"},
		{"two-line headline", ok, "a\nb", "one line"},
	}
	for _, c := range cases {
		_, err := Append(c.dir, c.headline, "", at)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err = %v, want it to mention %q", c.name, err, c.want)
		}
		if _, statErr := os.Stat(filepath.Join(c.dir, "phase-log.md")); statErr == nil {
			t.Errorf("%s: phase-log.md was written", c.name)
		}
	}
	if st := read(t, filepath.Join(outside, "phase-state.yaml")); st != stateYAML {
		t.Errorf("phase-state.yaml outside the set changed:\n%s", st)
	}
}

func TestInWorktreeSet(t *testing.T) {
	cases := map[string]bool{
		filepath.Join("D:", "ws", ".pal-worktrees", "PC-1", "specs", "001-x"): true,
		filepath.Join("D:", "ws", "PC-Pal Center", "specs", "001-x"):          false,
		filepath.Join("D:", "ws", "not.pal-worktrees", "specs"):               false,
	}
	for p, want := range cases {
		if got := InWorktreeSet(p); got != want {
			t.Errorf("InWorktreeSet(%q) = %v, want %v", p, got, want)
		}
	}
}
