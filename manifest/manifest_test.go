package manifest

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func mustReplace(t *testing.T, s, old, repl string) string {
	t.Helper()
	if !strings.Contains(s, old) {
		t.Fatalf("fixture does not contain %q", old)
	}
	return strings.Replace(s, old, repl, 1)
}

func hasProblem(err error, want string) bool {
	var ve *ValidationError
	if !errors.As(err, &ve) {
		return false
	}
	for _, p := range ve.Problems {
		if p == want {
			return true
		}
	}
	return false
}

func TestParseProjectSpecExample(t *testing.T) {
	p, err := ParseProject([]byte(readTestdata(t, "project.yaml")))
	if err != nil {
		t.Fatal(err)
	}
	if p.Key != "fhph" || len(p.Components) != 3 {
		t.Fatalf("unexpected project: %+v", p)
	}
	if !reflect.DeepEqual(p.Components[2].Stack, []string{"tbd"}) {
		t.Fatalf("stack = %v", p.Components[2].Stack)
	}
	if p.Postman.Workspace != nil {
		t.Fatal("workspace should be nil")
	}
}

func TestMarshalProjectRoundTrip(t *testing.T) {
	p, err := ParseProject([]byte(readTestdata(t, "project.yaml")))
	if err != nil {
		t.Fatal(err)
	}
	out, err := MarshalProject(p)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{
		"workspace: null",
		"overrides: []",
		"backlog: { owner: althenlimzixuan, number: 0 }",
		"stack: [go_api]",
		"\n  - folder: api\n",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "\r") {
		t.Error("output contains CR")
	}
	p2, err := ParseProject(out)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p, p2) {
		t.Fatalf("round trip differs:\n%+v\n%+v", p, p2)
	}
	// Marshalling a project with nil overrides still emits [].
	p.Harness.Overrides = nil
	out, err = MarshalProject(p)
	if err != nil || !strings.Contains(string(out), "overrides: []") {
		t.Fatalf("nil overrides: %v\n%s", err, out)
	}
}

func TestMarshalProjectWithOverrideAndWorkspaceID(t *testing.T) {
	p, err := ParseProject([]byte(readTestdata(t, "project.yaml")))
	if err != nil {
		t.Fatal(err)
	}
	id := "ws-123"
	p.Postman.Workspace = &id
	p.Harness.Overrides = []Override{{Item: "a", Source: "b", Reason: "c", Since: "2026-10-17", Upstream: "https://example.com/pr/1"}}
	out, err := MarshalProject(p)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := ParseProject(out)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !reflect.DeepEqual(p, p2) {
		t.Fatalf("round trip differs:\n%+v\n%+v", p, p2)
	}
}

