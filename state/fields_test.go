package state

import (
	"reflect"
	"strings"
	"testing"
)

const stop = "owner_stops:\n  - {kind: spec-review, since: 2026-11-22T10:00:00+08:00, answered: 2026-11-22T11:30:00+08:00, could_decide: true, via: owner}\n"

// pal-harness 1.9.0: the trial measure in the phase state (design A).
func TestPhaseStateOwnerStops(t *testing.T) {
	base := string(read(t, "phase-state.yaml"))
	ps, findings := ParsePhaseState([]byte(base + stop))
	if len(findings) != 0 {
		t.Fatalf("valid owner_stops: findings %v", findings)
	}
	if len(ps.OwnerStops) != 1 || ps.OwnerStops[0].Kind != "spec-review" || !ps.OwnerStops[0].CouldDecide ||
		ps.OwnerStops[0].Via != "owner" || ps.OwnerStops[0].Answered.Sub(ps.OwnerStops[0].Since).Minutes() != 90 {
		t.Fatalf("owner_stops = %+v", ps.OwnerStops)
	}
	for _, v := range ViaValues {
		if _, f := ParsePhaseState([]byte(base + strings.Replace(stop, "via: owner", "via: "+v, 1))); len(f) != 0 {
			t.Errorf("via %s: findings %v", v, f)
		}
	}
	rep := func(old, new string) string {
		if !strings.Contains(stop, old) {
			t.Fatalf("stop lacks %q", old)
		}
		return base + strings.Replace(stop, old, new, 1)
	}
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"kind", rep("kind: spec-review", "kind: lunch"), []string{`owner_stops 1: kind "lunch" is not one of spec-review, *`}},
		{"since", rep("since: 2026-11-22T10:00:00+08:00, ", ""), []string{"owner_stops 1: since is required"}},
		{"answered", rep("answered: 2026-11-22T11:30:00+08:00, ", ""), []string{"owner_stops 1: answered is required"}},
		{"order", rep("answered: 2026-11-22T11:30:00+08:00", "answered: 2026-11-22T09:00:00+08:00"),
			[]string{"owner_stops 1: answered is before since"}},
		{"via", rep("via: owner", "via: slack"), []string{"owner_stops 1: via must be owner, approved, defaulted, changed or vetoed"}},
		{"via without decision", rep("could_decide: true, via: owner", "could_decide: false, via: defaulted"),
			[]string{"owner_stops 1: via defaulted needs could_decide: true"}},
		{"missing via", rep(", via: owner", ""), []string{"owner_stops 1: via must be owner, approved, defaulted, changed or vetoed"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, got := ParsePhaseState([]byte(tc.in))
			check(t, got, tc.want)
		})
	}
	_, got := ParsePhaseState([]byte(rep("via: owner", "via: owner, mood: ok")))
	if len(got) != 1 || !strings.Contains(got[0],
		"field mood not found in type state.OwnerStop (allowed fields: kind, since, answered, could_decide, via)") {
		t.Errorf("unknown key in an entry = %q", got)
	}
}

func TestParseVersion(t *testing.T) {
	for _, in := range []string{"1.8.0", "v1.8.0"} {
		v, err := ParseVersion(in)
		if err != nil || v != (Version{1, 8, 0}) || v.String() != "1.8.0" {
			t.Errorf("ParseVersion(%q) = %v, %v", in, v, err)
		}
	}
	for _, in := range []string{"", "1.8", "x", "1.8.x", "1.-1.0", "1.8.0.1"} {
		if _, err := ParseVersion(in); err == nil {
			t.Errorf("ParseVersion(%q) took a bad version", in)
		}
	}
	if !(Version{1, 8, 0}).Less(Version{1, 9, 0}) || (Version{1, 10, 0}).Less(Version{1, 9, 0}) ||
		(Version{1, 9, 0}).Less(Version{1, 9, 0}) || !(Version{1, 9, 0}).Less(Version{2, 0, 0}) ||
		!(Version{1, 9, 0}).Less(Version{1, 9, 1}) {
		t.Error("Less orders versions wrongly")
	}
}

// pal-harness 1.9.0: which version added each field, for check --reader (design B2).
func TestNewerFields(t *testing.T) {
	v := func(s string) Version {
		t.Helper()
		r, err := ParseVersion(s)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	ps := []byte(string(read(t, "phase-state.yaml")) + "qa: passed\n" + stop)
	tests := []struct {
		name   string
		kind   string
		in     []byte
		reader string
		want   []Newer
	}{
		{"reader 1.8.0", "phase-state", ps, "1.8.0", []Newer{{"owner_stops", "1.9.0"}}},
		{"reader 1.6.0", "phase-state", ps, "1.6.0", []Newer{{"owner_stops", "1.9.0"}, {"qa", "1.8.0"}}},
		{"reader 1.9.0", "phase-state", ps, "1.9.0", nil},
		{"older fields only", "phase-state", read(t, "phase-state.yaml"), "1.3.0", nil},
		{"features check", "features",
			[]byte(strings.Replace(validFeature, "results: []", "results:\n    - {date: 2026-11-19, env: local, check: offline-banner, result: pass}", 1)),
			"1.6.0", []Newer{{"results.check", "1.8.0"}}},
		{"features step", "features",
			[]byte(strings.Replace(validFeature, "results: []", "results:\n    - {date: 2026-11-19, env: local, step: 1, result: pass}", 1)),
			"1.6.0", nil},
		{"broken YAML", "phase-state", []byte("phase: [\n"), "1.0.0", nil},
		{"unknown kind", "other", ps, "1.0.0", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NewerFields(tc.kind, tc.in, v(tc.reader)); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("NewerFields = %v, want %v", got, tc.want)
			}
		})
	}
}
