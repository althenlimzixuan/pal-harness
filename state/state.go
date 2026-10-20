// Package state parses and validates the two files every project keeps for
// the phase flow: doc/features.yaml (the feature list) and
// specs/NNN-slug/phase-state.yaml (a phase's saved state).
//
// Findings are plain sentences with no file prefix; the caller adds it.
package state

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	FeatureStatuses = []string{"todo", "recommended", "picked", "brainstorming", "spec-review", "planning",
		"plan-review", "building", "code-review", "not-tested", "passed", "fixing", "deferred", "retest"}
	Stages = []string{"spec", "spec-review", "planning", "plan-review", "building", "code-review",
		"signoff", "staging-test", "shipping", "done"}
	WaitingKinds = []string{"spec-review", "plan-review", "questions", "findings", "signoff", "staging-test"}
)

type Feature struct {
	ID         string   `yaml:"id"`
	Name       string   `yaml:"name"`
	Label      string   `yaml:"label"`
	Why        string   `yaml:"why"`
	UseCase    string   `yaml:"use_case"`
	Status     string   `yaml:"status"`
	Phase      string   `yaml:"phase"`
	TestScript []Step   `yaml:"test_script"`
	Results    []Result `yaml:"results"`
}

type Step struct {
	Step   int    `yaml:"step"`
	Do     string `yaml:"do"`
	Expect string `yaml:"expect"`
}

type Result struct {
	Date   string `yaml:"date"`
	Env    string `yaml:"env"`
	Step   int    `yaml:"step"`
	Result string `yaml:"result"`
	Note   string `yaml:"note"`
	Issue  string `yaml:"issue"`
}

type PhaseState struct {
	Phase          string    `yaml:"phase"`
	Slug           string    `yaml:"slug"`
	Title          string    `yaml:"title"`
	Vehicle        string    `yaml:"vehicle"`
	Bundle         Bundle    `yaml:"bundle"`
	Stage          string    `yaml:"stage"`
	WaitingOnOwner *Waiting  `yaml:"waiting_on_owner"`
	Approach       string    `yaml:"approach"`
	Tasks          Tasks     `yaml:"tasks"`
	CI             string    `yaml:"ci"`
	LoadTest       bool      `yaml:"load_test"`
	Codex          Codex     `yaml:"codex"`
	Gates          Gates     `yaml:"gates"`
	LastProgressAt time.Time `yaml:"last_progress_at"`
	Notes          []string  `yaml:"notes"`
}

type Bundle struct {
	Features []string `yaml:"features"`
	Issues   []string `yaml:"issues"`
}

type Waiting struct {
	Kind      string    `yaml:"kind"`
	Since     time.Time `yaml:"since"`
	Link      string    `yaml:"link"`
	Questions int       `yaml:"questions"`
}

type Tasks struct {
	Done  int `yaml:"done"`
	Total int `yaml:"total"`
}

type Codex struct {
	Spec    int      `yaml:"spec"`
	Code    int      `yaml:"code"`
	UIUX    int      `yaml:"uiux"`
	Skipped []string `yaml:"skipped"`
}

type Gates struct {
	SpecApproved    *string `yaml:"spec_approved"`
	PlanApproved    *string `yaml:"plan_approved"`
	SignoffMerged   *string `yaml:"signoff_merged"`
	StagingTestDone *string `yaml:"staging_test_done"`
}

var (
	featureIDRe = regexp.MustCompile(`^F-\d{3,}$`)
	phaseRe     = regexp.MustCompile(`^\d{3}$`)
	issueRe     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9._-]+#\d+$`)
	dateRe      = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func oneOf(v string, set []string) bool {
	for _, s := range set {
		if v == s {
			return true
		}
	}
	return false
}

func notOneOf(what, v string, set []string) string {
	return fmt.Sprintf("%s %q is not one of %s", what, v, strings.Join(set, ", "))
}

func validDate(s string) bool {
	if !dateRe.MatchString(s) {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// decode reads exactly one YAML document strictly. A failure becomes one finding.
func decode(b []byte, out any) string {
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return "file is empty"
		}
		msg := strings.TrimPrefix(err.Error(), "yaml: unmarshal errors:")
		msg = strings.TrimPrefix(strings.TrimSpace(msg), "yaml: ")
		return "invalid YAML: " + strings.Join(strings.Fields(strings.ReplaceAll(msg, "\n", "; ")), " ")
	}
	var extra yaml.Node
	switch err := dec.Decode(&extra); {
	case errors.Is(err, io.EOF):
		return ""
	case err != nil:
		return "invalid YAML: " + err.Error()
	}
	return "invalid YAML: more than one YAML document"
}