func TestParseProjectRejects(t *testing.T) {
	base := readTestdata(t, "project.yaml")
	noComponents := base[:strings.Index(base, "components:")] + "components: []\n" + base[strings.Index(base, "boards:"):]
	cases := []struct {
		name  string
		yaml  string
		want  string // exact problem; empty means expect a plain decode error naming `field`
		field string
	}{
		{name: "unknown field", yaml: "shape: x\n" + base, field: "shape"},
		{name: "version", yaml: mustReplace(t, base, "version: 1\nkey", "version: 2\nkey"), want: "version: must be 1"},
		{name: "key", yaml: mustReplace(t, base, "key: fhph", "key: FHPH"), want: "key: must match ^[a-z0-9]{2,6}$"},
		{name: "rank", yaml: mustReplace(t, base, "role: consumer\n    rank: 2\n    stack: [nextjs_web]", "role: consumer\n    rank: 1\n    stack: [nextjs_web]"), want: "components[1].rank: must be 2 for role consumer"},
		{name: "role", yaml: mustReplace(t, base, "role: provider", "role: owner"), want: "components[0].role: must be one of provider, platform, consumer"},
		{name: "reserved", yaml: mustReplace(t, base, "folder: api", "folder: doc"), want: `components[0].folder: "doc" is reserved`},
		{name: "reserved underscore", yaml: mustReplace(t, base, "folder: api", "folder: _x"), want: `components[0].folder: "_x" is reserved`},
		{name: "folder pattern", yaml: mustReplace(t, base, "folder: api", "folder: Api"), want: "components[0].folder: must match ^[a-z][a-z0-9_-]{0,29}$"},
		{name: "duplicate", yaml: mustReplace(t, base, "folder: web", "folder: api"), want: `components[1].folder: duplicate "api"`},
		{name: "summary pipe", yaml: mustReplace(t, base, "summary: Frosthaven party, character and session companion (web + Discord)", "summary: a | b"), want: "summary: must be one line of 1–120 characters without |"},
		{name: "summary empty", yaml: mustReplace(t, base, "summary: Frosthaven party, character and session companion (web + Discord)", `summary: ""`), want: "summary: must be one line of 1–120 characters without |"},
		{name: "no components", yaml: noComponents, want: "components: at least one component is required"},
		{name: "harness version", yaml: mustReplace(t, base, "version: 1.0.0", "version: 1.0"), want: "harness.version: must be semver"},
		{name: "since", yaml: mustReplace(t, base, "overrides: []", "overrides:\n    - item: a\n      source: b\n      reason: c\n      since: 2/11/2026"), want: "harness.overrides[0].since: must be YYYY-MM-DD"},
		{name: "since impossible date", yaml: mustReplace(t, base, "overrides: []", "overrides:\n    - item: a\n      source: b\n      reason: c\n      since: 2026-13-45"), want: "harness.overrides[0].since: must be YYYY-MM-DD"},
		{name: "override empty", yaml: mustReplace(t, base, "overrides: []", "overrides:\n    - item: \"\"\n      source: b\n      reason: c\n      since: 2026-10-17\n      upstream: http://x"), want: "harness.overrides[0].item: must not be empty"},
		{name: "override upstream", yaml: mustReplace(t, base, "overrides: []", "overrides:\n    - item: a\n      source: b\n      reason: c\n      since: 2026-10-17\n      upstream: http://x"), want: "harness.overrides[0].upstream: must be an https URL"},
		{name: "ssot repo", yaml: mustReplace(t, base, "repo: https://github.com/althenlimzixuan/fhph-docs", "repo: github.com/o/r"), want: "ssot.repo: must be https://github.com/<owner>/<name>"},
		{name: "component repo", yaml: mustReplace(t, base, "repo: https://github.com/althenlimzixuan/fhph-api", "repo: git@github.com:o/r"), want: "components[0].repo: must be https://github.com/<owner>/<name>"},
		{name: "status", yaml: mustReplace(t, base, "status: active\nssot", "status: live\nssot"), want: "status: must be one of active, paused, dormant, archived"},
		{name: "component status", yaml: mustReplace(t, base, "stack: [go_api]\n    status: active", "stack: [go_api]\n    status: live"), want: "components[0].status: must be one of active, paused, dormant, archived"},
		{name: "stack item", yaml: mustReplace(t, base, "stack: [go_api]", "stack: [cobol]"), want: "components[0].stack[0]: must be one of go_api, nextjs_web, react_native, django_admin, tbd"},
		{name: "stack empty", yaml: mustReplace(t, base, "stack: [go_api]", "stack: []"), want: "components[0].stack: at least one item is required"},
		{name: "name", yaml: mustReplace(t, base, "name: FrostHaven Party Helper", `name: ""`), want: "name: must be one line of 1–60 characters"},
		{name: "board owner", yaml: mustReplace(t, base, "roadmap:  { owner: althenlimzixuan, number: 0 }", "roadmap:  { owner: \"\", number: 0 }"), want: "boards.roadmap.owner: must not be empty"},
		{name: "board number", yaml: mustReplace(t, base, "feedback: { owner: althenlimzixuan, number: 0 }", "feedback: { owner: althenlimzixuan, number: -1 }"), want: "boards.feedback.number: must be >= 0"},
		{name: "postman empty", yaml: mustReplace(t, base, "workspace: null", `workspace: ""`), want: "postman.workspace: must be null or a non-empty string"},
		{name: "slack prefix", yaml: mustReplace(t, base, "slack_prefix: fhph", "slack_prefix: Pal!"), want: "inbox.slack_prefix: must match ^[a-z0-9]{2,6}$"},
		{name: "slack prefix pal", yaml: mustReplace(t, base, "slack_prefix: fhph", "slack_prefix: pal"), want: `inbox.slack_prefix: must not be "pal"`},
		{name: "rank provider", yaml: mustReplace(t, base, "role: provider          # provider | platform | consumer\n    rank: 1", "role: provider\n    rank: 2"), want: "components[0].rank: must be 1 for role provider"},
	}
	override := func(item string) string {
		return mustReplace(t, base, "overrides: []", "overrides:\n    - item: "+item+"\n      source: b\n      reason: c\n      since: 2026-10-17")
	}
	for _, c := range []struct{ item, want string }{
		{`a\b`, "must use forward slashes"},
		{"/practices/x.md", "must be relative"},
		{"../x", `must not contain ".."`},
		{"practices/../x", `must not contain ".."`},
		{"practices//x.md", "must be a clean path"},
		{"./practices", "must be a clean path"},
		{"practices/", "must be a clean path"},
		{".", `must not be "."`},
		{".pal-harness/practices/x.md", `must not start with ".pal-harness/"`},
		{"VERSION", `must not be "VERSION"`},
	} {
		cases = append(cases, struct {
			name  string
			yaml  string
			want  string
			field string
		}{name: "override item " + c.item, yaml: override(c.item), want: "harness.overrides[0].item: " + c.want})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseProject([]byte(tc.yaml))
			if err == nil {
				t.Fatal("expected an error")
			}
			if tc.want == "" {
				if !strings.Contains(err.Error(), tc.field) {
					t.Fatalf("error %q does not name %q", err, tc.field)
				}
				return
			}
			if !hasProblem(err, tc.want) {
				t.Fatalf("missing problem %q in: %v", tc.want, err)
			}
		})
	}
}

