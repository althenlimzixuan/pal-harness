package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// minorTags are the tags a deferred minor may carry in sdd's ledger, in the
// order they print. "parked" and "untagged" are groups the command adds:
// sdd's own parked lines, and minors with no tag or an unknown one.
var (
	minorTags   = []string{"carry", "final", "later", "issue", "fixed"}
	minorGroups = []string{"carry", "final", "later", "issue", "parked", "untagged", "fixed"}
)

var (
	// Task 3: minor (deferred) [carry]: text   (the tag is optional)
	minorLine = regexp.MustCompile(`^(?:[-*]\s+)?Task (\d+): (minor \(deferred\)(?: \[([^\]]*)\])?: (.+))$`)
	// Task 4: parked — finding — Ruling: why
	parkedLine = regexp.MustCompile(`^(?:[-*]\s+)?Task (\d+): parked — (.+)$`)
)

type minor struct {
	task  int
	group string
	text  string
}

// parseMinors reads the deferred minors and parked findings from a ledger.
func parseMinors(text string) []minor {
	var out []minor
	for l := range strings.SplitSeq(text, "\n") {
		l = strings.TrimSpace(l)
		if m := minorLine.FindStringSubmatch(l); m != nil {
			n, _ := strconv.Atoi(m[1])
			if tag := m[3]; slices.Contains(minorTags, tag) {
				out = append(out, minor{n, tag, m[4]})
			} else {
				// An untagged line prints whole, so the missing or unknown tag shows.
				out = append(out, minor{n, "untagged", m[2]})
			}
		} else if m := parkedLine.FindStringSubmatch(l); m != nil {
			n, _ := strconv.Atoi(m[1])
			out = append(out, minor{n, "parked", m[2]})
		}
	}
	return out
}

// runMinors lists a ledger's deferred minors by tag: 0 all tagged, 1 one is
// untagged or has an unknown tag, 2 usage or unreadable file.
func runMinors(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("minors", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	tagArg := fs.String("tag", "", "")
	taskArg := fs.String("task", "", "")
	if err := fs.Parse(args); err != nil {
		return usageErr(stderr, "minors: "+err.Error())
	}
	if fs.NArg() != 1 {
		return usageErr(stderr, "minors takes one ledger file")
	}
	if *tagArg != "" && !slices.Contains(minorGroups, *tagArg) {
		return usageErr(stderr, "minors: --tag must be one of "+strings.Join(minorGroups, ", "))
	}
	task := 0
	if *taskArg != "" {
		n, err := strconv.Atoi(*taskArg)
		if err != nil || n < 1 {
			return usageErr(stderr, "minors: --task must be a task number")
		}
		task = n
	}
	b, err := os.ReadFile(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "cannot read the ledger: %v\n", err)
		return 2
	}
	groups := map[string][]minor{}
	total := 0
	for _, m := range parseMinors(string(b)) {
		if (task != 0 && m.task != task) || (*tagArg != "" && m.group != *tagArg) {
			continue
		}
		groups[m.group] = append(groups[m.group], m)
		total++
	}
	if total == 0 {
		fmt.Fprintln(stdout, "no deferred minors")
		return 0
	}
	for _, g := range minorGroups {
		ms := groups[g]
		if len(ms) == 0 {
			continue
		}
		fmt.Fprintf(stdout, "%s (%d)\n", g, len(ms))
		if g == "fixed" {
			continue
		}
		for _, m := range ms {
			fmt.Fprintf(stdout, "  Task %d: %s\n", m.task, m.text)
		}
	}
	if n := len(groups["untagged"]); n > 0 {
		verb := "minors have"
		if n == 1 {
			verb = "minor has"
		}
		fmt.Fprintf(stderr, "%d %s no known tag (%s)\n", n, verb, strings.Join(minorTags, ", "))
		return 1
	}
	return 0
}
