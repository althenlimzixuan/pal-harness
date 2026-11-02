// Package contrast measures the WCAG 2 contrast ratio of two CSS colours, so a
// contrast finding carries a measured ratio, never an estimate.
package contrast

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Color is an opaque sRGB colour, each channel gamma-encoded in 0..1. Clipped
// reports an oklch() value that fell outside sRGB and was clipped into it.
type Color struct {
	R, G, B float64
	Clipped bool
}

// The WCAG 2 AA thresholds.
const (
	TextMin  = 4.5 // normal text
	LargeMin = 3.0 // large text and UI parts
)

var errAlpha = errors.New("colour has alpha: composite it over its background first")

// Parse reads #rgb, #rrggbb, rgb(r g b), rgb(r, g, b) and oklch(L C H), where L
// is 0..1 or a percentage and H is in degrees. Alpha is refused.
func Parse(s string) (Color, error) {
	in := strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasPrefix(in, "#"):
		return parseHex(in[1:], s)
	case strings.HasPrefix(in, "rgba("):
		return Color{}, errAlpha
	case strings.HasPrefix(in, "rgb("):
		args, err := fnArgs(in, "rgb(")
		if err != nil {
			return Color{}, fmt.Errorf("%q: %w", s, err)
		}
		var c [3]float64
		for i, a := range args {
			v, err := finite(a)
			if err != nil || v < 0 || v > 255 {
				return Color{}, fmt.Errorf("%q: rgb() channels are numbers 0..255", s)
			}
			c[i] = v / 255
		}
		return Color{R: c[0], G: c[1], B: c[2]}, nil
	case strings.HasPrefix(in, "oklch("):
		args, err := fnArgs(in, "oklch(")
		if err != nil {
			return Color{}, fmt.Errorf("%q: %w", s, err)
		}
		return parseOKLCH(args, s)
	}
	return Color{}, fmt.Errorf("%q: give #rgb, #rrggbb, rgb() or oklch()", s)
}

func parseHex(h, orig string) (Color, error) {
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) == 4 || len(h) == 8 {
		return Color{}, errAlpha
	}
	if len(h) != 6 {
		return Color{}, fmt.Errorf("%q: a hex colour has 3 or 6 digits", orig)
	}
	n, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return Color{}, fmt.Errorf("%q: not a hex colour", orig)
	}
	return Color{R: float64(n>>16&0xff) / 255, G: float64(n>>8&0xff) / 255, B: float64(n&0xff) / 255}, nil
}

// fnArgs splits "name(a b c)" or "name(a, b, c)" into exactly three arguments.
func fnArgs(in, name string) ([]string, error) {
	if !strings.HasSuffix(in, ")") {
		return nil, errors.New("missing )")
	}
	body := in[len(name) : len(in)-1]
	if strings.Contains(body, "/") {
		return nil, errAlpha
	}
	args := strings.Fields(strings.ReplaceAll(body, ",", " "))
	if len(args) != 3 {
		return nil, errors.New("want three values")
	}
	return args, nil
}

func parseOKLCH(args []string, orig string) (Color, error) {
	bad := fmt.Errorf("%q: oklch() takes L (0..1 or 0%%..100%%), C and H (degrees)", orig)
	l, err := number(args[0], 0.01)
	if err != nil || l < 0 || l > 1 {
		return Color{}, bad
	}
	ch, err := number(args[1], 0.004) // CSS: 100% chroma is 0.4
	if err != nil || ch < 0 {
		return Color{}, bad
	}
	h, err := finite(strings.TrimSuffix(args[2], "deg"))
	if err != nil {
		return Color{}, bad
	}
	rad := h * math.Pi / 180
	a, b := ch*math.Cos(rad), ch*math.Sin(rad)

	// OKLab to linear sRGB (Björn Ottosson's matrices, as in CSS Color 4).
	l3 := cube(l + 0.3963377774*a + 0.2158037573*b)
	m3 := cube(l - 0.1055613458*a - 0.0638541728*b)
	s3 := cube(l - 0.0894841775*a - 1.2914855480*b)
	lin := [3]float64{
		+4.0767416621*l3 - 3.3077115913*m3 + 0.2309699292*s3,
		-1.2684380046*l3 + 2.6097574011*m3 - 0.3413193965*s3,
		-0.0041960863*l3 - 0.7034186147*m3 + 1.7076147010*s3,
	}
	var c Color
	var out [3]float64
	const tol = 1e-3 // oklch() values written to 3-4 digits land just outside sRGB
	for i, v := range lin {
		if v < -tol || v > 1+tol {
			c.Clipped = true
		}
		out[i] = encode(math.Min(1, math.Max(0, v)))
	}
	c.R, c.G, c.B = out[0], out[1], out[2]
	return c, nil
}

// finite parses a float and refuses NaN and infinities.
func finite(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err == nil && (math.IsNaN(v) || math.IsInf(v, 0)) {
		err = errors.New("not a finite number")
	}
	return v, err
}

// number reads a plain number or a percentage scaled by pct.
func number(s string, pct float64) (float64, error) {
	if p, ok := strings.CutSuffix(s, "%"); ok {
		v, err := finite(p)
		return v * pct, err
	}
	return finite(s)
}

func cube(x float64) float64 { return x * x * x }

// encode applies the sRGB transfer function to a linear channel.
func encode(v float64) float64 {
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return 1.055*math.Pow(v, 1/2.4) - 0.055
}

// decode undoes encode.
func decode(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// Luminance is the WCAG relative luminance.
func Luminance(c Color) float64 {
	return 0.2126*decode(c.R) + 0.7152*decode(c.G) + 0.0722*decode(c.B)
}

// Ratio is the WCAG contrast ratio, 1 to 21, in either order.
func Ratio(a, b Color) float64 {
	la, lb := Luminance(a), Luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// Verdict says which AA thresholds a ratio meets.
type Verdict struct{ Text, Large bool }

// Check compares a ratio with the AA thresholds, unrounded.
func Check(ratio float64) Verdict {
	return Verdict{Text: ratio >= TextMin, Large: ratio >= LargeMin}
}

// Format prints the ratio truncated to two decimals (so 4.499 never shows as
// 4.50) and the AA verdicts.
func Format(ratio float64) string {
	v := Check(ratio)
	word := func(ok bool) string {
		if ok {
			return "pass"
		}
		return "fail"
	}
	return fmt.Sprintf("%.2f:1 — text (4.5:1): %s · large text and UI parts (3:1): %s",
		math.Floor(ratio*100)/100, word(v.Text), word(v.Large))
}
