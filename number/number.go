// Package number allocates the next spec and ADR numbers and validates slugs.
package number

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var (
	specRe = regexp.MustCompile(`^(\d{3})-`)
	adrRe  = regexp.MustCompile(`^(\d{4})-.*\.md$`)
	slugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// next returns 1 + the highest number captured by re among entries of dir
// (directories only when wantDir, files only otherwise); 1 when dir is absent.
func next(dir string, re *regexp.Regexp, wantDir bool) (int, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}
	max := 0
	for _, e := range entries {
		if e.IsDir() != wantDir {
			continue
		}
		m := re.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		if n, _ := strconv.Atoi(m[1]); n > max {
			max = n
		}
	}
	return max + 1, nil
}

// NextSpec is 1 + the highest NNN among root/specs/NNN-* directories.
func NextSpec(root string) (int, error) {
	return next(filepath.Join(root, "specs"), specRe, true)
}

// NextADR is 1 + the highest NNNN among root/doc/ADR/NNNN-*.md files.
func NextADR(root string) (int, error) {
	return next(filepath.Join(root, "doc", "ADR"), adrRe, false)
}

// ValidSlug reports whether s is a lowercase-dash slug of at most 50 characters.
func ValidSlug(s string) bool { return len(s) <= 50 && slugRe.MatchString(s) }
