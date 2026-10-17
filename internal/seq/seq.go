// Package seq parses spec and ADR numbers out of names: directory entries,
// git tree paths and branch names. It is internal so the numbering rules stay
// one implementation without becoming public API.
package seq

import (
	"path"
	"regexp"
	"strconv"
)

var (
	specRe = regexp.MustCompile(`^(\d{3})-`)
	adrRe  = regexp.MustCompile(`^(\d{4})-.*\.md$`)
)

func match(re *regexp.Regexp, name string) (int, bool) {
	m := re.FindStringSubmatch(name)
	if m == nil {
		return 0, false
	}
	n, _ := strconv.Atoi(m[1])
	return n, true
}

// Spec returns NNN when name is shaped like "NNN-...".
func Spec(name string) (int, bool) { return match(specRe, name) }

// ADR returns NNNN when name is shaped like "NNNN-....md".
func ADR(name string) (int, bool) { return match(adrRe, name) }

func highest(names []string, num func(string) (int, bool)) int {
	m := 0
	for _, n := range names {
		if v, ok := num(path.Base(n)); ok && v > m {
			m = v
		}
	}
	return m
}

// MaxSpec is the highest NNN among the last elements of names
// ("specs/NNN-x", "docs/NNN-x", "origin/docs/NNN-x"); 0 when none match.
func MaxSpec(names []string) int { return highest(names, Spec) }

// MaxADR is the highest NNNN among the last elements of names
// ("doc/ADR/NNNN-x.md"); 0 when none match.
func MaxADR(names []string) int { return highest(names, ADR) }
