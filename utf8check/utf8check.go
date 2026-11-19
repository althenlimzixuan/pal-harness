// Package utf8check finds text files tracked by git that are not valid UTF-8,
// such as a Windows-1252 quote pasted into a Markdown file.
package utf8check

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// ErrNotRepo is returned when the folder is not inside a git work tree.
var ErrNotRepo = errors.New("not a git repository")

// binarySniff is how many leading bytes are searched for a NUL, as git does
// to tell a binary file from a text file.
const binarySniff = 8000

// Finding is one tracked text file that is not UTF-8, with the first bad line.
type Finding struct {
	Path string // slash-separated, relative to the repo folder, as git lists it
	Line int
}

func (f Finding) String() string { return fmt.Sprintf("%s:%d: not UTF-8", f.Path, f.Line) }

// Check reads every file git tracks under dir and returns the text files that
// are not valid UTF-8, in git's order. Binary files (a NUL in the first 8000
// bytes), symlinks, submodules and tracked files missing from disk are skipped.
func Check(dir string) ([]Finding, error) {
	cmd := exec.Command("git", "-C", dir, "ls-files", "-z")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(stderr.String(), "not a git repository") {
			return nil, ErrNotRepo
		}
		return nil, fmt.Errorf("git ls-files: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	var found []Finding
	for _, p := range strings.Split(string(out), "\x00") {
		if p == "" {
			continue
		}
		full := filepath.Join(dir, filepath.FromSlash(p))
		info, err := os.Lstat(full)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		b, err := os.ReadFile(full)
		if err != nil {
			return nil, err
		}
		if line, bad := firstBadLine(b); bad {
			found = append(found, Finding{Path: p, Line: line})
		}
	}
	return found, nil
}

// firstBadLine reports the 1-based line of the first invalid UTF-8 byte; a
// binary file is never bad.
func firstBadLine(b []byte) (int, bool) {
	if bytes.IndexByte(b[:min(len(b), binarySniff)], 0) >= 0 || utf8.Valid(b) {
		return 0, false
	}
	line := 1
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r == utf8.RuneError && size == 1 {
			return line, true
		}
		if r == '\n' {
			line++
		}
		i += size
	}
	return line, true
}
