package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	harness "github.com/althenlimzixuan/pal-harness"
	"github.com/althenlimzixuan/pal-harness/blocks"
	"github.com/althenlimzixuan/pal-harness/manifest"
	"github.com/althenlimzixuan/pal-harness/pin"
)

func do(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var o, e bytes.Buffer
	code := run(args, &o, &e)
	return code, o.String(), e.String()
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// fixture builds a clean project SSOT: project.yaml, README.md, CLAUDE.md,
// pinned copy and ssot.yml.
func fixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src, err := os.ReadFile("../../manifest/testdata/project.yaml")
	if err != nil {
		t.Fatal(err)
	}
	p, err := manifest.ParseProject(src)
	if err != nil {
		t.Fatal(err)
	}
	p.Harness.Version = harness.Version
	out, err := manifest.MarshalProject(p)
	if err != nil {
		t.Fatal(err)
	}
	write(t, dir, "project.yaml", string(out))
	doc := "# Doc\n\nSome prose.\n\n" + blocks.ComponentsTable(p) + "\n\nMore prose.\n"
	write(t, dir, "README.md", doc)
	write(t, dir, "CLAUDE.md", doc)
	if err := pin.Write(dir); err != nil {
		t.Fatal(err)
	}
	write(t, dir, ".github/workflows/ssot.yml", daemonSSOT)
	return dir
}

// daemonSSOT is the exact ssot.yml the Pal Center daemon writes at kickstart.
const daemonSSOT = "name: ssot\non:\n  push:\n    branches: [dev, staging, main]\n  pull_request:\n    branches: [dev, staging, main]\n" +
	"jobs:\n  check:\n    uses: althenlimzixuan/pal-harness/.github/workflows/check.yml@v" + harness.Version +
	"\n    with:\n      version: " + harness.Version + "\n"

const ssotFinding = ".github/workflows/ssot.yml: must call althenlimzixuan/pal-harness/.github/workflows/check.yml@v" +
	harness.Version + " with version " + harness.Version

