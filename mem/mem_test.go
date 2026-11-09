package mem

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseSize(t *testing.T) {
	for in, want := range map[string]uint64{
		"2GiB":   2 << 30,
		"512MiB": 512 << 20,
		"1.5GiB": 3 << 29,
		"64KiB":  64 << 10,
		"100":    100,
		" 2GiB ": 2 << 30,
	} {
		got, err := ParseSize(in)
		if err != nil || got != want {
			t.Errorf("ParseSize(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	for _, in := range []string{"", "GiB", "2XB", "-1GiB", "abc", "2 GiB"} {
		if _, err := ParseSize(in); err == nil {
			t.Errorf("ParseSize(%q) succeeded", in)
		}
	}
}

func TestFormat(t *testing.T) {
	for in, want := range map[uint64]string{2 << 30: "2.0 GiB", 3 << 29: "1.5 GiB", 512 << 20: "512 MiB", 1000: "0 MiB"} {
		if got := Format(in); got != want {
			t.Errorf("Format(%d) = %q, want %q", in, got, want)
		}
	}
}

// fake gives readings in turn and a clock that sleeps instantly.
type fake struct {
	readings []uint64
	err      error
	n        int
	clock    time.Time
	slept    []time.Duration
}

func (f *fake) read() (uint64, error) {
	if f.err != nil {
		return 0, f.err
	}
	v := f.readings[min(f.n, len(f.readings)-1)]
	f.n++
	return v, nil
}
func (f *fake) now() time.Time        { return f.clock }
func (f *fake) sleep(d time.Duration) { f.slept = append(f.slept, d); f.clock = f.clock.Add(d) }
func (f *fake) opts(wait time.Duration) Options {
	return Options{Min: 2 << 30, Wait: wait, Every: 30 * time.Second, Read: f.read, Now: f.now, Sleep: f.sleep}
}

func TestCheckEnoughAtOnce(t *testing.T) {
	f := &fake{readings: []uint64{3 << 30}}
	var out strings.Builder
	ok, err := Check(f.opts(0), &out)
	if err != nil || !ok {
		t.Fatalf("Check = %v, %v", ok, err)
	}
	if want := "free memory: 3.0 GiB (minimum 2.0 GiB): ok\n"; out.String() != want {
		t.Errorf("output %q, want %q", out.String(), want)
	}
	if len(f.slept) != 0 {
		t.Errorf("slept %v", f.slept)
	}
}

func TestCheckShortNoWait(t *testing.T) {
	f := &fake{readings: []uint64{1 << 30}}
	var out strings.Builder
	ok, err := Check(f.opts(0), &out)
	if err != nil || ok {
		t.Fatalf("Check = %v, %v", ok, err)
	}
	if want := "free memory: 1.0 GiB (minimum 2.0 GiB): short\n"; out.String() != want {
		t.Errorf("output %q, want %q", out.String(), want)
	}
}

func TestCheckRecoversWithinWait(t *testing.T) {
	f := &fake{readings: []uint64{1 << 30, 1 << 30, 3 << 30}}
	var out strings.Builder
	ok, err := Check(f.opts(9*time.Minute), &out)
	if err != nil || !ok {
		t.Fatalf("Check = %v, %v", ok, err)
	}
	if len(f.slept) != 2 {
		t.Errorf("slept %d times, want 2", len(f.slept))
	}
	if !strings.HasSuffix(out.String(), "free memory: 3.0 GiB (minimum 2.0 GiB): ok\n") ||
		!strings.Contains(out.String(), "waiting up to 9m0s") {
		t.Errorf("output %q", out.String())
	}
}

func TestCheckStillShortAfterWait(t *testing.T) {
	f := &fake{readings: []uint64{1 << 30}}
	var out strings.Builder
	ok, err := Check(f.opts(2*time.Minute), &out)
	if err != nil || ok {
		t.Fatalf("Check = %v, %v", ok, err)
	}
	var total time.Duration
	for _, d := range f.slept {
		total += d
	}
	if total != 2*time.Minute {
		t.Errorf("slept %v in all, want 2m", total)
	}
	if !strings.HasSuffix(out.String(), "still short after 2m0s\n") {
		t.Errorf("output %q", out.String())
	}
}

func TestCheckLastSleepTrimmedToWait(t *testing.T) {
	f := &fake{readings: []uint64{1 << 30}}
	o := f.opts(45 * time.Second)
	var out strings.Builder
	if ok, _ := Check(o, &out); ok {
		t.Fatal("ok")
	}
	if len(f.slept) != 2 || f.slept[0] != 30*time.Second || f.slept[1] != 15*time.Second {
		t.Errorf("slept %v, want [30s 15s]", f.slept)
	}
}

func TestCheckReadError(t *testing.T) {
	f := &fake{err: errors.New("boom")}
	var out strings.Builder
	if _, err := Check(f.opts(0), &out); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckDefaults(t *testing.T) {
	// A zero Every falls back to 30 s; nil Now and Sleep use the real clock (not reached here).
	f := &fake{readings: []uint64{1 << 30, 3 << 30}}
	o := f.opts(time.Minute)
	o.Every = 0
	var out strings.Builder
	if ok, err := Check(o, &out); err != nil || !ok {
		t.Fatalf("Check = %v, %v", ok, err)
	}
	if len(f.slept) != 1 || f.slept[0] != 30*time.Second {
		t.Errorf("slept %v", f.slept)
	}
}

func TestParseMeminfo(t *testing.T) {
	sample := "MemTotal:       16303424 kB\nMemFree:          812344 kB\nMemAvailable:    5242880 kB\nBuffers:  1 kB\n"
	got, err := parseMeminfo(strings.NewReader(sample))
	if err != nil || got != 5242880*1024 {
		t.Fatalf("parseMeminfo = %d, %v", got, err)
	}
	for _, bad := range []string{"MemTotal: 1 kB\n", "MemAvailable: x kB\n"} {
		if _, err := parseMeminfo(strings.NewReader(bad)); err == nil {
			t.Errorf("parseMeminfo(%q) succeeded", bad)
		}
	}
}
