package harness_test

import (
	"os"
	"strings"
	"testing"

	harness "github.com/althenlimzixuan/pal-harness"
	"gopkg.in/yaml.v3"
)

type step struct {
	Uses  string            `yaml:"uses"`
	Run   string            `yaml:"run"`
	Shell string            `yaml:"shell"`
	With  map[string]any    `yaml:"with"`
	Env   map[string]string `yaml:"env"`
}

type workflow struct {
	On          map[string]any    `yaml:"on"`
	Permissions map[string]string `yaml:"permissions"`
	Jobs        map[string]struct {
		RunsOn string `yaml:"runs-on"`
		Steps  []step `yaml:"steps"`
	} `yaml:"jobs"`
}

func loadWorkflow(t *testing.T, name string) workflow {
	t.Helper()
	b, err := os.ReadFile(".github/workflows/" + name)
	if err != nil {
		t.Fatal(err)
	}
	var w workflow
	if err := yaml.Unmarshal(b, &w); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return w
}

func TestCheckWorkflowHardened(t *testing.T) {
	w := loadWorkflow(t, "check.yml")
	if w.Permissions["contents"] != "read" {
		t.Errorf("permissions = %v, want contents: read", w.Permissions)
	}
	steps := w.Jobs["check"].Steps
	if len(steps) == 0 {
		t.Fatal("no check job steps")
	}
	validated := false
	for i, s := range steps {
		if strings.Contains(s.Run, "${{") {
			t.Errorf("step %d interpolates an expression into run; pass it via env", i)
		}
		if strings.HasPrefix(s.Uses, "actions/setup-go@") && s.With["cache"] != false {
			t.Errorf("setup-go must set cache: false, with = %v", s.With)
		}
		if s.Env["VERSION"] == "${{ inputs.version }}" && strings.Contains(s.Run, `^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$`) {
			validated = true
		}
		if strings.Contains(s.Run, "pal-harness@v") && !validated {
			t.Errorf("step %d runs pal-harness before the version is validated", i)
		}
		if strings.Contains(s.Run, "pal-harness@v") && s.Env["VERSION"] != "${{ inputs.version }}" {
			t.Errorf("step %d must take the version from env", i)
		}
	}
	if !validated {
		t.Error("no step validates inputs.version")
	}
}

func TestCIWorkflowHasWindowsJob(t *testing.T) {
	w := loadWorkflow(t, "ci.yml")
	found := false
	for name, j := range w.Jobs {
		if j.RunsOn != "windows-latest" {
			continue
		}
		found = true
		var runs []string
		checkout := -1
		for i, s := range j.Steps {
			if strings.HasPrefix(s.Uses, "actions/checkout@") && checkout < 0 {
				checkout = i
			}
			runs = append(runs, s.Run)
		}
		if checkout < 1 || j.Steps[0].Run != "git config --global core.autocrlf true" {
			t.Errorf("%s: core.autocrlf must be set before checkout", name)
		}
		all := strings.Join(runs, "\n")
		if !strings.Contains(all, "go vet ./...") || !strings.Contains(all, "go test ./...") {
			t.Errorf("%s: must run vet and tests:\n%s", name, all)
		}
	}
	if !found {
		t.Fatal("no windows-latest job")
	}
}

func TestReleaseGuardWorkflow(t *testing.T) {
	w := loadWorkflow(t, "release-guard.yml")
	push, _ := w.On["push"].(map[string]any)
	tags, _ := push["tags"].([]any)
	if len(tags) != 1 || tags[0] != "v*" {
		t.Errorf("on.push.tags = %v, want [v*]", tags)
	}
	if len(push) != 1 {
		t.Errorf("on.push must only filter tags: %v", push)
	}
	var uses, runs []string
	for _, j := range w.Jobs {
		for _, s := range j.Steps {
			uses = append(uses, s.Uses)
			runs = append(runs, s.Run)
			if strings.HasPrefix(s.Uses, "actions/setup-go@") && s.With["go-version"] != "1.26" {
				t.Errorf("setup-go version = %v, want 1.26", s.With["go-version"])
			}
		}
	}
	u, r := strings.Join(uses, "\n"), strings.Join(runs, "\n")
	if !strings.Contains(u, "actions/checkout@") || !strings.Contains(u, "actions/setup-go@") {
		t.Errorf("must check out and set up Go: %v", uses)
	}
	if !strings.Contains(r, "GITHUB_REF_NAME") || !strings.Contains(r, "go run ./cmd/pal-harness version") {
		t.Errorf("must compare GITHUB_REF_NAME with the version command:\n%s", r)
	}
}

// loadYAML reads a workflow file as a plain map, for jobs that call reusable workflows.
func loadYAML(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(b, &m); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return m
}

func mapAt(m map[string]any, keys ...string) map[string]any {
	for _, k := range keys {
		m, _ = m[k].(map[string]any)
	}
	return m
}

func strAt(m map[string]any, keys ...string) string {
	last := keys[len(keys)-1]
	s, _ := mapAt(m, keys[:len(keys)-1]...)[last].(string)
	return s
}