// TestCheckCRLFProject is what a Windows checkout with core.autocrlf=true
// looks like: every file CRLF. check must still pass.
func TestCheckCRLFProject(t *testing.T) {
	dir := fixture(t)
	n := 0
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		n++
		return os.WriteFile(p, []byte(strings.ReplaceAll(string(b), "\n", "\r\n")), 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if n < 10 || !strings.Contains(read(t, dir, "project.yaml"), "\r\n") || !strings.Contains(read(t, dir, ".github/workflows/ssot.yml"), "\r\n") {
		t.Fatalf("fixture not converted (%d files)", n)
	}
	if code := run([]string{"check", dir}, io.Discard, io.Discard); code != 0 {
		_, out, errs := do(t, "check", dir)
		t.Fatalf("code %d out=%q err=%q", code, out, errs)
	}
}

func TestCheckSSOTWorkflowRejects(t *testing.T) {
	v := harness.Version
	uses := "althenlimzixuan/pal-harness/.github/workflows/check.yml@v" + v
	cases := map[string]string{
		"longer version tag": strings.Replace(daemonSSOT, uses, uses+"0", 1),
		"comment only":       "# uses: " + uses + "\n# with:\n#   version: " + v + "\njobs: {}\n",
		"fork path":          strings.Replace(daemonSSOT, "althenlimzixuan/pal-harness", "someone/pal-harness", 1),
		"version mismatch":   strings.Replace(daemonSSOT, "version: "+v+"\n", "version: 9.9.9\n", 1),
		"no with":            strings.Replace(daemonSSOT, "\n    with:\n      version: "+v, "", 1),
		"not yaml":           "jobs: [\n",
		"jobs not a map":     "jobs: [a, b]\n",
	}
	for name, wf := range cases {
		t.Run(name, func(t *testing.T) {
			if wf == daemonSSOT {
				t.Fatal("case did not change the file")
			}
			dir := fixture(t)
			write(t, dir, ".github/workflows/ssot.yml", wf)
			code, out, errs := do(t, "check", dir)
			if code != 1 || out != ssotFinding+"\n" {
				t.Fatalf("code %d out=%q err=%q", code, out, errs)
			}
		})
	}
	t.Run("missing", func(t *testing.T) {
		dir := fixture(t)
		if err := os.Remove(filepath.Join(dir, ".github", "workflows", "ssot.yml")); err != nil {
			t.Fatal(err)
		}
		if code, out, _ := do(t, "check", dir); code != 1 || out != ssotFinding+"\n" {
			t.Fatalf("code %d out=%q", code, out)
		}
	})
}

func TestCheckSSOTWorkflowAcceptsDaemonFile(t *testing.T) {
	dir := fixture(t)
	if got := read(t, dir, ".github/workflows/ssot.yml"); got != daemonSSOT {
		t.Fatalf("fixture ssot.yml %q", got)
	}
	if code, out, errs := do(t, "check", dir); code != 0 {
		t.Fatalf("daemon file: code %d out=%q err=%q", code, out, errs)
	}
	// Another job alongside the check job is fine.
	write(t, dir, ".github/workflows/ssot.yml", daemonSSOT+"  lint:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n")
	if code, out, errs := do(t, "check", dir); code != 0 {
		t.Fatalf("extra job: code %d out=%q err=%q", code, out, errs)
	}
}

func TestCheckCleanProject(t *testing.T) {
	dir := fixture(t)
	code, out, errs := do(t, "check", dir)
	if code != 0 {
		t.Fatalf("code %d out=%q err=%q", code, out, errs)
	}
}

func TestCheckFindings(t *testing.T) {
	dir := fixture(t)
	readme := read(t, dir, "README.md")
	write(t, dir, "README.md", strings.Replace(readme, "provider", "providerx", 1))
	write(t, dir, ".pal-harness/practices/workflow.md", "changed\n")
	write(t, dir, ".github/workflows/ssot.yml", "nothing\n")
	code, out, _ := do(t, "check", dir)
	if code != 1 {
		t.Fatalf("code %d", code)
	}
	for _, want := range []string{
		"README.md: components block does not match project.yaml",
		".pal-harness: changed practices/workflow.md",
		ssotFinding,
	} {
		if !strings.Contains(out, want+"\n") {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "CLAUDE.md:") {
		t.Errorf("CLAUDE.md should be fine:\n%s", out)
	}
}

func TestCheckVersionMismatchReportsOnlyVersion(t *testing.T) {
	dir := fixture(t)
	proj := read(t, dir, "project.yaml")
	write(t, dir, "project.yaml", strings.Replace(proj, "version: "+harness.Version, "version: 0.9.0", 1))
	write(t, dir, "README.md", "no block\n")
	write(t, dir, ".pal-harness/practices/workflow.md", "changed\n")
	write(t, dir, ".github/workflows/ssot.yml", "nothing\n")
	code, out, _ := do(t, "check", dir)
	want := "project.yaml: harness.version is 0.9.0 but this checker is " + harness.Version + "\n" +
		"run: go run github.com/althenlimzixuan/pal-harness/cmd/pal-harness@v0.9.0 check .\n"
	if code != 1 || out != want {
		t.Fatalf("code %d out:\n%s", code, out)
	}
}

func TestCheckValidationOneFindingPerProblem(t *testing.T) {
	dir := fixture(t)
	proj := read(t, dir, "project.yaml")
	proj = strings.Replace(proj, "version: 1\n", "version: 3\n", 1)
	proj = strings.Replace(proj, "key: fhph", "key: X", 1)
	write(t, dir, "project.yaml", proj)
	code, out, _ := do(t, "check", dir)
	want := "project.yaml: version: must be 1\nproject.yaml: key: must match ^[a-z0-9]{2,6}$\n"
	if code != 1 || out != want {
		t.Fatalf("code %d out:\n%s", code, out)
	}
	code, _, errs := do(t, "generate", dir)
	if code != 1 || errs != want {
		t.Fatalf("generate: code %d err:\n%s", code, errs)
	}
}

func TestCheckOverrideSkipsPin(t *testing.T) {
	dir := fixture(t)
	write(t, dir, ".pal-harness/practices/workflow.md", "changed\n")
	proj := read(t, dir, "project.yaml")
	if !strings.Contains(proj, "overrides: []") {
		t.Fatalf("fixture lacks empty overrides:\n%s", proj)
	}
	proj = strings.Replace(proj, "overrides: []", "overrides:\n    - item: practices/workflow.md\n      source: local\n      reason: r\n      since: 2026-10-17", 1)
	write(t, dir, "project.yaml", proj)
	code, out, _ := do(t, "check", dir)
	if code != 0 {
		t.Fatalf("code %d out=%s", code, out)
	}
}

func TestCheckMissingAndBrokenBlocks(t *testing.T) {
	dir := fixture(t)
	write(t, dir, "README.md", "no block here\n")
	write(t, dir, "CLAUDE.md", read(t, dir, "CLAUDE.md")+"<!-- pal:begin components (generated from project.yaml; do not edit) -->\n")
	code, out, _ := do(t, "check", dir)
	if code != 1 {
		t.Fatalf("code %d", code)
	}
	if !strings.Contains(out, "README.md: components block: ") || !strings.Contains(out, "CLAUDE.md: components block: ") {
		t.Errorf("out:\n%s", out)
	}
}

func TestCheckMissingFile(t *testing.T) {
	dir := fixture(t)
	if err := os.Remove(filepath.Join(dir, "README.md")); err != nil {
		t.Fatal(err)
	}
	code, out, _ := do(t, "check", dir)
	if code != 1 || !strings.Contains(out, "README.md: components block: ") {
		t.Fatalf("code %d out=%s", code, out)
	}
}

func TestCheckProjectParseError(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "project.yaml", "version: 1\nbogus: [\n")
	code, out, _ := do(t, "check", dir)
	if code != 1 || !strings.HasPrefix(out, "project.yaml: ") {
		t.Fatalf("code %d out=%q", code, out)
	}
}

func TestCheckNeitherManifest(t *testing.T) {
	code, _, errs := do(t, "check", t.TempDir())
	if code != 2 || errs == "" {
		t.Fatalf("code %d err=%q", code, errs)
	}
}

func wsFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src, err := os.ReadFile("../../manifest/testdata/workspace.yaml")
	if err != nil {
		t.Fatal(err)
	}
	w, err := manifest.ParseWorkspace(src)
	if err != nil {
		t.Fatal(err)
	}
	write(t, dir, "workspace.yaml", string(src))
	write(t, dir, "CLAUDE.md", "# WS\n\n"+blocks.RegistryTable(w)+"\n")
	return dir
}

