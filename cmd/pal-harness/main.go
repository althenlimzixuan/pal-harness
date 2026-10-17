// Command pal-harness checks and maintains a project's SSOT: generated
// blocks, the pinned harness copy, and spec/ADR numbering.
package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	harness "github.com/althenlimzixuan/pal-harness"
	"github.com/althenlimzixuan/pal-harness/blocks"
	"github.com/althenlimzixuan/pal-harness/manifest"
	"github.com/althenlimzixuan/pal-harness/number"
	"github.com/althenlimzixuan/pal-harness/pin"
	"gopkg.in/yaml.v3"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

// now is the clock; tests replace it.
var now = time.Now

const usage = `usage:
  pal-harness check [dir]
  pal-harness generate [dir]
  pal-harness pin [dir]
  pal-harness spec new <slug> [dir]
  pal-harness adr new <slug> [--conflict] [dir]
  pal-harness version
`

func usageErr(stderr io.Writer, msg string) int {
	if msg != "" {
		fmt.Fprintln(stderr, msg)
	}
	fmt.Fprint(stderr, usage)
	return 2
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return usageErr(stderr, "")
	}
	rest := args[1:]
	switch args[0] {
	case "version":
		if len(rest) != 0 {
			return usageErr(stderr, "version takes no arguments")
		}
		fmt.Fprintln(stdout, harness.Version)
		return 0
	case "check", "generate", "pin":
		if len(rest) > 1 {
			return usageErr(stderr, args[0]+" takes at most one directory")
		}
		dir := dirArg(rest)
		switch args[0] {
		case "check":
			return runCheck(dir, stdout, stderr)
		case "generate":
			return runGenerate(dir, stderr)
		}
		if err := pin.Write(dir); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		return 0
	case "spec", "adr":
		if len(rest) == 0 || rest[0] != "new" {
			return usageErr(stderr, "expected: "+args[0]+" new <slug>")
		}
		return runNew(args[0], rest[1:], stdout, stderr)
	}
	return usageErr(stderr, "unknown command: "+args[0])
}

