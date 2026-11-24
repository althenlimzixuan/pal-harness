package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/althenlimzixuan/pal-harness/cpu"
)

// cpuRead and cpuSleep are the OS reader and the sleeper; tests replace them.
var (
	cpuRead  = cpu.Read
	cpuSleep = time.Sleep
)

// runCPU checks the machine is quiet before a timing run: 0 quiet, 1 still busy, 2 usage or unreadable.
func runCPU(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("cpu", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	maxArg := fs.String("max", "25%", "")
	over := fs.Duration("for", 10*time.Second, "")
	wait := fs.Duration("wait", 0, "")
	every := fs.Duration("every", 30*time.Second, "")
	if err := fs.Parse(args); err != nil {
		return usageErr(stderr, "cpu: "+err.Error())
	}
	if fs.NArg() != 0 {
		return usageErr(stderr, "cpu takes only --max, --for, --wait and --every")
	}
	if *over <= 0 || *wait < 0 || *every <= 0 {
		return usageErr(stderr, "cpu: --for and --every must be above 0, and --wait 0 or more")
	}
	maxBusy, err := cpu.ParsePercent(*maxArg)
	if err != nil {
		return usageErr(stderr, "cpu: "+err.Error())
	}
	_, ok, err := cpu.Check(cpu.Options{Max: maxBusy, For: *over, Wait: *wait, Every: *every, Read: cpuRead, Now: now, Sleep: cpuSleep}, stdout)
	if err != nil {
		fmt.Fprintf(stderr, "cannot read CPU use: %v\n", err)
		return 2
	}
	if !ok {
		return 1
	}
	return 0
}