func TestOverrideItemAccepted(t *testing.T) {
	base := readTestdata(t, "project.yaml")
	for _, item := range []string{"practices/workflow.md", "templates", "skills/phase", "practices/..x.md"} {
		y := mustReplace(t, base, "overrides: []", "overrides:\n    - item: "+item+"\n      source: b\n      reason: c\n      since: 2026-10-17")
		if _, err := ParseProject([]byte(y)); err != nil {
			t.Errorf("%s: %v", item, err)
		}
	}
}

func TestParseRejectsTrailingDocuments(t *testing.T) {
	const want = "manifest has more than one YAML document"
	for name, c := range map[string]struct {
		parse func([]byte) error
		file  string
	}{
		"project":   {func(b []byte) error { _, err := ParseProject(b); return err }, "project.yaml"},
		"workspace": {func(b []byte) error { _, err := ParseWorkspace(b); return err }, "workspace.yaml"},
		"pointer":   {func(b []byte) error { _, err := ParsePointer(b); return err }, "ssot.yaml"},
	} {
		t.Run(name, func(t *testing.T) {
			base := readTestdata(t, c.file)
			if err := c.parse([]byte(base)); err != nil {
				t.Fatalf("base: %v", err)
			}
			if err := c.parse([]byte(base + "---\nversion: 1\n")); err == nil || err.Error() != want {
				t.Fatalf("got %v, want %q", err, want)
			}
			if err := c.parse([]byte(base + "---\n[\n")); err == nil {
				t.Fatal("malformed trailing document should fail")
			}
		})
	}
}

