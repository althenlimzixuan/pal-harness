package main

import (
	"strings"
	"testing"
)

func TestUTF8Command(t *testing.T) {
	clean := committedRepo(t, map[string]string{"a.md": "café\n"})
	code, out, errOut := do(t, "utf8", clean)
	if code != 0 || errOut != "" || !strings.Contains(out, "UTF-8") {
		t.Fatalf("clean: %d %q %q", code, out, errOut)
	}
	bad := committedRepo(t, map[string]string{"a.md": "ok\n\x93bad\x94\n"})
	code, out, _ = do(t, "utf8", bad)
	if code != 1 || !strings.Contains(out, "a.md:2: not UTF-8") {
		t.Fatalf("bad: %d %q", code, out)
	}
}

func TestUTF8CommandErrors(t *testing.T) {
	committedRepo(t, map[string]string{"a.md": "a\n"}) // skips without git
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"two dirs", []string{"utf8", "a", "b"}, "at most one directory"},
		{"not a repo", []string{"utf8", t.TempDir()}, "not a git repository"},
	} {
		t.Run(c.name, func(t *testing.T) {
			code, _, errOut := do(t, c.args...)
			if code != 2 || !strings.Contains(errOut, c.want) {
				t.Fatalf("%d %q, want 2 and %q", code, errOut, c.want)
			}
		})
	}
}