func TestCheckWorkspace(t *testing.T) {
	dir := wsFixture(t)
	if code, out, _ := do(t, "check", dir); code != 0 {
		t.Fatalf("code %d out=%s", code, out)
	}
	write(t, dir, "CLAUDE.md", strings.Replace(read(t, dir, "CLAUDE.md"), "etp", "etpx", 1))
	code, out, _ := do(t, "check", dir)
	if code != 1 || !strings.Contains(out, "CLAUDE.md: registry block does not match workspace.yaml\n") {
		t.Fatalf("code %d out=%s", code, out)
	}
	write(t, dir, "CLAUDE.md", "none\n")
	code, out, _ = do(t, "check", dir)
	if code != 1 || !strings.Contains(out, "CLAUDE.md: registry block: ") {
		t.Fatalf("code %d out=%s", code, out)
	}
	write(t, dir, "workspace.yaml", "version: 1\nzzz: [\n")
	code, out, _ = do(t, "check", dir)
	if code != 1 || !strings.HasPrefix(out, "workspace.yaml: ") {
		t.Fatalf("code %d out=%s", code, out)
	}
}

func TestGenerateProject(t *testing.T) {
	dir := fixture(t)
	for _, f := range []string{"README.md", "CLAUDE.md"} {
		write(t, dir, f, strings.Replace(read(t, dir, f), "provider", "providerx", 1))
	}
	if code, out, errs := do(t, "generate", dir); code != 0 {
		t.Fatalf("code %d %s %s", code, out, errs)
	}
	if code, out, _ := do(t, "check", dir); code != 0 {
		t.Fatalf("check after generate: %d %s", code, out)
	}
}

func TestGenerateWorkspace(t *testing.T) {
	dir := wsFixture(t)
	write(t, dir, "CLAUDE.md", strings.Replace(read(t, dir, "CLAUDE.md"), "etp", "etpx", 1))
	if code, _, _ := do(t, "generate", dir); code != 0 {
		t.Fatal(code)
	}
	if code, _, _ := do(t, "check", dir); code != 0 {
		t.Fatal(code)
	}
}

func TestGenerateRefusesBadBlock(t *testing.T) {
	dir := fixture(t)
	stale := strings.Replace(read(t, dir, "README.md"), "provider", "providerx", 1)
	write(t, dir, "README.md", stale)
	bad := "<!-- pal:begin components (generated from project.yaml; do not edit) -->\n" + read(t, dir, "CLAUDE.md")
	write(t, dir, "CLAUDE.md", bad)
	code, _, _ := do(t, "generate", dir)
	if code != 1 {
		t.Fatalf("code %d", code)
	}
	if read(t, dir, "CLAUDE.md") != bad || read(t, dir, "README.md") != stale {
		t.Fatal("files were modified")
	}
}

func TestGenerateNoManifestAndNoBlock(t *testing.T) {
	if code, _, _ := do(t, "generate", t.TempDir()); code != 2 {
		t.Fatalf("code %d", code)
	}
	dir := fixture(t)
	write(t, dir, "README.md", "plain\n")
	if code, _, _ := do(t, "generate", dir); code != 1 {
		t.Fatalf("code %d", code)
	}
	write(t, dir, "project.yaml", "version: 1\nx: [\n")
	if code, _, _ := do(t, "generate", dir); code != 1 {
		t.Fatalf("code %d", code)
	}
}

func TestGenerateKeepsCRLF(t *testing.T) {
	dir := fixture(t)
	doc := strings.ReplaceAll(read(t, dir, "README.md"), "\n", "\r\n")
	doc = strings.Replace(doc, "provider", "providerx", 1)
	write(t, dir, "README.md", doc)
	if code, _, _ := do(t, "generate", dir); code != 0 {
		t.Fatal(code)
	}
	got := read(t, dir, "README.md")
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Fatal("mixed line endings")
	}
}