func TestMarshalProjectKeepsMultilineReason(t *testing.T) {
	p, err := ParseProject([]byte(readTestdata(t, "project.yaml")))
	if err != nil {
		t.Fatal(err)
	}
	p.Harness.Overrides = []Override{{Item: "practices/workflow.md", Source: "1.0.0+patch", Reason: "first line\nx: {y}\n  backlog: {z}", Since: "2026-10-17"}}
	out, err := MarshalProject(p)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := ParseProject(out)
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !reflect.DeepEqual(p, p2) {
		t.Fatalf("round trip differs:\n%q\n%q\n%s", p.Harness.Overrides[0].Reason, p2.Harness.Overrides[0].Reason, out)
	}
	if !strings.Contains(string(out), "  backlog: { owner: althenlimzixuan, number: 0 }\n") {
		t.Fatalf("boards no longer spaced:\n%s", out)
	}
}

func TestParseProjectCollectsEveryProblem(t *testing.T) {
	y := mustReplace(t, readTestdata(t, "project.yaml"), "key: fhph", "key: X")
	y = mustReplace(t, y, "version: 1\nkey", "version: 3\nkey")
	_, err := ParseProject([]byte(y))
	var ve *ValidationError
	if !errors.As(err, &ve) || len(ve.Problems) != 2 {
		t.Fatalf("want 2 problems, got %v", err)
	}
	if !strings.Contains(err.Error(), "version: must be 1") || !strings.Contains(err.Error(), "key: must match") {
		t.Fatalf("Error() = %q", err.Error())
	}
}

func TestParseEmptyAndMalformed(t *testing.T) {
	if _, err := ParseProject(nil); err == nil {
		t.Error("empty project should fail")
	}
	if _, err := ParseWorkspace([]byte("")); err == nil {
		t.Error("empty workspace should fail")
	}
	if _, err := ParsePointer([]byte("version: [")); err == nil {
		t.Error("malformed pointer should fail")
	}
	if _, err := ParseWorkspace([]byte("version: 1\nbogus: 1\n")); err == nil {
		t.Error("unknown workspace field should fail")
	}
}

func TestParseWorkspace(t *testing.T) {
	w, err := ParseWorkspace([]byte(readTestdata(t, "workspace.yaml")))
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Projects) != 2 || w.Projects[0].Summary != "Group-trip itinerary planner" {
		t.Fatalf("unexpected: %+v", w)
	}
	out, err := MarshalWorkspace(w)
	if err != nil {
		t.Fatal(err)
	}
	w2, err := ParseWorkspace(out)
	if err != nil || !reflect.DeepEqual(w, w2) {
		t.Fatalf("round trip: %v\n%s", err, out)
	}

	base := readTestdata(t, "workspace.yaml")
	cases := []struct{ name, yaml, want string }{
		{"duplicate key", mustReplace(t, base, "key: fhph", "key: etp"), `projects[1].key: duplicate "etp"`},
		{"parent folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: ../x"), `projects[0].folder: must not contain ".."`},
		{"absolute folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: D:/x"), "projects[0].folder: must be relative"},
		{"backslash folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", `folder: a\b`), "projects[0].folder: must use forward slashes"},
		{"duplicate folder", mustReplace(t, base, "folder: FHPH-FrostHaven Party Helper", "folder: ETP-Essential Traveller Planner"), `projects[1].folder: duplicate "ETP-Essential Traveller Planner"`},
		{"bad version", mustReplace(t, base, "version: 1\n", "version: 2\n"), "version: must be 1"},
		{"bad status", mustReplace(t, base, "status: active ", "status: x "), "projects[0].status: must be one of active, paused, dormant, archived"},
		{"bad ssot", mustReplace(t, base, "ssot: https://github.com/althenlimzixuan/etp-docs", "ssot: nope"), "projects[0].ssot: must be https://github.com/<owner>/<name>"},
		{"bad summary", mustReplace(t, base, "summary: Group-trip itinerary planner", "summary: a|b"), "projects[0].summary: must be one line of 1–120 characters without |"},
		{"bad key", mustReplace(t, base, "key: etp", "key: E"), "projects[0].key: must match ^[a-z0-9]{2,6}$"},
		{"bad name", mustReplace(t, base, "name: Essential Traveller Planner", `name: ""`), "projects[0].name: must be one line of 1–60 characters"},
		{"dot folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: ."), `projects[0].folder: must not be "."`},
		{"unclean folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: a/./b"), "projects[0].folder: must be a clean path"},
		{"trailing slash folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: a/"), "projects[0].folder: must be a clean path"},
		{"double slash folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: a//b"), "projects[0].folder: must be a clean path"},
		{"pipe folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: a|b"), "projects[0].folder: must not contain |, ` or control characters"},
		{"backtick folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: a`b"), "projects[0].folder: must not contain |, ` or control characters"},
		{"control folder", mustReplace(t, base, "folder: ETP-Essential Traveller Planner", `folder: "a\tb"`), "projects[0].folder: must not contain |, ` or control characters"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseWorkspace([]byte(tc.yaml))
			if !hasProblem(err, tc.want) {
				t.Fatalf("missing %q in %v", tc.want, err)
			}
			var ve *ValidationError
			errors.As(err, &ve)
			if len(ve.Problems) != 1 {
				t.Fatalf("want exactly one problem, got %v", ve.Problems)
			}
		})
	}
	// summary is optional
	noSummary := strings.ReplaceAll(base, "    summary: Group-trip itinerary planner   # optional, one line\n", "")
	if _, err := ParseWorkspace([]byte(noSummary)); err != nil {
		t.Fatalf("summary should be optional: %v", err)
	}
	// an empty registry is valid
	if w, err := ParseWorkspace([]byte("version: 1\nprojects: []\n")); err != nil || len(w.Projects) != 0 {
		t.Fatalf("empty projects: %v %v", w, err)
	}
	// nested folders and spaces stay valid
	nested := mustReplace(t, base, "folder: ETP-Essential Traveller Planner", "folder: _IT-Internal-Tools/PC-Pal Center")
	if _, err := ParseWorkspace([]byte(nested)); err != nil {
		t.Fatalf("nested folder: %v", err)
	}
}

