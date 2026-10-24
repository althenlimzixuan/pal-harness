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
	"github.com/althenlimzixuan/pal-harness/internal/seq"
	"github.com/althenlimzixuan/pal-harness/manifest"
	"github.com/althenlimzixuan/pal-harness/number"
	"github.com/althenlimzixuan/pal-harness/pin"
	"github.com/althenlimzixuan/pal-harness/render"
	"github.com/althenlimzixuan/pal-harness/state"
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
  pal-harness render <file.md>
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
		return runPin(dir, stderr)
	case "render":
		if len(rest) != 1 {
			return usageErr(stderr, "render takes one Markdown file")
		}
		return runRender(rest[0], stderr)
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

// msgNotRoot is the hint for a command run outside the SSOT root.
const msgNotRoot = "run from the SSOT root (the folder with project.yaml)"

func isRoot(dir string) bool { return exists(filepath.Join(dir, "project.yaml")) }

// runPin writes the pinned copy, leaving the project's overridden items alone.
func runPin(dir string, stderr io.Writer) int {
	if !isRoot(dir) {
		return fail(stderr, msgNotRoot)
	}
	p, err := parseProject(dir)
	if err != nil {
		for _, f := range manifestFindings("project.yaml", err) {
			fmt.Fprintln(stderr, f)
		}
		return 2
	}
	if err := pin.WriteExcept(dir, p.Harness.Overrides); err != nil {
		return fail(stderr, "%v", err)
	}
	return 0
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
	if p != nil && p.Harness.Version != harness.Version {
		// The rest is judged by the pinned release, not by this one.
		fmt.Fprintf(stdout, "project.yaml: harness.version is %s but this checker is %s\n", p.Harness.Version, harness.Version)
		fmt.Fprintf(stdout, "run: go run github.com/althenlimzixuan/pal-harness/cmd/pal-harness@v%s check .\n", p.Harness.Version)
		return 1
	}
	var findings []string
	add := func(f string, a ...any) { findings = append(findings, fmt.Sprintf(f, a...)) }
	findings = append(findings, manifestFindings(mfile, err)...)
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
		ldiff, lerr := pin.LoaderDiff(dir, p.Harness.Overrides)
		if lerr != nil {
			fmt.Fprintf(stderr, ".claude: %v\n", lerr)
			return 2
		}
		findings = append(findings, ldiff...)
	}
	if mode == "project" {
		sf, serr := stateFindings(dir, p)
		if serr != nil {
			fmt.Fprintln(stderr, serr)
			return 2
		}
		findings = append(findings, sf...)
	}
	if mode == "project" && !callsCheck(dir) {
		add(".github/workflows/ssot.yml: must call %s with version %s", checkUses, harness.Version)
	}
	for _, f := range findings {
		fmt.Fprintln(stdout, f)
	}
	if p.HasFlowOverride("auto-promote") {
		fmt.Fprintln(stdout, "warning: auto-promote is on: sign-off and ship PRs merge without the owner")
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}

