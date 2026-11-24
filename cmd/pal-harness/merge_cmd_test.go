package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// mergeRepo is a repo on branch feature/x (from dev) with a bare origin that
// holds main and dev as they are now.
func mergeRepo(t *testing.T) string {
	t.Helper()
	dir := gitRepo(t)
	withOrigin(t, dir)
	tgit(t, dir, "push", "-q", "origin", "main", "dev")
	tgit(t, dir, "switch", "-q", "-c", "feature/x", "dev")
	return dir
}

// commitFile commits one file on the current branch.
func commitFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	write(t, dir, rel, content)
	tgit(t, dir, "add", "--", rel)
	tgit(t, dir, "commit", "-q", "-m", "change "+rel)
}

// pushOnto moves origin's branch on by one commit that writes rel, leaving
// the caller's branch, tree and remote-tracking refs as they were.
func pushOnto(t *testing.T, dir, branch, rel, content string) {
	t.Helper()
	other := filepath.Join(t.TempDir(), "other")
	tgit(t, dir, "clone", "-q", "--branch", branch, tgit(t, dir, "remote", "get-url", "origin"), other)
	commitFile(t, other, rel, content)
	tgit(t, other, "push", "-q", "origin", branch)
}

// repoState is what merge-check must never change.
func repoState(t *testing.T, dir string) string {
	t.Helper()
	return strings.Join([]string{
		tgit(t, dir, "rev-parse", "HEAD"),
		tgit(t, dir, "branch", "--show-current"),
		tgit(t, dir, "status", "--porcelain"),
		tgit(t, dir, "worktree", "list", "--porcelain"),
	}, "\n--\n")
}

func TestMergeCheckClean(t *testing.T) {
	dir := mergeRepo(t)
	commitFile(t, dir, "doc/decisions.md", "106\n")
	pushOnto(t, dir, "dev", "doc/later.md", "later\n")
	before := repoState(t, dir)
	code, out, errs := do(t, "merge-check", dir)
	if code != 0 || out != "merge-check: clean against origin/dev\n" {
		t.Fatalf("code %d out %q err %s", code, out, errs)
	}
	if after := repoState(t, dir); after != before {
		t.Fatalf("repo changed:\n%s\nwas:\n%s", after, before)
	}
}

func TestMergeCheckConflict(t *testing.T) {
	dir := mergeRepo(t)
	commitFile(t, dir, "doc/decisions.md", "106\n")
	commitFile(t, dir, "doc/other.md", "mine\n")
	pushOnto(t, dir, "dev", "doc/decisions.md", "105\n")
	before := repoState(t, dir)
	code, out, errs := do(t, "merge-check", dir)
	if code != 1 || out != "merge-check: conflicts with origin/dev:\n  doc/decisions.md\n" {
		t.Fatalf("code %d out %q err %s", code, out, errs)
	}
	if after := repoState(t, dir); after != before {
		t.Fatalf("repo changed:\n%s\nwas:\n%s", after, before)
	}
}

func TestMergeCheckBase(t *testing.T) {
	dir := mergeRepo(t)
	commitFile(t, dir, "doc/decisions.md", "106\n")
	pushOnto(t, dir, "main", "doc/decisions.md", "hotfix\n")
	if code, out, errs := do(t, "merge-check", "--base", "origin/main", dir); code != 1 ||
		!strings.Contains(out, "conflicts with origin/main") {
		t.Fatalf("origin/main: code %d out %q err %s", code, out, errs)
	}
	if code, out, errs := do(t, "merge-check", dir); code != 0 {
		t.Fatalf("origin/dev: code %d out %q err %s", code, out, errs)
	}
}

func TestMergeCheckErrors(t *testing.T) {
	dir := mergeRepo(t)
	tgit(t, dir, "remote", "set-url", "origin", filepath.Join(t.TempDir(), "gone.git"))
	before := repoState(t, dir)
	if code, _, errs := do(t, "merge-check", dir); code != 2 || !strings.Contains(errs, "fetch") {
		t.Fatalf("failed fetch: code %d err %s", code, errs)
	}
	if repoState(t, dir) != before {
		t.Fatal("a failed fetch changed the repo")
	}
	for name, args := range map[string][]string{
		"not a repo":  {"merge-check", t.TempDir()},
		"bad base":    {"merge-check", "--base", "dev", dir},
		"two folders": {"merge-check", dir, dir},
	} {
		if code, _, errs := do(t, args...); code != 2 || errs == "" {
			t.Errorf("%s: code %d err %q", name, code, errs)
		}
	}
}
