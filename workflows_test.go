package harness_test

import (
	"os"
	"strings"
	"testing"

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
