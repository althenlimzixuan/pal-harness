package fingerprint

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSpec(t *testing.T) {
	jobs, err := ParseSpec(strings.NewReader("# the jobs\n\napi: cmd internal/ go.mod\n  web:  web  # the site\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 || jobs[0].Name != "api" || strings.Join(jobs[0].Paths, ",") != "cmd,internal,go.mod" ||
		jobs[1].Name != "web" || strings.Join(jobs[1].Paths, ",") != "web" {
		t.Fatalf("%+v", jobs)
	}
}

func TestParseSpecErrors(t *testing.T) {
	for _, c := range []struct{ name, spec, want string }{
		{"empty", "# nothing\n", "no jobs"},
		{"no colon", "api cmd\n", "line 1"},
		{"bad name", "API: cmd\n", "job name"},
		{"no paths", "api:\n", "no paths"},
		{"duplicate", "api: cmd\napi: internal\n", "twice"},
		{"absolute", "api: /etc\n", "relative"},
		{"drive", "api: C:/x\n", "relative"},
		{"parent", "api: cmd/../..\n", "relative"},
		{"backslash", `api: cmd\x` + "\n", "forward slashes"},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseSpec(strings.NewReader(c.spec))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("err = %v, want it to mention %q", err, c.want)
			}
		})
	}
}

func TestCheckPath(t *testing.T) {
	for _, ok := range []string{".", "cmd", "doc/openapi.yaml", ".github/workflows/ci.yml"} {
		if _, err := CheckPath(ok); err != nil {
			t.Errorf("%q: %v", ok, err)
		}
	}
	if p, _ := CheckPath("internal/"); p != "internal" {
		t.Errorf("internal/ -> %q", p)
	}
	for _, bad := range []string{"", "..", "/x", "a/../../b", `a\b`, "C:/x"} {
		if _, err := CheckPath(bad); err == nil {
			t.Errorf("%q: want an error", bad)
		}
	}
}

func TestComputeForm(t *testing.T) {
	ids := map[string]string{"cmd": "aaa", ".github/workflows/ci.yml": "ccc"}
	lookup := func(p string) (string, error) { return ids[p], nil }
	got, err := Compute([]Job{{Name: "api", Paths: []string{"cmd", "go.mod"}}}, []string{".github/workflows/ci.yml"}, lookup)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha1.Sum([]byte("api\ncmd aaa\ngo.mod missing\n.github/workflows/ci.yml ccc"))
	if len(got) != 1 || got[0].Name != "api" || got[0].Sum != hex.EncodeToString(sum[:]) {
		t.Fatalf("%+v", got)
	}
}

func TestComputeLookupError(t *testing.T) {
	lookup := func(string) (string, error) { return "", os.ErrPermission }
	if _, err := Compute([]Job{{Name: "api", Paths: []string{"cmd"}}}, nil, lookup); err == nil {
		t.Fatal("want the lookup error")
	}
}

// repo is a throwaway git repository for the fingerprint cases.
type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) *repo {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	r := &repo{t, t.TempDir()}
	r.git("init", "-q")
	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com",
		"-c", "core.autocrlf=false", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = r.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// commit applies the files (nil deletes) and commits, even when nothing changed.
func (r *repo) commit(files map[string]*string) {
	r.t.Helper()
	for p, body := range files {
		full := filepath.Join(r.dir, filepath.FromSlash(p))
		if body == nil {
			if err := os.Remove(full); err != nil {
				r.t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(*body), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
	r.git("add", "-A")
	r.git("commit", "-q", "--allow-empty", "-m", "c")
}

func (r *repo) sums() map[string]string {
	r.t.Helper()
	lookup, err := NewGit(r.dir, "HEAD")
	if err != nil {
		r.t.Fatal(err)
	}
	jobs := []Job{{Name: "daemon", Paths: []string{"daemon", "doc/openapi.yaml"}}, {Name: "desktop", Paths: []string{"desktop"}}}
	got, err := Compute(jobs, []string{".github/workflows/ci.yml"}, lookup)
	if err != nil {
		r.t.Fatal(err)
	}
	out := map[string]string{}
	for _, s := range got {
		out[s.Name] = s.Sum
	}
	return out
}

func s(v string) *string { return &v }

// The cases of Pal Center's ci_fingerprint.json: a fingerprint changes only when a file its job reads changes.
func TestFingerprintCases(t *testing.T) {
	base := map[string]*string{
		"daemon/main.go":                        s("package main\n"),
		"desktop/src/App.tsx":                   s("export {}\n"),
		"doc/openapi.yaml":                      s("openapi: 3.0.0\n"),
		"doc/API.md":                            s("# API\n"),
		"specs/003-start-work/phase-state.yaml": s("stage: shipping\n"),
		"README.md":                             s("# Pal Center\n"),
		".github/workflows/ci.yml":              s("name: ci\n"),
	}
	for _, c := range []struct {
		name            string
		change          map[string]*string
		daemon, desktop bool // true: changed
	}{
		{"docs and specs only", map[string]*string{"specs/003-start-work/phase-state.yaml": s("stage: done\n"),
			"doc/API.md": s("# API v2\n"), "README.md": s("# Pal Center 0.8.0\n")}, false, false},
		{"same tree (a merge or a promotion)", map[string]*string{}, false, false},
		{"daemon change", map[string]*string{"daemon/main.go": s("package main\n\nfunc main() {}\n")}, true, false},
		{"desktop change", map[string]*string{"desktop/src/App.tsx": s("export const x = 1\n")}, false, true},
		{"contract copy", map[string]*string{"doc/openapi.yaml": s("openapi: 3.1.0\n")}, true, false},
		{"workflow change", map[string]*string{".github/workflows/ci.yml": s("name: ci\n# changed\n")}, true, true},
		{"new file", map[string]*string{"desktop/src/New.tsx": s("export {}\n")}, false, true},
		{"deleted file", map[string]*string{"daemon/main.go": nil, "daemon/other.go": s("package main\n")}, true, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := newRepo(t)
			r.commit(base)
			before := r.sums()
			r.commit(c.change)
			after := r.sums()
			if (before["daemon"] != after["daemon"]) != c.daemon || (before["desktop"] != after["desktop"]) != c.desktop {
				t.Fatalf("daemon changed %v (want %v), desktop changed %v (want %v)",
					before["daemon"] != after["daemon"], c.daemon, before["desktop"] != after["desktop"], c.desktop)
			}
		})
	}
}

func TestGitMissingAndRoot(t *testing.T) {
	r := newRepo(t)
	r.commit(map[string]*string{"a.txt": s("a\n")})
	lookup, err := NewGit(r.dir, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if id, err := lookup("nope"); err != nil || id != "" {
		t.Fatalf("missing path: %q %v", id, err)
	}
	root, err := lookup(".")
	if err != nil || root != strings.TrimSpace(r.git("rev-parse", "HEAD^{tree}")) {
		t.Fatalf("root: %q %v", root, err)
	}
}

func TestNewGitErrors(t *testing.T) {
	r := newRepo(t)
	if _, err := NewGit(r.dir, "HEAD"); err == nil {
		t.Fatal("a repo with no commit: want an error")
	}
	if _, err := NewGit(t.TempDir(), "HEAD"); err == nil {
		t.Fatal("not a repo: want an error")
	}
}