func TestParsePointer(t *testing.T) {
	base := readTestdata(t, "ssot.yaml")
	p, err := ParsePointer([]byte(base))
	if err != nil {
		t.Fatal(err)
	}
	if p.Project != "fhph" || p.Role != "consumer" || p.Rank != 2 {
		t.Fatalf("unexpected: %+v", p)
	}
	out, err := MarshalPointer(p)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := ParsePointer(out)
	if err != nil || !reflect.DeepEqual(p, p2) {
		t.Fatalf("round trip: %v\n%s", err, out)
	}
	cases := []struct{ name, yaml, want string }{
		{"rank", mustReplace(t, base, "rank: 2", "rank: 1"), "rank: must be 2 for role consumer"},
		{"role", mustReplace(t, base, "role: consumer", "role: boss"), "role: must be one of provider, platform, consumer"},
		{"version", mustReplace(t, base, "version: 1", "version: 9"), "version: must be 1"},
		{"project", mustReplace(t, base, "project: fhph", "project: F"), "project: must match ^[a-z0-9]{2,6}$"},
		{"ssot", mustReplace(t, base, "ssot: https://github.com/althenlimzixuan/fhph-docs", "ssot: http://x/y"), "ssot: must be https://github.com/<owner>/<name>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParsePointer([]byte(tc.yaml)); !hasProblem(err, tc.want) {
				t.Fatalf("missing %q in %v", tc.want, err)
			}
		})
	}
}

