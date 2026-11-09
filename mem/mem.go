// Package mem checks the machine's available physical memory before a heavy
// run (test reports, full test runs, native builds), so a session can free
// memory or wait instead of having the run killed mid-way.
package mem

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Options configures Check. Read, Now and Sleep are injected so tests need no
// real clock or OS; nil Now and Sleep use the real clock.
type Options struct {
	Min   uint64        // bytes that must be available
	Wait  time.Duration // how long to keep re-checking while short; 0 checks once
	Every time.Duration // the pause between checks; 0 means 30 s
	Read  func() (uint64, error)
	Now   func() time.Time
	Sleep func(time.Duration)
}

// Check reports whether at least o.Min bytes are available, re-checking for up
// to o.Wait while short. It writes one line per verdict to out.
func Check(o Options, out io.Writer) (bool, error) {
	if o.Every <= 0 {
		o.Every = 30 * time.Second
	}
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Sleep == nil {
		o.Sleep = time.Sleep
	}
	v, err := o.Read()
	if err != nil {
		return false, err
	}
	if v >= o.Min {
		fmt.Fprintf(out, "free memory: %s (minimum %s): ok\n", Format(v), Format(o.Min))
		return true, nil
	}
	fmt.Fprintf(out, "free memory: %s (minimum %s): short\n", Format(v), Format(o.Min))
	if o.Wait <= 0 {
		return false, nil
	}
	fmt.Fprintf(out, "waiting up to %v, checking every %v\n", o.Wait, o.Every)
	deadline := o.Now().Add(o.Wait)
	for {
		left := deadline.Sub(o.Now())
		if left <= 0 {
			fmt.Fprintf(out, "free memory: %s (minimum %s): still short after %v\n", Format(v), Format(o.Min), o.Wait)
			return false, nil
		}
		o.Sleep(min(o.Every, left))
		if v, err = o.Read(); err != nil {
			return false, err
		}
		if v >= o.Min {
			fmt.Fprintf(out, "free memory: %s (minimum %s): ok\n", Format(v), Format(o.Min))
			return true, nil
		}
	}
}

var units = []struct {
	suffix string
	shift  uint
}{{"GiB", 30}, {"MiB", 20}, {"KiB", 10}, {"B", 0}}

// ParseSize reads a size such as "2GiB", "512MiB", "1.5GiB" or a plain byte count.
func ParseSize(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	num, shift := s, uint(0)
	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			num, shift = strings.TrimSuffix(s, u.suffix), u.shift
			break
		}
	}
	f, err := strconv.ParseFloat(num, 64)
	if err != nil || f < 0 || num == "" {
		return 0, fmt.Errorf("bad size %q: want a number with KiB, MiB or GiB, such as 2GiB", s)
	}
	return uint64(f * float64(uint64(1)<<shift)), nil
}

// Format prints bytes as GiB with one decimal from 1 GiB up, else whole MiB.
func Format(b uint64) string {
	if b >= 1<<30 {
		return fmt.Sprintf("%.1f GiB", float64(b)/(1<<30))
	}
	return fmt.Sprintf("%d MiB", b>>20)
}

// parseMeminfo returns MemAvailable from a Linux /proc/meminfo, in bytes.
func parseMeminfo(r io.Reader) (uint64, error) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) >= 2 && f[0] == "MemAvailable:" {
			kb, err := strconv.ParseUint(f[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("bad MemAvailable line %q", sc.Text())
			}
			return kb * 1024, nil
		}
	}
	return 0, errors.New("no MemAvailable line in /proc/meminfo")
}