func TestPinAndVersion(t *testing.T) {
	dir := fixture(t)
	if err := os.RemoveAll(filepath.Join(dir, ".pal-harness")); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := do(t, "pin", dir); code != 0 {
		t.Fatal(code, errs)
	}
	if got := read(t, dir, ".pal-harness/VERSION"); got != harness.Version+"\n" {
		t.Fatalf("VERSION %q", got)
	}
	code, out, _ := do(t, "version")
	if code != 0 || out != harness.Version+"\n" {
		t.Fatalf("%d %q", code, out)
	}
}

func TestUsageErrors(t *testing.T) {
	cases := [][]string{
		{}, {"bogus"}, {"spec"}, {"spec", "new"}, {"spec", "new", "BAD"}, {"spec", "old", "x"},
		{"adr", "new"}, {"adr", "new", "Bad_Slug"}, {"adr"}, {"adr", "new", "x", "--nope"},
		{"check", "a", "b"}, {"version", "x"},
	}
	for _, c := range cases {
		code, _, errs := do(t, c...)
		if code != 2 || !strings.Contains(errs, "usage") {
			t.Errorf("%v: code %d err=%q", c, code, errs)
		}
	}
}

func tgit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func gitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	tgit(t, dir, "init", "-b", "main")
	write(t, dir, "project.yaml", "version: 1\n")
	write(t, dir, ".pal-harness/templates/spec.md", "# <Title>\n\nSpec <NNN>\n")
	tgit(t, dir, "add", "-A")
	tgit(t, dir, "commit", "-m", "init")
	tgit(t, dir, "branch", "dev")
	return dir
}

func TestSpecNew(t *testing.T) {
	dir := gitRepo(t)
	code, out, errs := do(t, "spec", "new", "first-spec", dir)
	if code != 0 {
		t.Fatalf("code %d err=%s", code, errs)
	}
	want := filepath.Join(dir, "specs", "001-first-spec", "spec.md")
	if strings.TrimSpace(out) != want {
		t.Fatalf("printed %q want %q", out, want)
	}
	if got := read(t, dir, "specs/001-first-spec/spec.md"); got != "# First spec\n\nSpec 001\n" {
		t.Fatalf("content %q", got)
	}
	if b := tgit(t, dir, "branch", "--show-current"); b != "docs/001-first-spec" {
		t.Fatalf("branch %q", b)
	}
	// Second run: the new spec is untracked, so the tree is not clean. That is
	// the refusal; it must change nothing.
	head := tgit(t, dir, "rev-parse", "HEAD")
	status := tgit(t, dir, "status", "--porcelain")
	code, _, errs = do(t, "spec", "new", "first-spec", dir)
	if code != 2 || !strings.Contains(errs, "not clean") {
		t.Fatalf("second run: %d %s", code, errs)
	}
	if tgit(t, dir, "rev-parse", "HEAD") != head || tgit(t, dir, "status", "--porcelain") != status ||
		tgit(t, dir, "branch", "--show-current") != "docs/001-first-spec" {
		t.Fatal("second run changed the repo")
	}
}

// specNewAt runs spec new and checks the printed path and the new branch.
func specNewAt(t *testing.T, dir, slug, want string) {
	t.Helper()
	code, out, errs := do(t, "spec", "new", slug, dir)
	if code != 0 {
		t.Fatalf("spec new %s: %d %s", slug, code, errs)
	}
	if strings.TrimSpace(out) != filepath.Join(dir, "specs", want, "spec.md") {
		t.Fatalf("spec new %s: out %q, want %s", slug, out, want)
	}
	if b := tgit(t, dir, "branch", "--show-current"); b != "docs/"+want {
		t.Fatalf("spec new %s: branch %q", slug, b)
	}
}

func TestSpecNewSkipsLocalBranchNumbers(t *testing.T) {
	dir := gitRepo(t)
	tgit(t, dir, "branch", "docs/001-dup")
	specNewAt(t, dir, "dup", "002-dup")
}

// Two specs in flight on unmerged branches must not share a number.
func TestSpecNewInFlightBranches(t *testing.T) {
	dir := gitRepo(t)
	specNewAt(t, dir, "alpha", "001-alpha")
	tgit(t, dir, "add", "-A")
	tgit(t, dir, "commit", "-m", "alpha")
	tgit(t, dir, "switch", "dev")
	specNewAt(t, dir, "beta", "002-beta")
}

// withOrigin gives dir a bare origin remote.
func withOrigin(t *testing.T, dir string) {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "origin.git")
	tgit(t, dir, "init", "--bare", "-b", "main", remote)
	tgit(t, dir, "remote", "add", "origin", remote)
}

