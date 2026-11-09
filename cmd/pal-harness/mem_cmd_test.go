package main

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeMem replaces the memory reader and the sleeper for one test.
func fakeMem(t *testing.T, readings ...uint64) {
	t.Helper()
	oldRead, oldSleep := memRead, memSleep
	n := 0
	memRead = func() (uint64, error) {
		v := readings[min(n, len(readings)-1)]
		n++
		return v, nil
	}
	clock := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	oldNow := now
	now = func() time.Time { return clock }
	memSleep = func(d time.Duration) { clock = clock.Add(d) }
	t.Cleanup(func() { memRead, memSleep, now = oldRead, oldSleep, oldNow })
}

func TestMemCommand(t *testing.T) {
	fakeMem(t, 3<<30)
	code, out, _ := do(t, "mem")
	if code != 0 || out != "free memory: 3.0 GiB (minimum 2.0 GiB): ok\n" {
		t.Fatalf("%d %q", code, out)
	}
}

func TestMemCommandShort(t *testing.T) {
	fakeMem(t, 1<<30)
	code, out, _ := do(t, "mem", "--min", "1536MiB")
	if code != 1 || !strings.Contains(out, "(minimum 1.5 GiB): short") {
		t.Fatalf("%d %q", code, out)
	}
}

func TestMemCommandWaits(t *testing.T) {
	fakeMem(t, 1<<30, 1<<30, 4<<30)
	code, out, _ := do(t, "mem", "--wait", "9m", "--every", "10s")
	if code != 0 || !strings.Contains(out, "waiting up to 9m0s, checking every 10s") || !strings.HasSuffix(out, "ok\n") {
		t.Fatalf("%d %q", code, out)
	}
	fakeMem(t, 1<<30)
	if code, out, _ = do(t, "mem", "--wait", "1m"); code != 1 || !strings.Contains(out, "still short after 1m0s") {
		t.Fatalf("%d %q", code, out)
	}
}

func TestMemCommandErrors(t *testing.T) {
	fakeMem(t, 3<<30)
	for _, c := range [][]string{{"mem", "x"}, {"mem", "--min", "lots"}, {"mem", "--wait", "soon"}, {"mem", "--nope"},
		{"mem", "--every", "-1s"}, {"mem", "--wait", "-1m"}} {
		if code, _, errs := do(t, c...); code != 2 || errs == "" {
			t.Errorf("%v: code %d err=%q", c, code, errs)
		}
	}
	memRead = func() (uint64, error) { return 0, errors.New("unsupported") }
	if code, _, errs := do(t, "mem"); code != 2 || !strings.Contains(errs, "unsupported") {
		t.Errorf("read error: code %d err=%q", code, errs)
	}
}