// stateFindings validates the optional doc/features.yaml and every
// specs/*/phase-state.yaml, prefixing each finding with its slash path from
// dir. Specs numbered below p's legacy_specs_before (when set) are skipped. With
// no valid manifest (p == nil) the flow overrides are unknown, so the shipped
// rule is skipped. A file that exists but cannot be read is an error (exit 2).
func stateFindings(dir string, p *manifest.Project) ([]string, error) {
	legacyBefore := ""
	features := func(b []byte) []string { _, f := state.ParseFeatures(b); return f }
	if p != nil {
		legacyBefore = p.Harness.LegacySpecsBefore
		features = func(b []byte) []string { _, f := state.ValidateFeaturesFor(b, p.Harness.FlowOverrides); return f }
	}
	var out []string
	check := func(rel string, parse func([]byte) []string) error {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		for _, f := range parse(b) {
			out = append(out, rel+": "+f)
		}
		return nil
	}
	if err := check("doc/features.yaml", features); err != nil {
		return nil, err
	}
	matches, err := filepath.Glob(filepath.Join(dir, "specs", "*", "phase-state.yaml"))
	if err != nil {
		return nil, err
	}
	for _, m := range matches {
		rel, err := filepath.Rel(dir, m)
		if err != nil {
			return nil, err
		}
		if isLegacySpec(filepath.Base(filepath.Dir(m)), legacyBefore) {
			continue
		}
		if err := check(filepath.ToSlash(rel), func(b []byte) []string { _, f := state.ParsePhaseState(b); return f }); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// isLegacySpec reports whether a spec folder "NNN-slug" is numbered below the
// project's legacy_specs_before marker. Both numbers are three digits.
func isLegacySpec(folder, before string) bool {
	if before == "" || len(folder) < 3 || folder[:3] >= before {
		return false
	}
	for _, c := range folder[:3] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// manifestFindings turns a manifest error into findings: one per validation
// problem ("<file>: <path>: <message>"), or one for a decode error.
func manifestFindings(file string, err error) []string {
	if err == nil {
		return nil
	}
	var ve *manifest.ValidationError
	if !errors.As(err, &ve) {
		return []string{file + ": " + err.Error()}
	}
	out := make([]string, len(ve.Problems))
	for i, pr := range ve.Problems {
		out[i] = file + ": " + pr
	}
	return out
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
		var ioe ioError
		if errors.As(err, &ioe) {
			fmt.Fprintf(stderr, "%s: %v\n", mfile, err)
			return 2
		}
		for _, f := range manifestFindings(mfile, err) {
			fmt.Fprintln(stderr, f)
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
		return "", fmt.Errorf("template %s/templates/%s not found; %s", pin.Dir, name, msgNotRoot)
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
	if !isRoot(dir) {
		return fail(stderr, msgNotRoot)
	}
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
	// ADRs merged to dev or origin/dev but not in this checkout count too.
	names, err := devEntries(dir, "doc/ADR/", "blob")
	if err != nil {
		return fail(stderr, "%v", err)
	}
	if m := seq.MaxADR(names) + 1; m > n {
		n = m
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

// devRefs are the branches whose trees count when numbering.
var devRefs = []string{"refs/heads/dev", "refs/remotes/origin/dev"}

// devEntries lists the entries of kind ("tree" or "blob") under path on each
// of devRefs that exists. Outside a git repo no ref exists, so it is empty.
func devEntries(dir, path, kind string) ([]string, error) {
	var names []string
	for _, ref := range devRefs {
		if _, err := git(dir, "rev-parse", "--verify", "-q", ref); err != nil {
			continue
		}
		out, err := git(dir, "ls-tree", "-z", ref, path)
		if err != nil {
			return nil, fmt.Errorf("git ls-tree %s %s failed: %v: %s", ref, path, err, out)
		}
		for _, entry := range strings.Split(out, "\x00") {
			meta, name, _ := strings.Cut(entry, "\t")
			if f := strings.Fields(meta); len(f) == 3 && f[1] == kind {
				names = append(names, name)
			}
		}
	}
	return names, nil
}

// writeFile creates the spec file; tests replace it to force a failure.
var writeFile = writeNew

func newSpec(dir, slug string, stdout, stderr io.Writer) int {
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
	n, err := number.NextSpec(dir)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	// Numbers already taken elsewhere: dev's and origin/dev's specs/ (the new
	// branch starts from dev) and every docs/NNN-* branch, local or remote,
	// so two specs in flight never share a number.
	names, err := devEntries(dir, "specs/", "tree")
	if err != nil {
		return fail(stderr, "%v", err)
	}
	// refs/remotes/*/docs (no trailing /*) matches no remote branch in git's
	// for-each-ref, so the remote pattern names the branches explicitly.
	refs, err := git(dir, "for-each-ref", "--format=%(refname:short)", "refs/heads/docs", "refs/remotes/*/docs/*")
	if err != nil {
		return fail(stderr, "git for-each-ref failed: %v: %s", err, refs)
	}
	names = append(names, strings.Fields(refs)...)
	if m := seq.MaxSpec(names) + 1; m > n {
		n = m
	}
	name := fmt.Sprintf("%03d-%s", n, slug)
	branch := "docs/" + name
	folder := filepath.Join(dir, "specs", name)
	if exists(folder) {
		return fail(stderr, "%s already exists", folder)
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

// runRender writes <file>.html beside the Markdown file. Diagram links are
// read relative to the Markdown file's folder; nothing is written on error.
func runRender(file string, stderr io.Writer) int {
	md, err := os.ReadFile(file)
	if err != nil {
		return fail(stderr, "render: %v", err)
	}
	base := filepath.Dir(file)
	page, err := render.Page(md, func(rel string) ([]byte, error) {
		return os.ReadFile(filepath.Join(base, filepath.FromSlash(rel)))
	})
	if err != nil {
		fmt.Fprintln(stderr, "render: "+err.Error())
		return 1
	}
	out := strings.TrimSuffix(file, filepath.Ext(file)) + ".html"
	if err := os.WriteFile(out, page, 0o644); err != nil {
		return fail(stderr, "render: %v", err)
	}
	return 0
}