// pushCommit commits files on a throwaway branch, pushes it to origin as
// dst, and deletes the local branch; the working tree is left as it was.
func pushCommit(t *testing.T, dir, dst string, files ...string) {
	t.Helper()
	tgit(t, dir, "switch", "-q", "-c", "tmp-push", "dev")
	for _, f := range files {
		write(t, dir, f, "x\n")
		tgit(t, dir, "add", "--", f)
	}
	tgit(t, dir, "commit", "-q", "--allow-empty", "-m", "remote")
	tgit(t, dir, "push", "-q", "origin", "tmp-push:"+dst)
	tgit(t, dir, "switch", "-q", "main")
	tgit(t, dir, "branch", "-D", "tmp-push")
}

func TestSpecNewRemoteNumbering(t *testing.T) {
	t.Run("remote docs branch", func(t *testing.T) {
		dir := gitRepo(t)
		withOrigin(t, dir)
		pushCommit(t, dir, "docs/004-remote")
		specNewAt(t, dir, "next", "005-next")
	})
	t.Run("origin dev specs", func(t *testing.T) {
		dir := gitRepo(t)
		withOrigin(t, dir)
		pushCommit(t, dir, "dev", "specs/006-merged/spec.md")
		specNewAt(t, dir, "next", "007-next")
	})
}

func TestADRNewUsesDevAndOriginDev(t *testing.T) {
	dir := gitRepo(t)
	write(t, dir, ".pal-harness/templates/adr.md", "# <NNNN>\n")
	tgit(t, dir, "switch", "-q", "dev")
	write(t, dir, "doc/ADR/0003-on-dev.md", "x\n")
	tgit(t, dir, "add", "--", "doc/ADR/0003-on-dev.md")
	tgit(t, dir, "commit", "-q", "-m", "adr on dev")
	tgit(t, dir, "switch", "-q", "main")
	code, out, errs := do(t, "adr", "new", "a", dir)
	if code != 0 || strings.TrimSpace(out) != filepath.Join(dir, "doc", "ADR", "0004-a.md") {
		t.Fatalf("dev: %d %q %s", code, out, errs)
	}
	withOrigin(t, dir)
	// Untracked doc/ADR/0004-a.md stays in the tree through the switches.
	pushCommit(t, dir, "dev", "doc/ADR/0009-on-origin.md")
	code, out, errs = do(t, "adr", "new", "b", dir)
	if code != 0 || strings.TrimSpace(out) != filepath.Join(dir, "doc", "ADR", "0010-b.md") {
		t.Fatalf("origin/dev: %d %q %s", code, out, errs)
	}
}

func TestSpecNewUsesDevNumbering(t *testing.T) {
	dir := gitRepo(t)
	tgit(t, dir, "switch", "dev")
	write(t, dir, "specs/001-a/spec.md", "a")
	tgit(t, dir, "add", "-A")
	tgit(t, dir, "commit", "-m", "dev spec")
	tgit(t, dir, "switch", "main")
	code, out, errs := do(t, "spec", "new", "b", dir)
	if code != 0 {
		t.Fatalf("%d %s", code, errs)
	}
	if strings.TrimSpace(out) != filepath.Join(dir, "specs", "002-b", "spec.md") {
		t.Fatalf("out %q", out)
	}
	if b := tgit(t, dir, "branch", "--show-current"); b != "docs/002-b" {
		t.Fatalf("branch %q", b)
	}
}

func TestSpecNewRollsBackOnWriteFailure(t *testing.T) {
	dir := gitRepo(t)
	old := writeFile
	writeFile = func(string, string) error { return os.ErrPermission }
	defer func() { writeFile = old }()
	head := tgit(t, dir, "rev-parse", "HEAD")
	code, _, errs := do(t, "spec", "new", "x", dir)
	if code != 2 || !strings.Contains(errs, "rolled back") {
		t.Fatalf("%d %s", code, errs)
	}
	if b := tgit(t, dir, "branch", "--show-current"); b != "main" {
		t.Fatalf("branch %q", b)
	}
	if br := tgit(t, dir, "branch", "--list", "docs/*"); br != "" {
		t.Fatalf("leftover branch %q", br)
	}
	if tgit(t, dir, "rev-parse", "HEAD") != head || tgit(t, dir, "status", "--porcelain") != "" {
		t.Fatal("repo changed")
	}
}