func needsOf(job map[string]any) []string {
	switch n := job["needs"].(type) {
	case string:
		return []string{n}
	case []any:
		var out []string
		for _, v := range n {
			s, _ := v.(string)
			out = append(out, s)
		}
		return out
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestCIGateWorkflowsHardened(t *testing.T) {
	for _, c := range []struct {
		file, perm string
		inputs     []string
		outputs    []string
		runHas     []string
	}{
		{"ci-gate.yml", "read", []string{"version", "jobs", "shared"}, []string{"decisions", "fingerprints"},
			[]string{"fingerprint", "--with .github/workflows/ci.yml", "refs/ci-passed/", "GITHUB_STEP_SUMMARY"}},
		{"ci-mark.yml", "write", []string{"fingerprints", "results"}, nil,
			[]string{"refs/ci-passed/", "::warning::"}},
	} {
		w := loadYAML(t, ".github/workflows/"+c.file)
		call := mapAt(w, "on", "workflow_call")
		if call == nil {
			t.Errorf("%s: not a workflow_call workflow", c.file)
			continue
		}
		for _, in := range c.inputs {
			if mapAt(call, "inputs", in) == nil {
				t.Errorf("%s: no input %q", c.file, in)
			}
		}
		for _, out := range c.outputs {
			if mapAt(call, "outputs", out) == nil {
				t.Errorf("%s: no output %q", c.file, out)
			}
		}
		if got := strAt(w, "permissions", "contents"); got != c.perm {
			t.Errorf("%s: permissions.contents = %q, want %q", c.file, got, c.perm)
		}
		var runs []string
		validated := false
		for name, j := range mapAt(w, "jobs") {
			steps, _ := j.(map[string]any)["steps"].([]any)
			for i, st := range steps {
				s, _ := st.(map[string]any)
				run, _ := s["run"].(string)
				uses, _ := s["uses"].(string)
				runs = append(runs, run)
				if strings.Contains(run, "${{") {
					t.Errorf("%s %s step %d interpolates an expression into run; pass it via env", c.file, name, i)
				}
				if strings.HasPrefix(uses, "actions/setup-go@") && mapAt(s, "with")["cache"] != false {
					t.Errorf("%s: setup-go must set cache: false", c.file)
				}
				if strings.Contains(run, `^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$`) {
					validated = true
				}
				if strings.Contains(run, "pal-harness@v") {
					if !validated {
						t.Errorf("%s step %d runs pal-harness before the version is validated", c.file, i)
					}
					if strAt(s, "env", "VERSION") != "${{ inputs.version }}" {
						t.Errorf("%s step %d must take the version from env", c.file, i)
					}
				}
			}
		}
		all := strings.Join(runs, "\n")
		for _, sub := range c.runHas {
			if !strings.Contains(all, sub) {
				t.Errorf("%s: no step runs %q", c.file, sub)
			}
		}
	}
}

func TestCITemplateWiring(t *testing.T) {
	w := loadYAML(t, "content/templates/ci.yml")
	if got := strAt(w, "concurrency", "group"); got != "ci-${{ github.event.pull_request.number || github.ref }}" {
		t.Errorf("concurrency.group = %q", got)
	}
	if mapAt(w, "concurrency")["cancel-in-progress"] != true {
		t.Error("concurrency must set cancel-in-progress: true")
	}
	on := mapAt(w, "on")
	for _, ev := range []string{"pull_request", "push", "workflow_dispatch"} {
		if _, ok := on[ev]; !ok {
			t.Errorf("on lacks %s", ev)
		}
	}
	if b, _ := mapAt(on, "push")["branches"].([]any); len(b) != 3 {
		t.Errorf("on.push.branches = %v, want dev, staging, main", b)
	}
	jobs := mapAt(w, "jobs")
	gate, mark := mapAt(jobs, "gate"), mapAt(jobs, "mark")
	base := "althenlimzixuan/pal-harness/.github/workflows/"
	if got := strAt(gate, "uses"); got != base+"ci-gate.yml@v"+harness.Version {
		t.Errorf("gate.uses = %q", got)
	}
	if got := strAt(gate, "with", "version"); got != harness.Version {
		t.Errorf("gate.with.version = %q, want %q", got, harness.Version)
	}
	if strAt(gate, "with", "jobs") == "" {
		t.Error("gate.with.jobs is empty")
	}
	if got := strAt(mark, "uses"); got != base+"ci-mark.yml@v"+harness.Version {
		t.Errorf("mark.uses = %q", got)
	}
	if got := strAt(mark, "with", "results"); got != "${{ toJSON(needs) }}" {
		t.Errorf("mark.with.results = %q", got)
	}
	if got := strAt(mark, "with", "fingerprints"); got != "${{ needs.gate.outputs.fingerprints }}" {
		t.Errorf("mark.with.fingerprints = %q", got)
	}
	if strAt(mark, "permissions", "contents") != "write" {
		t.Error("mark must have contents: write")
	}
	if !strings.Contains(strAt(mark, "if"), "always()") {
		t.Errorf("mark.if = %q, want always()", strAt(mark, "if"))
	}
	work := 0
	for name, j := range jobs {
		if name == "gate" || name == "mark" {
			continue
		}
		work++
		job, _ := j.(map[string]any)
		if !contains(needsOf(job), "gate") {
			t.Errorf("%s must need gate", name)
		}
		cond := strAt(job, "if")
		if !strings.Contains(cond, "!cancelled()") || !strings.Contains(cond, "needs.gate.outputs.decisions || '{}'") ||
			!strings.Contains(cond, "."+name+" != 'skip'") {
			t.Errorf("%s.if = %q", name, cond)
		}
		if !contains(needsOf(mark), name) {
			t.Errorf("mark must need %s", name)
		}
	}
	if work == 0 {
		t.Error("the template has no example job")
	}
}
