package main

import (
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// committedRepo makes a repo with one commit holding the files.
func committedRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	for p, body := range files {
		write(t, dir, p, body)
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"commit", "-q", "-m", "c"}} {
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com",
			"-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

var sumLine = regexp.MustCompile(`^[a-z0-9_-]+=[0-9a-f]{40}$`)

func TestFingerprintCommand(t *testing.T) {
	dir := committedRepo(t, map[string]string{"api/main.go": "package main\n", "web/a.ts": "x\n",
		".github/workflows/ci.yml": "name: ci\n", "jobs.txt": "api: api\nweb: web\n"})
	code, out, errOut := do(t, "fingerprint", "--dir", dir, "--with", ".github/workflows/ci.yml",
		filepath.Join(dir, "jobs.txt"))
	if code != 0 {
		t.Fatalf("%d %q", code, errOut)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "api=") || !strings.HasPrefix(lines[1], "web=") {
		t.Fatalf("%q", out)
	}
	for _, l := range lines {
		if !sumLine.MatchString(l) {
			t.Errorf("%q is not <job>=<sha1>", l)
		}
	}

	// The shared path is part of every fingerprint.
	_, plain, _ := do(t, "fingerprint", "--dir", dir, filepath.Join(dir, "jobs.txt"))
	if plain == out {
		t.Error("--with did not change the fingerprints")
	}

	// Stdin gives the same result as the file.
	old := stdin
	stdin = strings.NewReader("api: api\nweb: web\n")
	t.Cleanup(func() { stdin = old })
	code, fromStdin, _ := do(t, "fingerprint", "--rev", "HEAD", "--dir", dir, "--with", ".github/workflows/ci.yml", "-")
	if code != 0 || fromStdin != out {
		t.Fatalf("stdin: %d %q, want %q", code, fromStdin, out)
	}
}

func TestFingerprintCommandErrors(t *testing.T) {
	dir := committedRepo(t, map[string]string{"a.txt": "a\n", "jobs.txt": "api: a.txt\n", "bad.txt": "API: a.txt\n"})
	jobs := filepath.Join(dir, "jobs.txt")
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"no spec", []string{"fingerprint", "--dir", dir}, "one spec file"},
		{"two specs", []string{"fingerprint", "--dir", dir, jobs, jobs}, "one spec file"},
		{"bad flag", []string{"fingerprint", "--nope", jobs}, "fingerprint:"},
		{"bad with", []string{"fingerprint", "--dir", dir, "--with", "../x", jobs}, "--with"},
		{"missing spec", []string{"fingerprint", "--dir", dir, filepath.Join(dir, "none.txt")}, "cannot read"},
		{"bad spec", []string{"fingerprint", "--dir", dir, filepath.Join(dir, "bad.txt")}, "job name"},
		{"bad rev", []string{"fingerprint", "--dir", dir, "--rev", "nope", jobs}, "not a git repository"},
		{"not a repo", []string{"fingerprint", "--dir", t.TempDir(), jobs}, "not a git repository"},
	} {
		t.Run(c.name, func(t *testing.T) {
			code, _, errOut := do(t, c.args...)
			if code != 2 || !strings.Contains(errOut, c.want) {
				t.Fatalf("%d %q, want 2 and %q", code, errOut, c.want)
			}
		})
	}
}