func TestExitTwoOnUnreadableFiles(t *testing.T) {
	// workspace.yaml is a directory: a read error, not a parse error.
	ws := t.TempDir()
	if err := os.Mkdir(filepath.Join(ws, "workspace.yaml"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, c := range []string{"check", "generate"} {
		if code, _, errs := do(t, c, ws); code != 2 || errs == "" {
			t.Errorf("%s workspace: %d %q", c, code, errs)
		}
	}
	// pin.Diff I/O error: a pinned file is a directory.
	dir := fixture(t)
	p := filepath.Join(dir, ".pal-harness", "practices", "workflow.md")
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := do(t, "check", dir); code != 2 || errs == "" {
		t.Errorf("check pin: %d %q", code, errs)
	}
	// generate with an unreadable README.md.
	dir = fixture(t)
	r := filepath.Join(dir, "README.md")
	if err := os.Remove(r); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(r, 0o755); err != nil {
		t.Fatal(err)
	}
	if code, _, errs := do(t, "generate", dir); code != 2 || errs == "" {
		t.Errorf("generate readme: %d %q", code, errs)
	}
}

func TestProjectModeWinsOverWorkspace(t *testing.T) {
	dir := fixture(t)
	write(t, dir, "workspace.yaml", "garbage: [\n")
	if code, out, _ := do(t, "check", dir); code != 0 {
		t.Fatalf("project mode should win: %d %s", code, out)
	}
}

func TestSpecNewPreconditions(t *testing.T) {
	dir := gitRepo(t)
	write(t, dir, "dirty.txt", "x")
	if code, _, errs := do(t, "spec", "new", "a", dir); code != 2 || !strings.Contains(errs, "clean") {
		t.Fatalf("dirty: %d %s", code, errs)
	}
	if err := os.Remove(filepath.Join(dir, "dirty.txt")); err != nil {
		t.Fatal(err)
	}
	if code, _, _ := do(t, "spec", "new", "a", t.TempDir()); code != 2 {
		t.Fatalf("no project.yaml: %d", code)
	}
}

func TestSpecNewNoDev(t *testing.T) {
	dir := gitRepo(t)
	tgit(t, dir, "branch", "-D", "dev")
	if code, _, errs := do(t, "spec", "new", "a", dir); code != 2 || !strings.Contains(errs, "dev") {
		t.Fatalf("no dev: %d %s", code, errs)
	}
}

func TestSpecNewNoTemplateChangesNothing(t *testing.T) {
	dir := gitRepo(t)
	if err := os.Remove(filepath.Join(dir, ".pal-harness", "templates", "spec.md")); err != nil {
		t.Fatal(err)
	}
	tgit(t, dir, "commit", "-am", "rm")
	if code, _, errs := do(t, "spec", "new", "a", dir); code != 2 || !strings.Contains(errs, "template") {
		t.Fatalf("no template: %d %s", code, errs)
	}
	if b := tgit(t, dir, "branch", "--show-current"); b != "main" {
		t.Fatalf("branch changed to %s", b)
	}
}

func TestSpecNewFolderExists(t *testing.T) {
	dir := gitRepo(t)
	// 001 is taken by a non-numbered-looking collision: a file at the target path
	write(t, dir, "specs/001-a", "i am a file")
	tgit(t, dir, "add", "-A")
	tgit(t, dir, "commit", "-m", "x")
	// NextSpec ignores files, so 001 is chosen and the path exists
	if code, _, errs := do(t, "spec", "new", "a", dir); code != 2 || !strings.Contains(errs, "exists") {
		t.Fatalf("%d %s", code, errs)
	}
}

func TestADRNew(t *testing.T) {
	old := now
	now = func() time.Time { return time.Date(2026, 10, 17, 9, 0, 0, 0, time.UTC) }
	defer func() { now = old }()
	dir := t.TempDir()
	write(t, dir, "project.yaml", "version: 1\n")
	write(t, dir, ".pal-harness/templates/adr.md", "# <NNNN>. <Title>\nDate: <Date>\n")
	code, out, errs := do(t, "adr", "new", "use-sqlite", dir)
	if code != 0 {
		t.Fatalf("%d %s", code, errs)
	}
	if strings.TrimSpace(out) != filepath.Join(dir, "doc", "ADR", "0001-use-sqlite.md") {
		t.Fatalf("out %q", out)
	}
	if got := read(t, dir, "doc/ADR/0001-use-sqlite.md"); got != "# 0001. Use sqlite\nDate: 2026-10-17\n" {
		t.Fatalf("content %q", got)
	}
	if code, _, _ := do(t, "adr", "new", "second", dir); code != 0 {
		t.Fatal(code)
	}
	if _, err := os.Stat(filepath.Join(dir, "doc", "ADR", "0002-second.md")); err != nil {
		t.Fatal(err)
	}
}

func TestADRNewConflict(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "project.yaml", "version: 1\n")
	write(t, dir, ".pal-harness/templates/adr.md", "plain\n")
	write(t, dir, ".pal-harness/templates/adr-conflict.md", "C <NNNN> <Title>\n")
	code, _, errs := do(t, "adr", "new", "clash", "--conflict", dir)
	if code != 0 {
		t.Fatalf("%d %s", code, errs)
	}
	if got := read(t, dir, "doc/ADR/0001-clash.md"); got != "C 0001 Clash\n" {
		t.Fatalf("content %q", got)
	}
}

func TestADRNewNoTemplate(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "project.yaml", "version: 1\n")
	code, _, errs := do(t, "adr", "new", "x", dir)
	want := "template .pal-harness/templates/adr.md not found; " + notRoot + "\n"
	if code != 2 || errs != want {
		t.Fatalf("%d %q", code, errs)
	}
}

