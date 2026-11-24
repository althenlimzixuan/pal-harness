package state

import (
	"fmt"
	"regexp"
	"strings"
)

// Results are the values of an expected line's result in retro.md.
var Results = []string{"met", "partly", "missed"}

// Expected is one expectation of a phase and what it got, from the front
// matter of specs/NNN-slug/retro.md (pal-harness 1.12.0).
type Expected struct {
	Line     string `yaml:"line"`
	Source   string `yaml:"source"`
	Result   string `yaml:"result"`
	Evidence string `yaml:"evidence"`
	Why      string `yaml:"why"`
}

type retroFront struct {
	Expected []Expected `yaml:"expected"`
}

var sourceRe = regexp.MustCompile(`^(summary \d+|F-\d{3,} (step|expect) \d+)$`)

// ParseRetro reads retro.md's front matter strictly and validates its expected
// list. A retro without front matter (every retro before 1.12.0) returns nil, nil.
func ParseRetro(b []byte) ([]Expected, []string) {
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return nil, nil
	}
	body, _, ok := strings.Cut(s[len("---\n"):], "\n---\n")
	if !ok {
		if strings.HasSuffix(s, "\n---") {
			body, ok = strings.TrimSuffix(s[len("---\n"):], "\n---"), true
		}
	}
	if !ok {
		return nil, []string{"front matter is not closed (a line with --- after it)"}
	}
	var f retroFront
	if msg := decode([]byte(body), &f); msg != "" {
		return nil, []string{msg}
	}
	if len(f.Expected) == 0 {
		return nil, []string{"expected is empty: list each expectation, or drop the front matter"}
	}
	var out []string
	for i, e := range f.Expected {
		n := i + 1
		add := func(format string, a ...any) {
			out = append(out, fmt.Sprintf("expected %d: ", n)+fmt.Sprintf(format, a...))
		}
		if strings.TrimSpace(e.Line) == "" {
			add("line is required")
		}
		switch {
		case strings.TrimSpace(e.Source) == "":
			add("source is required")
		case !sourceRe.MatchString(e.Source):
			add("source %q is not summary N, F-NNN step N or F-NNN expect N", e.Source)
		}
		if !oneOf(e.Result, Results) {
			add("%s", notOneOf("result", e.Result, Results))
		}
		if strings.TrimSpace(e.Evidence) == "" {
			add("evidence is required")
		}
		if (e.Result == "partly" || e.Result == "missed") && strings.TrimSpace(e.Why) == "" {
			add("%s needs why", e.Result)
		}
	}
	return f.Expected, out
}
