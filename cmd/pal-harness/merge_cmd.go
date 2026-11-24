package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// runMergeCheck tries the merge of HEAD with a remote branch in a throwaway
// worktree, so a conflict shows before CI is waited on: 0 clean, 1 conflicts,
// 2 usage, not a repo, a failed fetch or a merge that could not run. The
// caller's tree, index and branch are never touched.
func runMergeCheck(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("merge-check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	base := fs.String("base", "origin/dev", "")
	if err := fs.Parse(args); err != nil {
		return usageErr(stderr, "merge-check: "+err.Error())
	}
	if fs.NArg() > 1 {
		return usageErr(stderr, "merge-check takes at most one directory")
	}
	remote, branch, ok := strings.Cut(*base, "/")
	if !ok || remote == "" || branch == "" {
		return usageErr(stderr, "merge-check: --base must be <remote>/<branch>, such as origin/dev")
	}
	dir := dirArg(fs.Args())
	if out, err := git(dir, "rev-parse", "--verify", "-q", "HEAD"); err != nil {
		return fail(stderr, "%s is not a git repo with a commit: %s", dir, out)
	}
	remotes, _ := git(dir, "remote")
	if !slices.Contains(strings.Fields(remotes), remote) {
		return fail(stderr, "no remote named %s", remote)
	}
	if out, err := git(dir, "fetch", "-q", remote, branch); err != nil {
		return fail(stderr, "git fetch %s %s failed: %v: %s", remote, branch, err, out)
	}
	ref := "refs/remotes/" + *base
	if _, err := git(dir, "rev-parse", "--verify", "-q", ref); err != nil {
		return fail(stderr, "%s does not exist after the fetch", *base)
	}

	tmp, err := os.MkdirTemp("", "pal-merge-check-")
	if err != nil {
		return fail(stderr, "cannot make a temp folder: %v", err)
	}
	wt := filepath.Join(tmp, "wt")
	defer func() {
		// Only the worktree this run created is removed, then its temp folder.
		git(dir, "worktree", "remove", "--force", wt)
		os.RemoveAll(tmp)
		git(dir, "worktree", "prune")
	}()
	if out, err := git(dir, "worktree", "add", "-q", "--detach", wt, "HEAD"); err != nil {
		return fail(stderr, "git worktree add failed: %v: %s", err, out)
	}
	out, err := git(wt, "-c", "user.name=pal-harness", "-c", "user.email=pal-harness@localhost",
		"merge", "-q", "--no-commit", "--no-ff", ref)
	if err == nil {
		fmt.Fprintf(stdout, "merge-check: clean against %s\n", *base)
		return 0
	}
	files, _ := git(wt, "diff", "--name-only", "--diff-filter=U")
	if files == "" {
		return fail(stderr, "git merge %s could not run: %v: %s", *base, err, out)
	}
	fmt.Fprintf(stdout, "merge-check: conflicts with %s:\n", *base)
	for f := range strings.FieldsSeq(files) {
		fmt.Fprintf(stdout, "  %s\n", f)
	}
	return 1
}
