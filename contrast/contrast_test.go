package contrast

import (
	"math"
	"strings"
	"testing"
)

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestRatioKnownPairs(t *testing.T) {
	cases := []struct {
		fg, bg string
		want   float64
	}{
		{"#000", "#fff", 21},
		{"#000000", "#FFFFFF", 21},
		{"#fff", "#fff", 1},
		{"#767676", "#ffffff", 4.54},
		{"#777777", "#ffffff", 4.48},
		{"rgb(255, 0, 0)", "#fff", 4.00},
		{"rgb(255 0 0)", "rgb(255,255,255)", 4.00},
		{"oklch(1 0 0)", "oklch(0 0 0)", 21},
		{"oklch(100% 0 0)", "#000", 21},
		// CSS Color 4's red in OKLCH.
		{"oklch(62.8% 0.2577 29.23)", "#fff", 4.00},
		{"OKLCH(0.628 0.2577 29.23deg)", "#ffffff", 4.00},
	}
	for _, c := range cases {
		fg, err := Parse(c.fg)
		if err != nil {
			t.Fatalf("Parse(%q): %v", c.fg, err)
		}
		bg, err := Parse(c.bg)
		if err != nil {
			t.Fatalf("Parse(%q): %v", c.bg, err)
		}
		got := Ratio(fg, bg)
		if !near(got, c.want, 0.01) {
			t.Errorf("Ratio(%s, %s) = %.4f, want %.2f", c.fg, c.bg, got, c.want)
		}
		if r := Ratio(bg, fg); !near(r, got, 1e-12) {
			t.Errorf("Ratio is not symmetric for %s, %s", c.fg, c.bg)
		}
	}
}

func TestParseClipsOutOfGamutOKLCH(t *testing.T) {
	c, err := Parse("oklch(0.7 0.4 140)")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Clipped {
		t.Error("oklch(0.7 0.4 140) is outside sRGB but Clipped is false")
	}
	for _, v := range []float64{c.R, c.G, c.B} {
		if v < 0 || v > 1 {
			t.Errorf("channel %v outside 0..1 after clipping", v)
		}
	}
	in, _ := Parse("oklch(0.628 0.2577 29.23)")
	if in.Clipped {
		t.Error("CSS red is inside sRGB but Clipped is true")
	}
}

func TestParseRejects(t *testing.T) {
	for _, s := range []string{
		"", "#12", "#12345", "#ggg", "#aabbccdd", "#abcd",
		"rgb(1, 2)", "rgb(256, 0, 0)", "rgb(1, 2, 3 / 0.5)", "rgba(1, 2, 3, 0.5)",
		"oklch(0.5 0.1)", "oklch(0.5 0.1 20 / 50%)", "oklch(x 0.1 20)", "oklch(150% 0.1 20)",
		"hsl(0 100% 50%)", "red",
		"rgb(nan 0 0)", "rgb(inf 0 0)", "oklch(nan 0.1 20)", "oklch(0.5 inf 20)", "oklch(0.5 0.1 nan)", "oklch(0.5 0.1 inf)",
	} {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) accepted a colour it should refuse", s)
		}
	}
	_, err := Parse("rgb(1 2 3 / 0.5)")
	if err == nil || !strings.Contains(err.Error(), "composite") {
		t.Errorf("alpha error should tell the reader to composite first, got %v", err)
	}
}

func TestVerdict(t *testing.T) {
	cases := []struct {
		ratio       float64
		text, large bool
	}{
		{21, true, true},
		{4.5, true, true},
		{4.499, false, true},
		{3, false, true},
		{2.99, false, false},
	}
	for _, c := range cases {
		v := Check(c.ratio)
		if v.Text != c.text || v.Large != c.large {
			t.Errorf("Check(%v) = %+v, want text %v large %v", c.ratio, v, c.text, c.large)
		}
	}
}

func TestFormatNeverRoundsUpPastAThreshold(t *testing.T) {
	if got := Format(4.4999); !strings.HasPrefix(got, "4.49:1") {
		t.Errorf("Format(4.4999) = %q, want it to start 4.49:1", got)
	}
	if got := Format(21); !strings.HasPrefix(got, "21.00:1") {
		t.Errorf("Format(21) = %q", got)
	}
	if got := Format(4.0); !strings.Contains(got, "text (4.5:1): fail") || !strings.Contains(got, "large text and UI parts (3:1): pass") {
		t.Errorf("Format(4.0) = %q", got)
	}
}
