// Package pin writes and verifies the committed .pal-harness/ copy that every
// project carries. Kickstart writes it with Files; `pal-harness check`
// verifies it with Diff.
package pin

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	harness "github.com/althenlimzixuan/pal-harness"
	"github.com/althenlimzixuan/pal-harness/manifest"
)

// Dir is the pinned copy's directory at a project's root.
const Dir = ".pal-harness"

// File is one file of the pinned copy. Path is slash-separated and starts with Dir + "/".
type File struct{ Path, Content string }

func lf(s string) string { return strings.ReplaceAll(s, "\r", "") }

// Files returns every embedded content file re-rooted at Dir, plus
// Dir/VERSION, sorted by Path, with LF line endings.
func Files() []File {
	var out []File
	err := fs.WalkDir(harness.Content, "content", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := harness.Content.ReadFile(p)
		if err != nil {
			return err
		}
		out = append(out, File{Path: Dir + "/" + strings.TrimPrefix(p, "content/"), Content: lf(string(b))})
		return nil
	})
	if err != nil {
		panic("pin: embedded content unreadable: " + err.Error())
	}
	out = append(out, File{Path: Dir + "/VERSION", Content: harness.Version + "\n"})
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// Write writes Files() under root, creating directories and overwriting
// existing files. It never deletes anything.
func Write(root string) error {
	for _, f := range Files() {
		p := filepath.Join(root, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(p, []byte(f.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func overridden(rel string, overrides []manifest.Override) bool {
	for _, o := range overrides {
		if rel == o.Item || strings.HasPrefix(rel, o.Item+"/") {
			return true
		}
	}
	return false
}

// Diff compares root/.pal-harness/ with Files(). Findings are sorted and
// read "missing <rel>", "changed <rel>" or "extra <rel>", with rel relative to
// Dir. CR is ignored; overridden items (and anything under them) are skipped.
// A missing Dir makes every file "missing"; only I/O errors are returned.
func Diff(root string, overrides []manifest.Override) ([]string, error) {
	findings := []string{}
	want := map[string]bool{}
	for _, f := range Files() {
		rel := strings.TrimPrefix(f.Path, Dir+"/")
		want[rel] = true
		if overridden(rel, overrides) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			findings = append(findings, "missing "+rel)
		case err != nil:
			return nil, err
		case lf(string(b)) != f.Content:
			findings = append(findings, "changed "+rel)
		}
	}

	base := filepath.Join(root, Dir)
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if p == base && errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		r, err := filepath.Rel(base, p)
		if err != nil {
			return err
		}
		rel := path.Clean(filepath.ToSlash(r))
		if rel != "." && !want[rel] && !overridden(rel, overrides) {
			findings = append(findings, "extra "+rel)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(findings)
	return findings, nil
}