// ParseFeatures decodes doc/features.yaml strictly and validates every entry.
func ParseFeatures(b []byte) ([]Feature, []string) {
	var fs []Feature
	if msg := decode(b, &fs); msg != "" {
		return nil, []string{msg}
	}
	var out []string
	add := func(id, format string, a ...any) { out = append(out, id+": "+fmt.Sprintf(format, a...)) }
	seen := map[string]bool{}
	for i, f := range fs {
		id := f.ID
		if id == "" {
			id = fmt.Sprintf("entry %d", i+1)
		}
		switch {
		case !featureIDRe.MatchString(f.ID):
			add(id, "id must look like F-012")
		case seen[f.ID]:
			add(id, "duplicate id")
		}
		seen[f.ID] = true
		if strings.TrimSpace(f.Name) == "" {
			add(id, "name is required")
		}
		if f.Label != "need" && f.Label != "might" {
			add(id, "label must be need or might")
		}
		if !oneOf(f.Status, FeatureStatuses) {
			add(id, "%s", notOneOf("status", f.Status, FeatureStatuses))
		}
		if f.Phase != "" && !phaseRe.MatchString(f.Phase) {
			add(id, "phase must be empty or three digits, like 096")
		}
		numbered := true
		steps := map[int]bool{}
		for j, s := range f.TestScript {
			steps[s.Step] = true
			if s.Step != j+1 {
				numbered = false
			}
		}
		if !numbered {
			add(id, "test_script steps must be numbered 1, 2, 3, …")
		}
		for j, s := range f.TestScript {
			if strings.TrimSpace(s.Do) == "" {
				add(id, "test_script step %d has no do", j+1)
			}
			if strings.TrimSpace(s.Expect) == "" {
				add(id, "test_script step %d has no expect", j+1)
			}
		}
		for j, r := range f.Results {
			n := j + 1
			if !validDate(r.Date) {
				add(id, "result %d: date must be YYYY-MM-DD", n)
			}
			if r.Env != "staging" && r.Env != "prod" {
				add(id, "result %d: env must be staging or prod", n)
			}
			if !steps[r.Step] {
				add(id, "result %d: step %d is not in the test script", n, r.Step)
			}
			if r.Result != "pass" && r.Result != "fail" {
				add(id, "result %d: result must be pass or fail", n)
			}
		}
	}
	return fs, out
}

// ParsePhaseState decodes phase-state.yaml strictly and validates it.
func ParsePhaseState(b []byte) (*PhaseState, []string) {
	var ps PhaseState
	if msg := decode(b, &ps); msg != "" {
		return nil, []string{msg}
	}
	var out []string
	add := func(format string, a ...any) { out = append(out, fmt.Sprintf(format, a...)) }
	if !phaseRe.MatchString(ps.Phase) {
		add("phase must be three digits, like 096")
	}
	if strings.TrimSpace(ps.Slug) == "" {
		add("slug is required")
	}
	if strings.TrimSpace(ps.Title) == "" {
		add("title is required")
	}
	if ps.Vehicle != "phase" && ps.Vehicle != "quick-fix" {
		add("vehicle must be phase or quick-fix")
	}
	if !oneOf(ps.Stage, Stages) {
		add("%s", notOneOf("stage", ps.Stage, Stages))
	}
	if ps.Approach != "inline" && ps.Approach != "subagent" {
		add("approach must be inline or subagent")
	}
	if !oneOf(ps.CI, []string{"green", "red", "running", "none"}) {
		add("ci must be green, red, running or none")
	}
	if w := ps.WaitingOnOwner; w != nil {
		if !oneOf(w.Kind, WaitingKinds) {
			add("%s", notOneOf("waiting_on_owner.kind", w.Kind, WaitingKinds))
		}
		if w.Since.IsZero() {
			add("waiting_on_owner.since is required")
		}
		if w.Questions < 0 {
			add("waiting_on_owner.questions must not be negative")
		}
	}
	if ps.Tasks.Done < 0 || ps.Tasks.Total < 0 {
		add("tasks.done and tasks.total must not be negative")
	} else if ps.Tasks.Done > ps.Tasks.Total {
		add("tasks.done is more than tasks.total")
	}
	if ps.LastProgressAt.IsZero() {
		add("last_progress_at is required")
	}
	for _, f := range ps.Bundle.Features {
		if !featureIDRe.MatchString(f) {
			add("bundle.features: %q must look like F-012", f)
		}
	}
	for _, is := range ps.Bundle.Issues {
		if !issueRe.MatchString(is) {
			add("bundle.issues: %q must look like owner/repo#431", is)
		}
	}
	for _, g := range []struct {
		name string
		v    *string
	}{
		{"spec_approved", ps.Gates.SpecApproved},
		{"plan_approved", ps.Gates.PlanApproved},
		{"signoff_merged", ps.Gates.SignoffMerged},
		{"staging_test_done", ps.Gates.StagingTestDone},
	} {
		if g.v != nil && !validDate(*g.v) {
			add("gates.%s must be a date (YYYY-MM-DD) or null", g.name)
		}
	}
	return &ps, out
}
