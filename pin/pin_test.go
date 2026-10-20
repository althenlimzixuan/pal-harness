package pin

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

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

func TestWriteExceptKeepsOverrides(t *testing.T) {
	root := written(t)
	mutate(t, root) // workflow.md edited, templates/spec.md removed, practices/extra.md added
	if err := os.WriteFile(abs(root, "practices/testing.md"), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	overrides := []manifest.Override{{Item: "practices/workflow.md"}, {Item: "templates"}}
	if err := WriteExcept(root, overrides); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(abs(root, "practices/workflow.md")); string(b) != "edited\n" {
		t.Fatalf("overridden file rewritten: %q", b)
	}
	if _, err := os.Stat(abs(root, "templates/spec.md")); err == nil {
		t.Fatal("file under an overridden folder was written")
	}
	if b, _ := os.ReadFile(abs(root, "practices/testing.md")); string(b) == "edited\n" {
		t.Fatal("non-overridden file was not restored")
	}
	got, err := Diff(root, overrides)
	if err != nil || !reflect.DeepEqual(got, []string{"extra practices/extra.md"}) {
		t.Fatalf("Diff = %v, %v", got, err)
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

var demoFS = fstest.MapFS{
	"content/skills/demo/SKILL.md": {Data: []byte("---\nname: demo\ndescription: Demo skill\n---\n\n# Demo\n")},
	"content/agents/review-x.md":   {Data: []byte("---\nname: pal-review-x\ndescription: X\ntools: Read, Grep\n---\n\nBody\n")},
}

func TestLoadersFromPathsAndContent(t *testing.T) {
	got, err := loadersFrom(demoFS)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Path != ".claude/agents/pal-review-x.md" || got[1].Path != ".claude/skills/demo/SKILL.md" {
		t.Fatalf("got %+v", got)
	}
	wantSkill := "---\nname: demo\ndescription: Demo skill\n---\n\n" +
		"<!-- generated by pal-harness pin from .pal-harness/skills/demo/SKILL.md; do not edit -->\n\n" +
		"Read `.pal-harness/skills/demo/SKILL.md` and follow it exactly. It is the pinned harness copy of this skill.\n"
	if got[1].Content != wantSkill {
		t.Errorf("skill loader = %q", got[1].Content)
	}
	if !strings.Contains(got[0].Content, "tools: Read, Grep\n---\n\n<!-- generated by pal-harness pin from .pal-harness/agents/review-x.md; do not edit -->") ||
		!strings.HasSuffix(got[0].Content, "copy of this agent.\n") {
		t.Errorf("agent loader = %q", got[0].Content)
	}
}

func TestLoadersFromRejectsBadFrontMatter(t *testing.T) {
	for name, data := range map[string]string{
		"no front matter": "# Demo\n",
		"no description":  "---\nname: demo\n---\n",
		"unterminated":    "---\nname: demo\ndescription: d\n",
	} {
		if _, err := loadersFrom(fstest.MapFS{"content/skills/demo/SKILL.md": {Data: []byte(data)}}); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	if _, err := loadersFrom(fstest.MapFS{"content/agents/a.md": {Data: []byte("---\nname: a\ndescription: d\n---\n")}}); err == nil {
		t.Error("agent without tools: want error")
	}
}

func writeLoaders(t *testing.T, root string, crlf bool) []File {
	t.Helper()
	ls, err := loadersFrom(demoFS)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range ls {
		c := f.Content
		if crlf {
			c = strings.ReplaceAll(c, "\n", "\r\n")
		}
		p := filepath.Join(root, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ls
}

func TestWriteWritesLoaders(t *testing.T) {
	root := written(t)
	for _, f := range Loaders() {
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil || string(b) != f.Content {
			t.Errorf("%s: err=%v", f.Path, err)
		}
	}
}

func TestLoaderDiffIgnoresCR(t *testing.T) {
	root := t.TempDir()
	ls := writeLoaders(t, root, true)
	got, err := loaderDiffFrom(root, nil, ls)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestLoaderDiffMissingAndChanged(t *testing.T) {
	root := t.TempDir()
	ls := writeLoaders(t, root, false)
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(ls[0].Path))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(ls[1].Path)), []byte("edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := loaderDiffFrom(root, nil, ls)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"changed .claude/skills/demo/SKILL.md", "missing .claude/agents/pal-review-x.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestLoaderDiffNoExtras(t *testing.T) {
	root := t.TempDir()
	ls := writeLoaders(t, root, false)
	if err := os.WriteFile(filepath.Join(root, ".claude", "skills", "mine.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := loaderDiffFrom(root, nil, ls)
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestLoadersRespectOverrides(t *testing.T) {
	root := t.TempDir()
	ents, err := loaderEntries(demoFS)
	if err != nil {
		t.Fatal(err)
	}
	ov := []manifest.Override{{Item: "skills/demo"}, {Item: "agents/review-x.md"}}
	for _, e := range ents {
		p := filepath.Join(root, filepath.FromSlash(e.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("hand edited\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeLoaderEntries(root, ents, ov); err != nil {
		t.Fatal(err)
	}
	for _, e := range ents {
		b, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(e.Path)))
		if string(b) != "hand edited\n" {
			t.Errorf("%s overwritten", e.Path)
		}
	}
	if got, err := loaderDiffEntries(root, ov, ents); err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
	if got, _ := loaderDiffEntries(root, nil, ents); len(got) != 2 {
		t.Fatalf("without overrides got %v", got)
	}
	if err := writeLoaderEntries(root, ents, nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := loaderDiffEntries(root, nil, ents); len(got) != 0 {
		t.Fatalf("after rewrite %v", got)
	}
}
