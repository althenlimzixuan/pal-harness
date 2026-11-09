package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/althenlimzixuan/pal-harness/mem"
)

// memRead and memSleep are the OS reader and the sleeper; tests replace them.
var (
	memRead  = mem.Available
	memSleep = time.Sleep
)

// runMem checks free memory before a heavy run: 0 enough, 1 still short, 2 usage or unreadable.
func runMem(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("mem", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	minArg := fs.String("min", "2GiB", "")
	wait := fs.Duration("wait", 0, "")
	every := fs.Duration("every", 30*time.Second, "")
	if err := fs.Parse(args); err != nil {
		return usageErr(stderr, "mem: "+err.Error())
	}
	if fs.NArg() != 0 {
		return usageErr(stderr, "mem takes only --min, --wait and --every")
	}
	if *wait < 0 || *every <= 0 {
		return usageErr(stderr, "mem: --wait must be 0 or more and --every above 0")
	}
	minBytes, err := mem.ParseSize(*minArg)
	if err != nil {
		return usageErr(stderr, "mem: "+err.Error())
	}
	ok, err := mem.Check(mem.Options{Min: minBytes, Wait: *wait, Every: *every, Read: memRead, Now: now, Sleep: memSleep}, stdout)
	if err != nil {
		fmt.Fprintf(stderr, "cannot read free memory: %v\n", err)
		return 2
	}
	if !ok {
		return 1
	}
	return 0
}
