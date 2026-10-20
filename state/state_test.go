package state

import (
	"os"
	"strings"
	"testing"
)

const validFeature = `- id: F-001
  name: Thing
  label: need
  why: w
  use_case: u
  status: todo
  phase: ""
  test_script:
    - {step: 1, do: "d", expect: "e"}
    - {step: 2, do: "d", expect: "e"}
  results: []
`

func read(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParseFeaturesValid(t *testing.T) {
	fs, findings := ParseFeatures(read(t, "features.yaml"))
	if len(findings) != 0 {
		t.Fatalf("findings: %v", findings)
	}
	if len(fs) != 2 {
		t.Fatalf("want 2 features, got %d", len(fs))
	}
	if got := fs[0].TestScript[0].Expect; got != "The budget shows 500 on the trip page" {
		t.Errorf("expect = %q", got)
	}
	if fs[1].Results[0].Env != "staging" || fs[1].Results[0].Date != "2026-10-20" {
		t.Errorf("result = %+v", fs[1].Results[0])
	}
}

// check compares findings with want; a want ending in "*" matches by prefix.
func check(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("findings = %q, want %q", got, want)
	}
	for i, w := range want {
		if p, ok := strings.CutSuffix(w, "*"); ok {
			if !strings.HasPrefix(got[i], p) {
				t.Errorf("finding %d = %q, want prefix %q", i, got[i], p)
			}
		} else if got[i] != w {
			t.Errorf("finding %d = %q, want %q", i, got[i], w)
		}
	}
}

func TestValidateFeaturesFindings(t *testing.T) {
	rep := func(old, new string) string { return strings.Replace(validFeature, old, new, 1) }
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"valid", validFeature, nil},
		{"bad id", rep("F-001", "F12"), []string{"F12: id must look like F-012"}},
		{"duplicate", validFeature + validFeature, []string{"F-001: duplicate id"}},
		{"label", rep("label: need", "label: maybe"), []string{"F-001: label must be need or might"}},
		{"status", rep("status: todo", "status: done"), []string{`F-001: status "done" is not one of todo, recommended, *`}},
		{"phase", rep(`phase: ""`, `phase: "96"`), []string{"F-001: phase must be empty or three digits, like 096"}},
		{"steps", rep("step: 2", "step: 3"), []string{"F-001: test_script steps must be numbered 1, 2, 3, …"}},
		{"empty expect", rep(`expect: "e"}`, `expect: ""}`), []string{"F-001: test_script step 1 has no expect"}},
		{"result env", rep("results: []", "results:\n    - {date: 2026-10-20, env: dev, step: 1, result: pass}"),
			[]string{"F-001: result 1: env must be staging or prod"}},
		{"result step", rep("results: []", "results:\n    - {date: 2026-10-20, env: prod, step: 9, result: fail}"),
			[]string{"F-001: result 1: step 9 is not in the test script"}},
		{"result value", rep("results: []", "results:\n    - {date: 2026-10-20, env: prod, step: 1, result: ok}"),
			[]string{"F-001: result 1: result must be pass or fail"}},
		{"unknown key", rep("label: need", "label: need\n  owner: me"), []string{"invalid YAML*"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, got := ParseFeatures([]byte(tc.in))
			check(t, got, tc.want)
		})
	}
	_, got := ParseFeatures([]byte(rep("label: need", "label: need\n  owner: me")))
	if len(got) != 1 || !strings.Contains(got[0], "line 4") || !strings.Contains(got[0], "owner") {
		t.Errorf("unknown key finding should name line and key: %q", got)
	}
}

func TestParsePhaseStateValid(t *testing.T) {
	ps, findings := ParsePhaseState(read(t, "phase-state.yaml"))
	if len(findings) != 0 {
		t.Fatalf("findings: %v", findings)
	}
	if ps.Stage != "building" || ps.Tasks.Done != 4 || ps.Tasks.Total != 9 || ps.WaitingOnOwner != nil {
		t.Errorf("ps = %+v", ps)
	}
	if ps.Gates.SpecApproved == nil || *ps.Gates.SpecApproved != "2026-10-18" || ps.Gates.SignoffMerged != nil {
		t.Errorf("gates = %+v", ps.Gates)
	}
}

func TestValidatePhaseStateFindings(t *testing.T) {
	base := string(read(t, "phase-state.yaml"))
	rep := func(old, new string) string {
		if !strings.Contains(base, old) {
			t.Fatalf("base lacks %q", old)
		}
		return strings.Replace(base, old, new, 1)
	}
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"phase", rep(`phase: "096"`, `phase: "96"`), []string{"phase must be three digits, like 096"}},
		{"stage", rep("stage: building", "stage: validate"), []string{`stage "validate" is not one of spec, *`}},
		{"vehicle", rep("vehicle: phase", "vehicle: owner-task"), []string{"vehicle must be phase or quick-fix"}},
		{"approach", rep("approach: subagent", "approach: x"), []string{"approach must be inline or subagent"}},
		{"ci", rep("ci: green", "ci: x"), []string{"ci must be green, red, running or none"}},
		{"waiting kind", rep("waiting_on_owner: null", "waiting_on_owner: {kind: lunch, since: 2026-10-20T09:00:00+08:00, link: x, questions: 0}"),
			[]string{`waiting_on_owner.kind "lunch" is not one of *`}},
		{"waiting since", rep("waiting_on_owner: null", "waiting_on_owner: {kind: questions, link: x, questions: 2}"),
			[]string{"waiting_on_owner.since is required"}},
		{"tasks", rep("{done: 4, total: 9}", "{done: 10, total: 9}"), []string{"tasks.done is more than tasks.total"}},
		{"progress", rep("last_progress_at: 2026-10-20T09:12:00+08:00\n", ""), []string{"last_progress_at is required"}},
		{"issues", rep(`"althenlimzixuan/etp#431"`, `"431"`), []string{`bundle.issues: "431" must look like owner/repo#431`}},
		{"features", rep("[F-012]", "[12]"), []string{`bundle.features: "12" must look like F-012`}},
		{"gate", rep("spec_approved: 2026-10-18", "spec_approved: yesterday"), []string{"gates.spec_approved must be a date (YYYY-MM-DD) or null"}},
		{"unknown key", base + "extra: 1\n", []string{"invalid YAML*"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, got := ParsePhaseState([]byte(tc.in))
			check(t, got, tc.want)
		})
	}
}