const notRoot = "run from the SSOT root (the folder with project.yaml)"

func TestCommandsOutsideSSOTRoot(t *testing.T) {
	for _, c := range [][]string{{"adr", "new", "x"}, {"adr", "new", "x", "--conflict"}, {"pin"}, {"spec", "new", "x"}} {
		dir := t.TempDir()
		write(t, dir, ".pal-harness/templates/adr.md", "x\n")
		write(t, dir, ".pal-harness/templates/adr-conflict.md", "x\n")
		code, out, errs := do(t, append(c, dir)...)
		if code != 2 || errs != notRoot+"\n" || out != "" {
			t.Errorf("%v: code %d out=%q err=%q", c, code, out, errs)
		}
		entries, _ := os.ReadDir(dir)
		if len(entries) != 1 {
			t.Errorf("%v: wrote files: %v", c, entries)
		}
	}
}

func TestPinKeepsOverrides(t *testing.T) {
	dir := fixture(t)
	write(t, dir, ".pal-harness/practices/workflow.md", "local patch\n")
	write(t, dir, ".pal-harness/practices/testing.md", "drifted\n")
	proj := strings.Replace(read(t, dir, "project.yaml"), "overrides: []",
		"overrides:\n    - item: practices/workflow.md\n      source: 1.0.0-rc.1+patch\n      reason: r\n      since: 2026-10-17", 1)
	write(t, dir, "project.yaml", proj)
	if code, _, errs := do(t, "pin", dir); code != 0 {
		t.Fatalf("pin: %d %s", code, errs)
	}
	if got := read(t, dir, ".pal-harness/practices/workflow.md"); got != "local patch\n" {
		t.Fatalf("override overwritten: %q", got)
	}
	if got := read(t, dir, ".pal-harness/practices/testing.md"); got == "drifted\n" {
		t.Fatal("non-overridden file not restored")
	}
	if code, out, _ := do(t, "check", dir); code != 0 {
		t.Fatalf("check: %d %s", code, out)
	}
}

func TestPinInvalidProject(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "project.yaml", "version: 3\n")
	code, _, errs := do(t, "pin", dir)
	if code != 2 || !strings.Contains(errs, "project.yaml: version: must be 1\n") {
		t.Fatalf("%d %q", code, errs)
	}
	if _, err := os.Stat(filepath.Join(dir, ".pal-harness")); err == nil {
		t.Fatal("pin wrote files for an invalid project")
	}
}

func stateFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("../../state/testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestCheckWithoutStateFiles(t *testing.T) {
	dir := fixture(t)
	code, out, errs := do(t, "check", dir)
	if code != 0 || out != "" {
		t.Fatalf("code %d out=%q err=%q", code, out, errs)
	}
}

func TestCheckStateFindings(t *testing.T) {
	dir := fixture(t)
	write(t, dir, "doc/features.yaml", strings.Replace(stateFixture(t, "features.yaml"), "label: need", "label: maybe", 1))
	write(t, dir, "specs/004-x/phase-state.yaml", strings.Replace(stateFixture(t, "phase-state.yaml"), "stage: building", "stage: validate", 1))
	code, out, errs := do(t, "check", dir)
	if code != 1 {
		t.Fatalf("code %d out=%q err=%q", code, out, errs)
	}
	for _, want := range []string{
		"doc/features.yaml: F-012: label must be need or might",
		`specs/004-x/phase-state.yaml: stage "validate" is not one of`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %q", want, out)
		}
	}
}

func TestCheckStateValidFiles(t *testing.T) {
	dir := fixture(t)
	write(t, dir, "doc/features.yaml", stateFixture(t, "features.yaml"))
	write(t, dir, "specs/004-x/phase-state.yaml", stateFixture(t, "phase-state.yaml"))
	code, out, errs := do(t, "check", dir)
	if code != 0 {
		t.Fatalf("code %d out=%q err=%q", code, out, errs)
	}
}

func TestCheckStateUnreadable(t *testing.T) {
	dir := fixture(t)
	write(t, dir, "doc/features.yaml/x", "dir where a file should be\n")
	code, _, errs := do(t, "check", dir)
	if code != 2 || !strings.Contains(errs, "doc/features.yaml") {
		t.Fatalf("code %d err=%q", code, errs)
	}
}

