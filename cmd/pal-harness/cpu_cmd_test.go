package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/althenlimzixuan/pal-harness/cpu"
)

// fakeCPU replaces the CPU reader and the sleeper for one test. Each share is
// one measurement (two reads).
func fakeCPU(t *testing.T, shares ...float64) {
	t.Helper()
	oldRead, oldSleep, oldNow := cpuRead, cpuSleep, now
	n := 0
	var idle, total uint64
	cpuRead = func() (cpu.Times, error) {
		if n%2 == 1 {
			total += 1000
			idle += uint64(1000 - shares[min(n/2, len(shares)-1)]*10)
		}
		n++
		return cpu.Times{Idle: idle, Total: total}, nil
	}
	clock := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
	now = func() time.Time { return clock }
	cpuSleep = func(d time.Duration) { clock = clock.Add(d) }
	t.Cleanup(func() { cpuRead, cpuSleep, now = oldRead, oldSleep, oldNow })
}

func TestCPUCommand(t *testing.T) {
	fakeCPU(t, 7)
	code, out, _ := do(t, "cpu")
	if code != 0 || out != "cpu busy: 7% over 10s (maximum 25%): ok\n" {
		t.Fatalf("%d %q", code, out)
	}
}

func TestCPUCommandBusy(t *testing.T) {
	fakeCPU(t, 30)
	code, out, _ := do(t, "cpu", "--max", "20%", "--for", "5s")
	if code != 1 || out != "cpu busy: 30% over 5s (maximum 20%): busy\n" {
		t.Fatalf("%d %q", code, out)
	}
}

func TestCPUCommandWaits(t *testing.T) {
	fakeCPU(t, 63, 10)
	code, out, _ := do(t, "cpu", "--wait", "9m", "--every", "20s")
	if code != 0 || !strings.Contains(out, "waiting up to 9m0s, checking every 20s") || !strings.HasSuffix(out, "ok\n") {
		t.Fatalf("%d %q", code, out)
	}
	fakeCPU(t, 63)
	if code, out, _ = do(t, "cpu", "--wait", "1m"); code != 1 || !strings.Contains(out, "still busy after 1m0s") {
		t.Fatalf("%d %q", code, out)
	}
}

func TestCPUCommandUnreadable(t *testing.T) {
	old := cpuRead
	cpuRead = func() (cpu.Times, error) { return cpu.Times{}, errors.New("no counters") }
	t.Cleanup(func() { cpuRead = old })
	code, _, errs := do(t, "cpu")
	if code != 2 || !strings.Contains(errs, "cannot read CPU use: no counters") {
		t.Fatalf("%d %q", code, errs)
	}
}

func TestCPUCommandErrors(t *testing.T) {
	fakeCPU(t, 7)
	for _, c := range [][]string{{"cpu", "x"}, {"cpu", "--max", "lots"}, {"cpu", "--max", "120%"},
		{"cpu", "--for", "0s"}, {"cpu", "--wait", "-1m"}, {"cpu", "--every", "0s"}, {"cpu", "--nope"}} {
		if code, _, errs := do(t, c...); code != 2 || errs == "" {
			t.Errorf("%v: code %d err=%q", c, code, errs)
		}
	}
}
