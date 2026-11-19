package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/althenlimzixuan/pal-harness/utf8check"
)

// runUTF8 lists the tracked text files that are not UTF-8: 0 clean, 1 found, 2 otherwise.
func runUTF8(dir string, stdout, stderr io.Writer) int {
	found, err := utf8check.Check(dir)
	if errors.Is(err, utf8check.ErrNotRepo) {
		return fail(stderr, "%s: not a git repository", dir)
	}
	if err != nil {
		return fail(stderr, "%v", err)
	}
	for _, f := range found {
		fmt.Fprintln(stdout, f)
	}
	if len(found) > 0 {
		fmt.Fprintf(stdout, "%d file(s) are not UTF-8: save them as UTF-8\n", len(found))
		return 1
	}
	fmt.Fprintln(stdout, "every tracked text file is UTF-8")
	return 0
}
