// Package pin writes and verifies the committed .pal-harness/ copy that every
// project carries. Kickstart writes it with Files; `pal-harness check`
// verifies it with Diff.
package pin

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	harness "github.com/althenlimzixuan/pal-harness"
	"github.com/althenlimzixuan/pal-harness/manifest"
	"gopkg.in/yaml.v3"
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
func Write(root string) error { return WriteExcept(root, nil) }

// WriteExcept is Write, but it leaves overridden items (and anything under
// them) untouched, so a recorded patch survives a re-pin.
// It also writes the skill and agent loaders (see Loaders).
func WriteExcept(root string, overrides []manifest.Override) error {
	for _, f := range Files() {
		if overridden(strings.TrimPrefix(f.Path, Dir+"/"), overrides) {
			continue
		}
		if err := writeFile(root, f); err != nil {
			return err
		}
	}
	return writeLoaderEntries(root, mustLoaderEntries(), overrides)
}

func writeFile(root string, f File) error {
	p := filepath.Join(root, filepath.FromSlash(f.Path))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(f.Content), 0o644)
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

// loaderEntry is a generated loader plus the pinned source it points at
// (Rel, relative to Dir), which is also the name overrides match against.
type loaderEntry struct {
	File
	Rel string
}

// Loaders returns the generated files Claude Code needs to find the pinned
// skills and agents: .claude/skills/<dir>/SKILL.md and .claude/agents/<name>.md.
// Each tells the model to read the pinned copy under Dir. It panics if the
// embedded content is malformed (it is fixed at build time).
func Loaders() []File {
	ls, err := loadersFrom(harness.Content)
	if err != nil {
		panic("pin: " + err.Error())
	}
	return ls
}

func mustLoaderEntries() []loaderEntry {
	es, err := loaderEntries(harness.Content)
	if err != nil {
		panic("pin: " + err.Error())
	}
	return es
}

func loadersFrom(fsys fs.FS) ([]File, error) {
	es, err := loaderEntries(fsys)
	if err != nil {
		return nil, err
	}
	out := make([]File, len(es))
	for i, e := range es {
		out[i] = e.File
	}
	return out, nil
}

func loaderEntries(fsys fs.FS) ([]loaderEntry, error) {
	var out []loaderEntry
	skills, err := fs.Glob(fsys, "content/skills/*/SKILL.md")
	if err != nil {
		return nil, err
	}
	for _, p := range skills {
		fm, err := frontMatter(fsys, p, "name", "description")
		if err != nil {
			return nil, err
		}
		rel := strings.TrimPrefix(p, "content/")
		dir := path.Base(path.Dir(p))
		out = append(out, loaderEntry{File{Path: ".claude/skills/" + dir + "/SKILL.md", Content: loaderText(fm.raw, rel, "skill")}, rel})
	}
	agents, err := fs.Glob(fsys, "content/agents/*.md")
	if err != nil {
		return nil, err
	}
	for _, p := range agents {
		fm, err := frontMatter(fsys, p, "name", "description", "tools")
		if err != nil {
			return nil, err
		}
		rel := strings.TrimPrefix(p, "content/")
		out = append(out, loaderEntry{File{Path: ".claude/agents/" + fm.name + ".md", Content: loaderText(fm.raw, rel, "agent")}, rel})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

type frontMatterInfo struct {
	raw  string // the lines between the --- lines, each ending in LF
	name string
}

// frontMatter reads the text between the first two --- lines of p, requiring
// every key in need to be a non-empty string.
func frontMatter(fsys fs.FS, p string, need ...string) (frontMatterInfo, error) {
	b, err := fs.ReadFile(fsys, p)
	if err != nil {
		return frontMatterInfo{}, err
	}
	lines := strings.Split(lf(string(b)), "\n")
	if lines[0] != "---" {
		return frontMatterInfo{}, fmt.Errorf("%s: no front matter", p)
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return frontMatterInfo{}, fmt.Errorf("%s: front matter is not closed", p)
	}
	raw := strings.Join(lines[1:end], "\n") + "\n"
	var m map[string]any
	if err := yaml.Unmarshal([]byte(raw), &m); err != nil {
		return frontMatterInfo{}, fmt.Errorf("%s: front matter: %w", p, err)
	}
	for _, k := range need {
		if v, ok := m[k].(string); !ok || strings.TrimSpace(v) == "" {
			return frontMatterInfo{}, fmt.Errorf("%s: front matter lacks %q", p, k)
		}
	}
	name, _ := m["name"].(string)
	return frontMatterInfo{raw: raw, name: name}, nil
}

func loaderText(front, rel, kind string) string {
	return "---\n" + front + "---\n\n" +
		"<!-- generated by pal-harness pin from " + Dir + "/" + rel + "; do not edit -->\n\n" +
		"Read `" + Dir + "/" + rel + "` and follow it exactly. It is the pinned harness copy of this " + kind + ".\n"
}

func writeLoaderEntries(root string, es []loaderEntry, overrides []manifest.Override) error {
	for _, e := range es {
		if overridden(e.Rel, overrides) {
			continue
		}
		if err := writeFile(root, e.File); err != nil {
			return err
		}
	}
	return nil
}

// LoaderDiff compares the loaders under root with Loaders(). Findings are
// sorted and read "missing <path>" or "changed <path>", with path relative to
// root. CR is ignored; overridden skills and agents are skipped. Loaders the
// project added itself are not reported.
func LoaderDiff(root string, overrides []manifest.Override) ([]string, error) {
	return loaderDiffEntries(root, overrides, mustLoaderEntries())
}

func loaderDiffFrom(root string, overrides []manifest.Override, files []File) ([]string, error) {
	es := make([]loaderEntry, len(files))
	for i, f := range files {
		es[i] = loaderEntry{File: f, Rel: f.Path}
	}
	return loaderDiffEntries(root, overrides, es)
}

func loaderDiffEntries(root string, overrides []manifest.Override, es []loaderEntry) ([]string, error) {
	findings := []string{}
	for _, e := range es {
		if overridden(e.Rel, overrides) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(e.Path)))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			findings = append(findings, "missing "+e.Path)
		case err != nil:
			return nil, err
		case lf(string(b)) != e.Content:
			findings = append(findings, "changed "+e.Path)
		}
	}
	sort.Strings(findings)
	return findings, nil
}
