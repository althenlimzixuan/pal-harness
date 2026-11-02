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

// InWorktreeSet reports whether path lies inside a .pal-worktrees folder.
func InWorktreeSet(path string) bool {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return slices.Contains(strings.FieldsFunc(path, func(r rune) bool { return r == '/' || r == '\\' }), SetDir)
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

	if err := os.WriteFile(logPath, []byte(prev+entry), 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(statePath, newState, 0o644); err != nil {
		return "", err
	}
	return stamp, nil
}
