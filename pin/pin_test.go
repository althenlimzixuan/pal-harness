package pin

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	harness "github.com/althenlimzixuan/pal-harness"
	"github.com/althenlimzixuan/pal-harness/manifest"
)

func abs(root, rel string) string { return filepath.Join(root, Dir, filepath.FromSlash(rel)) }

func written(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := Write(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func mutate(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(abs(root, "practices/workflow.md"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(abs(root, "templates/spec.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs(root, "practices/extra.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFilesIncludesVersion(t *testing.T) {
	var found bool
	for _, f := range Files() {
		if !strings.HasPrefix(f.Path, Dir+"/") {
			t.Errorf("path %q lacks prefix", f.Path)
		}
		if strings.Contains(f.Content, "\r") {
			t.Errorf("%s contains CR", f.Path)
		}
		if f.Path == Dir+"/VERSION" {
			found = true
			if f.Content != harness.Version+"\n" {
				t.Errorf("VERSION = %q", f.Content)
			}
		}
	}
	if !found {
		t.Fatal("VERSION missing")
	}
}

func TestFilesSorted(t *testing.T) {
	fs := Files()
	for i := 1; i < len(fs); i++ {
		if fs[i-1].Path >= fs[i].Path {
			t.Fatalf("not sorted: %s >= %s", fs[i-1].Path, fs[i].Path)
		}
	}
}

func TestWriteThenDiffClean(t *testing.T) {
	got, err := Diff(written(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want clean, got %v", got)
	}
}

func TestDiffFindings(t *testing.T) {
	root := written(t)
	mutate(t, root)
	got, err := Diff(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"changed practices/workflow.md", "extra practices/extra.md", "missing templates/spec.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDiffSkipsOverrides(t *testing.T) {
	root := written(t)
	mutate(t, root)
	got, err := Diff(root, []manifest.Override{{Item: "practices/workflow.md"}, {Item: "templates"}})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"extra practices/extra.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDiffIgnoresCRLF(t *testing.T) {
	root := written(t)
	for _, f := range Files() {
		p := filepath.Join(root, filepath.FromSlash(f.Path))
		if err := os.WriteFile(p, []byte(strings.ReplaceAll(f.Content, "\n", "\r\n")), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := Diff(root, nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestDiffMissingDirAllMissing(t *testing.T) {
	got, err := Diff(t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(Files()) {
		t.Fatalf("got %v", got)
	}
	for _, g := range got {
		if !strings.HasPrefix(g, "missing ") {
			t.Errorf("unexpected %q", g)
		}
	}
}

func TestDiffIOError(t *testing.T) {
	root := written(t)
	// a directory where a file is expected
	p := abs(root, "practices/workflow.md")
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Diff(root, nil); err == nil {
		t.Fatal("want error")
	}
}

func TestWriteError(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, Dir), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(root); err == nil {
		t.Fatal("want error")
	}
}
