package number

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mk(t *testing.T, root string, dirs []string, files []string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range files {
		p := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestNextSpec(t *testing.T) {
	root := t.TempDir()
	if n, err := NextSpec(root); err != nil || n != 1 {
		t.Fatalf("empty: %d %v", n, err)
	}
	mk(t, root, []string{"specs/001-a", "specs/007-b", "specs/notes", "specs/12-x"}, []string{"specs/009-file.md"})
	if n, err := NextSpec(root); err != nil || n != 8 {
		t.Fatalf("got %d %v, want 8", n, err)
	}
}

func TestNextADR(t *testing.T) {
	root := t.TempDir()
	if n, err := NextADR(root); err != nil || n != 1 {
		t.Fatalf("empty: %d %v", n, err)
	}
	mk(t, root, []string{"doc/ADR/0009-dir"}, []string{"doc/ADR/0001-a.md", "doc/ADR/0003-b.md", "doc/ADR/README.md", "doc/ADR/0004-c.txt"})
	if n, err := NextADR(root); err != nil || n != 4 {
		t.Fatalf("got %d %v, want 4", n, err)
	}
}

func TestValidSlug(t *testing.T) {
	for _, s := range []string{"ssot-layout", "a", "a1-b2"} {
		if !ValidSlug(s) {
			t.Errorf("%q should be valid", s)
		}
	}
	for _, s := range []string{"SSOT", "a--b", "-a", "a-", "", "a b", strings.Repeat("a", 51)} {
		if ValidSlug(s) {
			t.Errorf("%q should be invalid", s)
		}
	}
	if !ValidSlug(strings.Repeat("a", 50)) {
		t.Error("50 chars should be valid")
	}
}

func TestMaxSpecInNames(t *testing.T) {
	if got := MaxSpecInNames(nil); got != 0 {
		t.Fatalf("nil: %d", got)
	}
	got := MaxSpecInNames([]string{"specs/001-a", "specs/012-b", "specs/notes", "specs/7-x"})
	if got != 12 {
		t.Fatalf("got %d, want 12", got)
	}
}
