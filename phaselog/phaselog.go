// Package phaselog appends clock-stamped entries to a phase's phase-log.md and
// stamps last_progress_at in its phase-state.yaml, so no timestamp is typed by
// hand.
package phaselog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// SetDir is the folder that holds every worktree set.
const SetDir = ".pal-worktrees"

var stampLine = regexp.MustCompile(`(?m)^last_progress_at:.*$`)

// InWorktreeSet reports whether path lies inside a set under a .pal-worktrees
// folder (not the folder itself). The name is matched ignoring case, as
// Windows paths are.
func InWorktreeSet(path string) bool {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	parts := strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '\\' })
	i := slices.IndexFunc(parts, func(p string) bool { return strings.EqualFold(p, SetDir) })
	return i >= 0 && i < len(parts)-1
}

// replace writes data to path through a temp file in the same folder and a
// rename, so path holds either its old or its new content.
func replace(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr == nil {
		werr = cerr
	}
	if werr == nil {
		if info, err := os.Stat(path); err == nil && info.Mode().Perm()&0o200 == 0 {
			werr = fmt.Errorf("%s is read-only", path)
		}
	}
	if werr == nil {
		werr = os.Rename(name, path)
	}
	if werr != nil {
		_ = os.Remove(name)
	}
	return werr
}

// Append adds "## YYYY-MM-DD HH:MM — <headline>" (and the body, after a blank
// line) to <dir>/phase-log.md and sets last_progress_at in
// <dir>/phase-state.yaml, both from t. dir is a phase folder (specs/NNN-slug)
// inside a worktree set. It returns the entry's stamp. On any error it writes
// nothing.
func Append(dir, headline, body string, t time.Time) (string, error) {
	headline = strings.TrimSpace(headline)
	switch {
	case headline == "":
		return "", errors.New("the headline is empty")
	case strings.ContainsAny(headline, "\r\n"):
		return "", errors.New("the headline must be one line; put detail in the body file")
	}
	if !InWorktreeSet(dir) {
		return "", fmt.Errorf("%s is outside a worktree set (%s/): phase logs are written only in the set", dir, SetDir)
	}
	statePath := filepath.Join(dir, "phase-state.yaml")
	state, err := os.ReadFile(statePath)
	if err != nil {
		return "", fmt.Errorf("no phase-state.yaml in %s: give the phase folder (specs/NNN-slug)", dir)
	}
	if !stampLine.Match(state) {
		return "", fmt.Errorf("%s has no last_progress_at line", statePath)
	}
	logPath := filepath.Join(dir, "phase-log.md")
	old, err := os.ReadFile(logPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	stamp := t.Format("2006-01-02 15:04")
	entry := "## " + stamp + " — " + headline + "\n"
	if b := strings.TrimSpace(body); b != "" {
		entry += "\n" + b + "\n"
	}
	prev := string(old)
	if prev != "" {
		prev = strings.TrimRight(prev, "\n") + "\n\n"
	}
	newState := stampLine.ReplaceAll(state, []byte("last_progress_at: "+t.Format("2006-01-02T15:04:05-07:00")))

	// State first, then the log; if the log fails, put the old state back.
	if err := replace(statePath, newState); err != nil {
		return "", err
	}
	if err := replace(logPath, []byte(prev+entry)); err != nil {
		if rerr := replace(statePath, state); rerr != nil {
			return "", fmt.Errorf("%w (and restoring %s failed: %v)", err, statePath, rerr)
		}
		return "", err
	}
	return stamp, nil
}
