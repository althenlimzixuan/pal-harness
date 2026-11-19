package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/althenlimzixuan/pal-harness/fingerprint"
)

// stdin is the spec source for "-"; tests replace it.
var stdin io.Reader = os.Stdin

// paths is a repeatable flag.
type paths []string

func (p *paths) String() string     { return strings.Join(*p, " ") }
func (p *paths) Set(v string) error { *p = append(*p, v); return nil }

// runFingerprint prints "<job>=<fingerprint>" for each job in the spec: 0 printed, 2 otherwise.
func runFingerprint(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("fingerprint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	rev := fs.String("rev", "HEAD", "")
	dir := fs.String("dir", ".", "")
	var with paths
	fs.Var(&with, "with", "")
	if err := fs.Parse(args); err != nil {
		return usageErr(stderr, "fingerprint: "+err.Error())
	}
	if fs.NArg() != 1 {
		return usageErr(stderr, "fingerprint takes one spec file (or - for stdin)")
	}
	var shared []string
	for _, w := range with {
		p, err := fingerprint.CheckPath(w)
		if err != nil {
			return usageErr(stderr, "fingerprint --with: "+err.Error())
		}
		shared = append(shared, p)
	}
	src := stdin
	if name := fs.Arg(0); name != "-" {
		f, err := os.Open(name)
		if err != nil {
			return fail(stderr, "cannot read the spec: %v", err)
		}
		defer f.Close()
		src = f
	}
	jobs, err := fingerprint.ParseSpec(src)
	if err != nil {
		return fail(stderr, "spec: %v", err)
	}
	lookup, err := fingerprint.NewGit(*dir, *rev)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	sums, err := fingerprint.Compute(jobs, shared, lookup)
	if err != nil {
		return fail(stderr, "%v", err)
	}
	for _, s := range sums {
		fmt.Fprintf(stdout, "%s=%s\n", s.Name, s.Sum)
	}
	return 0
}
