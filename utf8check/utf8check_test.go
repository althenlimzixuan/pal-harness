package utf8check

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// repo makes a git repo with the tracked files committed and the untracked files only written.
func repo(t *testing.T, tracked, untracked map[string][]byte) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	put := func(files map[string][]byte) {
		for p, b := range files {
			full := filepath.Join(dir, filepath.FromSlash(p))
			if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(full, b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	put(tracked)
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"commit", "-q", "--allow-empty", "-m", "c"}} {
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com",
			"-c", "commit.gpgsign=false", "-c", "core.autocrlf=false"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	put(untracked)
	return dir
}

func TestCheck(t *testing.T) {
	dir := repo(t, map[string][]byte{
		"ok.md":            []byte("plain\nUTF-8: café ✓\n"),
		"docs/cp1252.md":   []byte("line one\nline two\nquote \x93here\x94\n"),
		"logo.png":         append([]byte("\x89PNG\r\n\x1a\n\x00\x00"), 0x93, 0xff),
		"bom.txt":          []byte("\xef\xbb\xbfwith a BOM\n"),
		"late-nul.txt":     append(append([]byte("x\n\xff\n"), bytes.Repeat([]byte("a"), 9000)...), 0),
	}, map[string][]byte{
		"untracked.md": []byte("\x93untracked\n"),
	})
	got, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []Finding{{Path: "docs/cp1252.md", Line: 3}, {Path: "late-nul.txt", Line: 2}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Check = %+v, want %+v", got, want)
	}
}

func TestCheckClean(t *testing.T) {
	dir := repo(t, map[string][]byte{"a.go": []byte("package a\n")}, nil)
	got, err := Check(dir)
	if err != nil || len(got) != 0 {
		t.Fatalf("Check = %v, %v; want none", got, err)
	}
}

func TestCheckNotARepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	_, err := Check(t.TempDir())
	if !errors.Is(err, ErrNotRepo) {
		t.Fatalf("err = %v, want ErrNotRepo", err)
	}
}

func TestFindingString(t *testing.T) {
	if got := (Finding{Path: "a/b.md", Line: 7}).String(); got != "a/b.md:7: not UTF-8" {
		t.Errorf("String = %q", got)
	}
}