func TestCheckReportsLoaderDrift(t *testing.T) {
	ls := pin.Loaders()
	if len(ls) == 0 {
		t.Fatal("no embedded loaders")
	}
	dir := fixture(t)
	if code, out, _ := do(t, "check", dir); code != 0 {
		t.Fatalf("clean fixture: %d %s", code, out)
	}
	if err := os.Remove(filepath.Join(dir, filepath.FromSlash(ls[0].Path))); err != nil {
		t.Fatal(err)
	}
	code, out, _ := do(t, "check", dir)
	if code != 1 || !strings.Contains(out, "missing "+ls[0].Path) {
		t.Fatalf("code %d out %q", code, out)
	}
}

func TestRenderMissingDiagram(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "spec.md", "# T\n\n[Flow](diagrams/none.mmd)\n")
	code, _, stderr := do(t, "render", filepath.Join(dir, "spec.md"))
	if code != 1 || stderr != "render: diagrams/none.mmd: file not found\n" {
		t.Errorf("code %d stderr %q", code, stderr)
	}
	if exists(filepath.Join(dir, "spec.html")) {
		t.Error("spec.html written")
	}
}

func TestRenderWritesHTML(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "spec.md", "# Hello\n\n[Flow](diagrams/flow.mmd)\n")
	write(t, dir, "diagrams/flow.mmd", "graph TD; A-->B")
	code, _, stderr := do(t, "render", filepath.Join(dir, "spec.md"))
	if code != 0 {
		t.Fatalf("code %d stderr %q", code, stderr)
	}
	html := read(t, dir, "spec.html")
	if !strings.Contains(html, "<title>Hello</title>") || !strings.Contains(html, "A--&gt;B") {
		t.Errorf("html = %s", html)
	}
}

func TestRenderArgCount(t *testing.T) {
	if code, _, _ := do(t, "render"); code != 2 {
		t.Errorf("no args: code %d", code)
	}
	if code, _, _ := do(t, "render", "a.md", "b.md"); code != 2 {
		t.Errorf("two args: code %d", code)
	}
}

func setLegacy(t *testing.T, dir, before string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "project.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	y := strings.Replace(string(b), "overrides:", "legacy_specs_before: \""+before+"\"\n  overrides:", 1)
	write(t, dir, "project.yaml", y)
}

func TestCheckLegacySpecs(t *testing.T) {
	old := "stage: old\n"
	t.Run("skipped before the marker", func(t *testing.T) {
		dir := fixture(t)
		setLegacy(t, dir, "097")
		write(t, dir, "specs/090-x/phase-state.yaml", old)
		if code, out, errs := do(t, "check", dir); code != 0 {
			t.Fatalf("code %d out=%q err=%q", code, out, errs)
		}
	})
	t.Run("reported without the marker", func(t *testing.T) {
		dir := fixture(t)
		write(t, dir, "specs/090-x/phase-state.yaml", old)
		code, out, _ := do(t, "check", dir)
		if code != 1 || !strings.Contains(out, "specs/090-x/phase-state.yaml") {
			t.Fatalf("code %d out=%q", code, out)
		}
	})
	t.Run("reported at the marker", func(t *testing.T) {
		dir := fixture(t)
		setLegacy(t, dir, "097")
		write(t, dir, "specs/097-y/phase-state.yaml", old)
		code, out, _ := do(t, "check", dir)
		if code != 1 || !strings.Contains(out, "specs/097-y/phase-state.yaml") {
			t.Fatalf("code %d out=%q", code, out)
		}
	})
}

func TestCheckAutoPromoteWarns(t *testing.T) {
	dir := fixture(t)
	proj := read(t, dir, "project.yaml")
	write(t, dir, "project.yaml", strings.Replace(proj, "overrides: []", "overrides: []\n  flow_overrides: [auto-promote]", 1))
	code, out, errs := do(t, "check", dir)
	if code != 0 {
		t.Fatalf("code %d out=%q err=%q", code, out, errs)
	}
	if !strings.Contains(out, "warning: auto-promote is on: sign-off and ship PRs merge without the owner") {
		t.Errorf("missing warning in %q", out)
	}
}

func TestCheckShippedWithoutOverride(t *testing.T) {
	dir := fixture(t)
	feat := strings.Replace(stateFixture(t, "features.yaml"), "status: todo", "status: shipped", 1)
	write(t, dir, "doc/features.yaml", feat)
	code, out, _ := do(t, "check", dir)
	if code != 1 || !strings.Contains(out, "doc/features.yaml: ") || !strings.Contains(out, "shipped needs the no-owner-test flow override") {
		t.Fatalf("code %d out=%q", code, out)
	}
	proj := read(t, dir, "project.yaml")
	write(t, dir, "project.yaml", strings.Replace(proj, "overrides: []", "overrides: []\n  flow_overrides: [no-owner-test]", 1))
	code, out, _ = do(t, "check", dir)
	if code != 0 {
		t.Fatalf("with override: code %d out=%q", code, out)
	}
}