func dirArg(a []string) string {
	if len(a) > 0 {
		return a[0]
	}
	return "."
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// docBlock names a document and the generated block it must carry.
type docBlock struct{ file, name, source, block string }

// ioError marks a file that could not be read, as opposed to one that is
// invalid; it makes the commands exit 2 rather than report a finding.
type ioError struct{ error }

func parseProject(dir string) (*manifest.Project, error) {
	b, err := os.ReadFile(filepath.Join(dir, "project.yaml"))
	if err != nil {
		return nil, ioError{err}
	}
	return manifest.ParseProject(b)
}

// targets returns the blocks to maintain in dir, whichever mode it is in.
// Project mode wins when both manifests exist. mode is "" when neither manifest exists; a manifest parse error is returned
// with the manifest's file name.
func targets(dir string) (mode string, p *manifest.Project, docs []docBlock, file string, err error) {
	switch {
	case exists(filepath.Join(dir, "project.yaml")):
		p, err = parseProject(dir)
		if err != nil {
			return "project", nil, nil, "project.yaml", err
		}
		t := blocks.ComponentsTable(p)
		return "project", p, []docBlock{
			{"README.md", blocks.Components, "project.yaml", t},
			{"CLAUDE.md", blocks.Components, "project.yaml", t},
		}, "", nil
	case exists(filepath.Join(dir, "workspace.yaml")):
		b, rerr := os.ReadFile(filepath.Join(dir, "workspace.yaml"))
		var w *manifest.Workspace
		if rerr != nil {
			rerr = ioError{rerr}
		} else {
			w, rerr = manifest.ParseWorkspace(b)
		}
		if rerr != nil {
			return "workspace", nil, nil, "workspace.yaml", rerr
		}
		return "workspace", nil, []docBlock{{"CLAUDE.md", blocks.Registry, "workspace.yaml", blocks.RegistryTable(w)}}, "", nil
	}
	return "", nil, nil, "", nil
}

func runCheck(dir string, stdout, stderr io.Writer) int {
	mode, p, docs, mfile, err := targets(dir)
	if mode == "" {
		fmt.Fprintln(stderr, "no project.yaml or workspace.yaml in "+dir)
		return 2
	}
	var ioe ioError
	if errors.As(err, &ioe) {
		fmt.Fprintf(stderr, "%s: %v\n", mfile, err)
		return 2
	}
	var findings []string
	add := func(f string, a ...any) { findings = append(findings, fmt.Sprintf(f, a...)) }
	if err != nil {
		add("%s: %v", mfile, err)
	}
	if p != nil && p.Harness.Version != harness.Version {
		add("project.yaml: harness.version is %s but this checker is %s", p.Harness.Version, harness.Version)
	}
	for _, d := range docs {
		doc, rerr := os.ReadFile(filepath.Join(dir, d.file))
		if rerr != nil {
			add("%s: %s block: %v", d.file, d.name, rerr)
			continue
		}
		ok, merr := blocks.Matches(string(doc), d.name, d.block)
		switch {
		case merr != nil:
			add("%s: %s block: %v", d.file, d.name, merr)
		case !ok:
			add("%s: %s block does not match %s", d.file, d.name, d.source)
		}
	}
	if p != nil {
		diff, derr := pin.Diff(dir, p.Harness.Overrides)
		if derr != nil {
			fmt.Fprintf(stderr, "%s: %v\n", pin.Dir, derr)
			return 2
		}
		for _, f := range diff {
			add("%s: %s", pin.Dir, f)
		}
	}
	if mode == "project" && !callsCheck(dir) {
		add(".github/workflows/ssot.yml: must call %s with version %s", checkUses, harness.Version)
	}
	for _, f := range findings {
		fmt.Fprintln(stdout, f)
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}

// checkUses is the reusable workflow a project's ssot.yml must call.
const checkUses = "althenlimzixuan/pal-harness/.github/workflows/check.yml@v" + harness.Version

// callsCheck reports whether dir/.github/workflows/ssot.yml parses as YAML and
// has a job that uses exactly checkUses with `with.version` exactly Version.
// A missing or unparsable file does not.
func callsCheck(dir string) bool {
	b, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "ssot.yml"))
	if err != nil {
		return false
	}
	var wf struct {
		Jobs map[string]struct {
			Uses string         `yaml:"uses"`
			With map[string]any `yaml:"with"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(b, &wf); err != nil {
		return false
	}
	for _, j := range wf.Jobs {
		if v, ok := j.With["version"].(string); ok && j.Uses == checkUses && v == harness.Version {
			return true
		}
	}
	return false
}

func runGenerate(dir string, stderr io.Writer) int {
	mode, _, docs, mfile, err := targets(dir)
	if mode == "" {
		fmt.Fprintln(stderr, "no project.yaml or workspace.yaml in "+dir)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "%s: %v\n", mfile, err)
		var ioe ioError
		if errors.As(err, &ioe) {
			return 2
		}
		return 1
	}
	// Compute every result before writing any file.
	type out struct{ path, text string }
	var outs []out
	failed := false
	for _, d := range docs {
		path := filepath.Join(dir, d.file)
		doc, rerr := os.ReadFile(path)
		if rerr != nil {
			fmt.Fprintf(stderr, "%s: %v\n", d.file, rerr)
			return 2
		}
		text, rerr := blocks.Replace(string(doc), d.name, d.block)
		if rerr != nil {
			fmt.Fprintf(stderr, "%s: %s block: %v\n", d.file, d.name, rerr)
			failed = true
			continue
		}
		outs = append(outs, out{path, text})
	}
	if failed {
		return 1
	}
	for _, o := range outs {
		if err := os.WriteFile(o.path, []byte(o.text), 0o644); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	}
	return 0
}

// title turns a slug into a title: dashes become spaces, first letter upper-cased.
func title(slug string) string {
	s := strings.ReplaceAll(slug, "-", " ")
	return strings.ToUpper(s[:1]) + s[1:]
}

func readTemplate(dir, name string) (string, error) {
	b, err := os.ReadFile(filepath.Join(dir, pin.Dir, "templates", name))
	if errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("template %s/templates/%s not found; run `pal-harness pin`", pin.Dir, name)
	}
	return string(b), err
}

func runNew(kind string, args []string, stdout, stderr io.Writer) int {
	var pos []string
	conflict := false
	for _, a := range args {
		switch {
		case a == "--conflict" && kind == "adr":
			conflict = true
		case strings.HasPrefix(a, "--"):
			return usageErr(stderr, "unknown flag: "+a)
		default:
			pos = append(pos, a)
		}
	}
	if len(pos) == 0 || len(pos) > 2 {
		return usageErr(stderr, "expected: "+kind+" new <slug> [dir]")
	}
	slug := pos[0]
	if !number.ValidSlug(slug) {
		return usageErr(stderr, "invalid slug (lowercase letters, digits and single dashes, at most 50 characters): "+slug)
	}
	dir := dirArg(pos[1:])
	if kind == "adr" {
		return newADR(dir, slug, conflict, stdout, stderr)
	}
	return newSpec(dir, slug, stdout, stderr)
}

func fail(stderr io.Writer, f string, a ...any) int {
	fmt.Fprintf(stderr, f+"\n", a...)
	return 2
}

func newADR(dir, slug string, conflict bool, stdout, stderr io.Writer) int {
	tpl := "adr.md"
	if conflict {
		tpl = "adr-conflict.md"
	}
	text, err := readTemplate(dir, tpl)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	n, err := number.NextADR(dir)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	num := fmt.Sprintf("%04d", n)
	text = strings.NewReplacer("<NNNN>", num, "<Title>", title(slug), "<Date>", now().Format("2006-01-02")).Replace(text)
	path := filepath.Join(dir, "doc", "ADR", num+"-"+slug+".md")
	if err := writeNew(path, text); err != nil {
		return fail(stderr, "%v", err)
	}
	fmt.Fprintln(stdout, path)
	return 0
}

// writeNew creates path (and parents) and refuses to overwrite.
func writeNew(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// git runs git with exec args (never a shell) in dir.
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// writeFile creates the spec file; tests replace it to force a failure.
var writeFile = writeNew

func newSpec(dir, slug string, stdout, stderr io.Writer) int {
	if !exists(filepath.Join(dir, "project.yaml")) {
		return fail(stderr, "no project.yaml in %s", dir)
	}
	st, err := git(dir, "status", "--porcelain")
	if err != nil {
		return fail(stderr, "git status failed: %v: %s", err, st)
	}
	if st != "" {
		return fail(stderr, "working tree is not clean; commit or stash first")
	}
	if _, err := git(dir, "rev-parse", "--verify", "-q", "refs/heads/dev"); err != nil {
		return fail(stderr, "local branch dev does not exist")
	}
	// The new branch starts from dev, so dev's specs/ counts as well as ours.
	tree, err := git(dir, "ls-tree", "-d", "--name-only", "dev", "specs/")
	if err != nil {
		return fail(stderr, "git ls-tree failed: %v: %s", err, tree)
	}
	var devSpecs []string
	if tree != "" {
		devSpecs = strings.Split(tree, "\n")
	}
	n, err := number.NextSpec(dir)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	if m := number.MaxSpecInNames(devSpecs) + 1; m > n {
		n = m
	}
	name := fmt.Sprintf("%03d-%s", n, slug)
	branch := "docs/" + name
	if _, err := git(dir, "rev-parse", "--verify", "-q", "refs/heads/"+branch); err == nil {
		return fail(stderr, "branch %s already exists", branch)
	}
	folder := filepath.Join(dir, "specs", name)
	if exists(folder) {
		return fail(stderr, "%s already exists", folder)
	}
	for _, d := range devSpecs {
		if strings.TrimSpace(d) == "specs/"+name {
			return fail(stderr, "specs/%s already exists on dev", name)
		}
	}
	text, err := readTemplate(dir, "spec.md")
	if err != nil {
		return fail(stderr, "%v", err)
	}
	orig, _ := git(dir, "branch", "--show-current")
	back := []string{"switch", orig}
	if orig == "" {
		sha, _ := git(dir, "rev-parse", "HEAD")
		back = []string{"switch", "--detach", sha}
	}
	if out, err := git(dir, "switch", "-c", branch, "dev"); err != nil {
		return fail(stderr, "git switch failed: %v: %s", err, out)
	}
	text = strings.NewReplacer("<NNN>", fmt.Sprintf("%03d", n), "<Title>", title(slug)).Replace(text)
	path := filepath.Join(folder, "spec.md")
	if err := writeFile(path, text); err != nil {
		os.RemoveAll(folder)
		git(dir, back...)
		git(dir, "branch", "-D", branch)
		return fail(stderr, "%v (rolled back to %s)", err, orig)
	}
	fmt.Fprintln(stdout, path)
	return 0
}
