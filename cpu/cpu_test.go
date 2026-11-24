package cpu

import (
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestBusy(t *testing.T) {
	a := Times{Idle: 100, Total: 200}
	for _, c := range []struct {
		b    Times
		want float64
	}{
		{Times{Idle: 190, Total: 300}, 10},  // 90 idle of 100
		{Times{Idle: 100, Total: 300}, 100}, // no idle
		{Times{Idle: 200, Total: 300}, 0},   // all idle
		{Times{Idle: 100, Total: 200}, 0},   // no time passed
	} {
		if got := Busy(a, c.b); got != c.want {
			t.Errorf("Busy(%v, %v) = %v, want %v", a, c.b, got, c.want)
		}
	}
}

// Windows' kernel time already includes idle time, so the total is kernel plus user.
func TestFromSystemTimes(t *testing.T) {
	got := fromSystemTimes(70, 100, 50)
	if got != (Times{Idle: 70, Total: 150}) {
		t.Fatalf("got %+v", got)
	}
	a, b := fromSystemTimes(0, 0, 0), fromSystemTimes(70, 100, 50) // 70 idle of 150
	if busy := Busy(a, b); busy < 53.3 || busy > 53.4 {
		t.Fatalf("busy %v", busy)
	}
}

func TestParseProcStat(t *testing.T) {
	in := "cpu  100 5 50 800 20 3 2 0 40 0\ncpu0 50 2 25 400 10 1 1 0 0 0\nintr 1\n"
	got, err := parseProcStat(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	// idle = idle + iowait; total = the first 8 fields (guest is already inside user).
	if got != (Times{Idle: 820, Total: 980}) {
		t.Fatalf("got %+v", got)
	}
	for _, bad := range []string{"", "intr 1\n", "cpu  1 2 x 4 5\n", "cpu  1 2 3\n"} {
		if _, err := parseProcStat(strings.NewReader(bad)); err == nil {
			t.Errorf("%q parsed", bad)
		}
	}
}

func TestParsePercent(t *testing.T) {
	for in, want := range map[string]float64{"25%": 25, "25": 25, " 7.5% ": 7.5, "0": 0, "100%": 100} {
		if got, err := ParsePercent(in); err != nil || got != want {
			t.Errorf("ParsePercent(%q) = %v, %v", in, got, err)
		}
	}
	for _, in := range []string{"", "%", "abc", "-1%", "101%", "2 5%"} {
		if _, err := ParsePercent(in); err == nil {
			t.Errorf("ParsePercent(%q) succeeded", in)
		}
	}
}

// fake gives busy shares in turn: each measurement takes two reads.
type fake struct {
	busy  []float64
	err   error
	n     int
	total uint64
	idle  uint64
	clock time.Time
	slept []time.Duration
}

func (f *fake) read() (Times, error) {
	if f.err != nil {
		return Times{}, f.err
	}
	// Even reads start a measurement; odd reads end it with the next busy share.
	if f.n%2 == 1 {
		b := f.busy[min(f.n/2, len(f.busy)-1)]
		f.total += 1000
		f.idle += uint64(1000 - b*10)
	}
	f.n++
	return Times{Idle: f.idle, Total: f.total}, nil
}
func (f *fake) now() time.Time        { return f.clock }
func (f *fake) sleep(d time.Duration) { f.slept = append(f.slept, d); f.clock = f.clock.Add(d) }
func (f *fake) opts(wait time.Duration) Options {
	return Options{Max: 25, For: 10 * time.Second, Wait: wait, Every: 30 * time.Second, Read: f.read, Now: f.now, Sleep: f.sleep}
}

func TestCheckQuiet(t *testing.T) {
	f := &fake{busy: []float64{7}}
	var out strings.Builder
	busy, ok, err := Check(f.opts(0), &out)
	if err != nil || !ok || busy != 7 {
		t.Fatalf("%v %v %v", busy, ok, err)
	}
	if out.String() != "cpu busy: 7% over 10s (maximum 25%): ok\n" {
		t.Fatalf("%q", out.String())
	}
	if len(f.slept) != 1 || f.slept[0] != 10*time.Second {
		t.Fatalf("slept %v", f.slept)
	}
}

func TestCheckBusy(t *testing.T) {
	f := &fake{busy: []float64{63}}
	var out strings.Builder
	busy, ok, err := Check(f.opts(0), &out)
	if err != nil || ok || busy != 63 {
		t.Fatalf("%v %v %v", busy, ok, err)
	}
	if out.String() != "cpu busy: 63% over 10s (maximum 25%): busy\n" {
		t.Fatalf("%q", out.String())
	}
}

// At exactly the maximum the machine counts as quiet.
func TestCheckAtMaximum(t *testing.T) {
	f := &fake{busy: []float64{25}}
	if _, ok, err := Check(f.opts(0), &strings.Builder{}); err != nil || !ok {
		t.Fatalf("%v %v", ok, err)
	}
}

func TestCheckWaitsUntilQuiet(t *testing.T) {
	f := &fake{busy: []float64{63, 40, 12}}
	var out strings.Builder
	busy, ok, err := Check(f.opts(9*time.Minute), &out)
	if err != nil || !ok || busy != 12 {
		t.Fatalf("%v %v %v", busy, ok, err)
	}
	s := out.String()
	if !strings.Contains(s, "waiting up to 9m0s, checking every 30s") || !strings.HasSuffix(s, "cpu busy: 12% over 10s (maximum 25%): ok\n") {
		t.Fatalf("%q", s)
	}
}

func TestCheckStillBusy(t *testing.T) {
	f := &fake{busy: []float64{63}}
	var out strings.Builder
	busy, ok, err := Check(f.opts(time.Minute), &out)
	if err != nil || ok || busy != 63 {
		t.Fatalf("%v %v %v", busy, ok, err)
	}
	if !strings.HasSuffix(out.String(), "cpu busy: 63% over 10s (maximum 25%): still busy after 1m0s\n") {
		t.Fatalf("%q", out.String())
	}
	if f.clock.Sub(time.Time{}) < time.Minute {
		t.Fatalf("stopped early at %v", f.clock)
	}
}

func TestCheckReadError(t *testing.T) {
	f := &fake{err: errors.New("no counters")}
	if _, _, err := Check(f.opts(0), &strings.Builder{}); err == nil {
		t.Fatal("no error")
	}
	// A read that fails mid-wait is an error too.
	g := &fake{busy: []float64{63}}
	o := g.opts(time.Minute)
	read := o.Read
	o.Read = func() (Times, error) {
		if g.n >= 2 {
			return Times{}, errors.New("gone")
		}
		return read()
	}
	if _, _, err := Check(o, &strings.Builder{}); err == nil {
		t.Fatal("no error mid-wait")
	}
}

// TestRead reads the real machine on the platforms it supports.
func TestRead(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "linux" {
		if _, err := Read(); err == nil {
			t.Fatal("unsupported platform read without error")
		}
		return
	}
	v, err := Read()
	if err != nil {
		t.Fatal(err)
	}
	if v.Total == 0 || v.Idle > v.Total {
		t.Fatalf("bad times %+v", v)
	}
}
