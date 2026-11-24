// Package cpu checks how busy the machine's processors are before a timing
// run (the paired speed check), so a session can wait for a quiet machine
// instead of comparing timings taken under someone else's load.
package cpu

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Times is a cumulative reading of the whole machine's processor time, in the
// platform's own ticks: Idle is the idle part of Total.
type Times struct {
	Idle, Total uint64
}

// Busy returns the share of time between a and b that was not idle, as a
// percent from 0 to 100. No time passed reads as 0.
func Busy(a, b Times) float64 {
	if b.Total <= a.Total {
		return 0
	}
	total := b.Total - a.Total
	idle := uint64(0)
	if b.Idle > a.Idle {
		idle = min(b.Idle-a.Idle, total)
	}
	return float64(total-idle) * 100 / float64(total)
}

// fromSystemTimes turns Windows' GetSystemTimes values into Times. Windows'
// kernel time already includes idle time, so the total is kernel plus user.
func fromSystemTimes(idle, kernel, user uint64) Times {
	return Times{Idle: idle, Total: kernel + user}
}

// parseProcStat reads the first "cpu" line of a Linux /proc/stat: idle is idle
// plus iowait, and the total is the first eight fields (user, nice, system,
// idle, iowait, irq, softirq, steal); guest time is already counted in user.
func parseProcStat(r io.Reader) (Times, error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 0 || f[0] != "cpu" {
			continue
		}
		if len(f) < 5 {
			return Times{}, fmt.Errorf("short cpu line %q", sc.Text())
		}
		var v [8]uint64
		for i := 1; i < len(f) && i <= 8; i++ {
			n, err := strconv.ParseUint(f[i], 10, 64)
			if err != nil {
				return Times{}, fmt.Errorf("bad cpu line %q", sc.Text())
			}
			v[i-1] = n
		}
		var t Times
		for _, n := range v {
			t.Total += n
		}
		t.Idle = v[3] + v[4]
		return t, nil
	}
	return Times{}, errors.New("no cpu line in /proc/stat")
}

// ParsePercent reads a share such as "25%" or "25", from 0 to 100.
func ParsePercent(s string) (float64, error) {
	num := strings.TrimSuffix(strings.TrimSpace(s), "%")
	f, err := strconv.ParseFloat(num, 64)
	if err != nil || num == "" || f < 0 || f > 100 {
		return 0, fmt.Errorf("bad percent %q: want a number from 0 to 100, such as 25%%", s)
	}
	return f, nil
}

// Options configures Check. Read, Now and Sleep are injected so tests need no
// real clock or OS; nil Now and Sleep use the real clock.
type Options struct {
	Max   float64       // the busiest share, in percent, that still counts as quiet
	For   time.Duration // how long one measurement lasts; 0 means 10 s
	Wait  time.Duration // how long to keep re-checking while busy; 0 checks once
	Every time.Duration // the pause between checks; 0 means 30 s
	Read  func() (Times, error)
	Now   func() time.Time
	Sleep func(time.Duration)
}

// measure reads the counters, waits o.For and reads them again.
func measure(o Options) (float64, error) {
	a, err := o.Read()
	if err != nil {
		return 0, err
	}
	o.Sleep(o.For)
	b, err := o.Read()
	if err != nil {
		return 0, err
	}
	return Busy(a, b), nil
}

// Check measures how busy the machine is over o.For and reports whether it is
// at most o.Max, re-measuring for up to o.Wait while busy. It writes one line
// per verdict to out and returns the last share measured.
func Check(o Options, out io.Writer) (float64, bool, error) {
	if o.For <= 0 {
		o.For = 10 * time.Second
	}
	if o.Every <= 0 {
		o.Every = 30 * time.Second
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Sleep == nil {
		o.Sleep = time.Sleep
	}
	line := func(busy float64, verdict string) {
		fmt.Fprintf(out, "cpu busy: %.0f%% over %v (maximum %g%%): %s\n", busy, o.For, o.Max, verdict)
	}
	start := o.Now()
	busy, err := measure(o)
	if err != nil {
		return 0, false, err
	}
	if busy <= o.Max {
		line(busy, "ok")
		return busy, true, nil
	}
	line(busy, "busy")
	if o.Wait <= 0 {
		return busy, false, nil
	}
	fmt.Fprintf(out, "waiting up to %v, checking every %v\n", o.Wait, o.Every)
	deadline := start.Add(o.Wait)
	for {
		left := deadline.Sub(o.Now())
		if left <= 0 {
			line(busy, fmt.Sprintf("still busy after %v", o.Wait))
			return busy, false, nil
		}
		o.Sleep(min(o.Every, left))
		if busy, err = measure(o); err != nil {
			return 0, false, err
		}
		if busy <= o.Max {
			line(busy, "ok")
			return busy, true, nil
		}
	}
}