func TestRepoFullName(t *testing.T) {
	ok := map[string]string{
		"https://github.com/o/r":                       "o/r",
		"https://github.com/o/r.git":                   "o/r",
		"https://github.com/althenlimzixuan/fhph-docs": "althenlimzixuan/fhph-docs",
	}
	for in, want := range ok {
		got, err := RepoFullName(in)
		if err != nil || got != want {
			t.Errorf("RepoFullName(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{
		"http://github.com/o/r", "git@github.com:o/r", "https://github.com/o",
		"https://github.com/o/r/extra", "https://gitlab.com/o/r", "", "https://github.com/o/..",
	} {
		if got, err := RepoFullName(in); err == nil {
			t.Errorf("RepoFullName(%q) = %q, want error", in, got)
		}
	}
}

func TestRankFor(t *testing.T) {
	for role, want := range map[string]int{"provider": 1, "platform": 1, "consumer": 2, "x": 0} {
		if got := RankFor(role); got != want {
			t.Errorf("RankFor(%q) = %d, want %d", role, got, want)
		}
	}
}

func TestLegacySpecsBefore(t *testing.T) {
	base := readTestdata(t, "project.yaml")
	with := func(v string) string {
		return mustReplace(t, base, "overrides: []", "legacy_specs_before: "+v+"\n  overrides: []")
	}
	p, err := ParseProject([]byte(with(`"097"`)))
	if err != nil || p.Harness.LegacySpecsBefore != "097" {
		t.Fatalf("valid: %v %+v", err, p)
	}
	out, err := MarshalProject(p)
	if err != nil || !strings.Contains(string(out), "legacy_specs_before:") {
		t.Fatalf("marshal: %v\n%s", err, out)
	}
	q, err := ParseProject([]byte(base))
	if err != nil {
		t.Fatal(err)
	}
	out, err = MarshalProject(q)
	if err != nil || strings.Contains(string(out), "legacy_specs_before") {
		t.Fatalf("unset must be omitted: %v\n%s", err, out)
	}
	for _, v := range []string{`"97"`, `"0097"`, `"abc"`, `"09a"`} {
		_, err := ParseProject([]byte(with(v)))
		if !hasProblem(err, "harness.legacy_specs_before: must be three digits") {
			t.Errorf("%s: %v", v, err)
		}
	}
}

func TestFlowOverridesParse(t *testing.T) {
	base := readTestdata(t, "project.yaml")
	src := mustReplace(t, base, "overrides: []", "overrides: []\n  flow_overrides: [auto-promote, ci-test-summary]")
	p, err := ParseProject([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasFlowOverride("auto-promote") || !p.HasFlowOverride("ci-test-summary") {
		t.Errorf("overrides = %v", p.Harness.FlowOverrides)
	}
	if p.HasFlowOverride("no-owner-test") {
		t.Error("no-owner-test should be off")
	}
	out, err := MarshalProject(p)
	if err != nil || !strings.Contains(string(out), "flow_overrides:") {
		t.Fatalf("marshal: %v\n%s", err, out)
	}
	// No flow_overrides: the key is not written.
	q, err := ParseProject([]byte(base))
	if err != nil {
		t.Fatal(err)
	}
	out, err = MarshalProject(q)
	if err != nil || strings.Contains(string(out), "flow_overrides") {
		t.Fatalf("marshal without: %v\n%s", err, out)
	}
}

func TestFlowOverridesRejected(t *testing.T) {
	base := readTestdata(t, "project.yaml")
	tests := []struct{ list, want string }{
		{"[auto-merge]", "harness.flow_overrides[0]: must be one of auto-promote, no-owner-test, installer-ready, ci-test-summary"},
		{"[auto-promote, auto-promote]", "harness.flow_overrides[1]: duplicate"},
	}
	for _, tc := range tests {
		_, err := ParseProject([]byte(mustReplace(t, base, "overrides: []", "overrides: []\n  flow_overrides: "+tc.list)))
		if !hasProblem(err, tc.want) {
			t.Errorf("%s: want %q, got %v", tc.list, tc.want, err)
		}
	}
}

func TestSingleRepo(t *testing.T) {
	p, err := ParseProject([]byte(readTestdata(t, "project.yaml")))
	if err != nil {
		t.Fatal(err)
	}
	if p.SingleRepo() {
		t.Error("own repos: want false")
	}
	for i := range p.Components {
		p.Components[i].Repo = p.SSOT.Repo
	}
	if !p.SingleRepo() {
		t.Error("all SSOT repo: want true")
	}
	p.Components[1].Repo = "https://github.com/althenlimzixuan/fhph-web"
	if p.SingleRepo() {
		t.Error("mixed: want false")
	}
}
