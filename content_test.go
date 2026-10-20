package harness_test

import (
	"io/fs"
	"path"
	"sort"
	"strings"
	"testing"

	harness "github.com/althenlimzixuan/pal-harness"
	"github.com/althenlimzixuan/pal-harness/state"
)

func TestContentComplete(t *testing.T) {
	want := []string{
		"content/practices/conflict_protocol.md",
		"content/practices/devops_cicd.md",
		"content/practices/django_admin.md",
		"content/practices/go_api.md",
		"content/practices/nextjs_web.md",
		"content/practices/product_docs.md",
		"content/practices/react_native.md",
		"content/practices/ssot_standard.md",
		"content/practices/testing.md",
		"content/practices/workflow.md",
		"content/templates/adr-conflict.md",
		"content/templates/adr.md",
		"content/templates/decisions.md",
		"content/templates/features.yaml",
		"content/templates/later.md",
		"content/templates/north-star.md",
		"content/templates/phase-state.yaml",
		"content/templates/plan.md",
		"content/templates/retro.md",
		"content/templates/signoff-pr.md",
		"content/templates/spec.md",
		"content/templates/test-summary.md",
	}
	var got []string
	err := fs.WalkDir(harness.Content, "content", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			got = append(got, p)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("content files:\n got %v\nwant %v", got, want)
	}

	for _, p := range got {
		b, _ := harness.Content.ReadFile(p)
		text := string(b)
		if strings.Contains(text, `C:\Users`) {
			t.Errorf("%s contains an absolute C:\\Users path", p)
		}
		if strings.HasPrefix(p, "content/practices/") {
			for _, bad := range []string{"Shape B", "single Service", "squash merge to `main`"} {
				if strings.Contains(text, bad) {
					t.Errorf("%s contains %q", p, bad)
				}
			}
		}
		for _, line := range strings.Split(text, "\n") {
			if strings.HasPrefix(line, "<!-- pal:begin") || strings.HasPrefix(line, "<!-- pal:end") {
				t.Errorf("%s has a line-leading block marker: %q", p, line)
			}
		}
	}

	lines := func(p string) int {
		b, _ := harness.Content.ReadFile(p)
		return strings.Count(strings.TrimRight(string(b), "\n"), "\n") + 1
	}
	if n := lines("content/practices/ssot_standard.md"); n > 150 {
		t.Errorf("ssot_standard.md has %d lines, max 150", n)
	}
	if n := lines("content/practices/conflict_protocol.md"); n > 60 {
		t.Errorf("conflict_protocol.md has %d lines, max 60", n)
	}

	text := func(p string) string {
		b, err := harness.Content.ReadFile("content/" + p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	has := func(p string, subs ...string) {
		t.Helper()
		for _, s := range subs {
			if !strings.Contains(text(p), s) {
				t.Errorf("%s lacks %q", p, s)
			}
		}
	}
	lacks := func(p string, subs ...string) {
		t.Helper()
		for _, s := range subs {
			if strings.Contains(text(p), s) {
				t.Errorf("%s still contains %q", p, s)
			}
		}
	}

	// R26: the contract-sync lines point to product_docs' split rule; no
	// practice tells a component to carry doc/API.md.
	for _, p := range []string{"practices/workflow.md", "practices/go_api.md"} {
		has(p, "split rule in `product_docs.md`", "the SSOT repo's PR updates its `doc/openapi.yaml` copy and `doc/API.md`")
		lacks(p, "regenerate `openapi.yaml`, update `doc/API.md`")
	}
	has("practices/ssot_standard.md",
		"**Deliberately absent from the SSOT repo:**",
		"never edit `.pal-harness/` by hand",
		"`harness.overrides` entry (item, source, reason, since)",
		"only by a PR to `pal-harness`",
		"The §4.5 copy drift check is manual until a later harness release",
		"Regenerating it is manual until a later harness release")
	has("practices/product_docs.md",
		"The §4.5 copy drift check is manual until a later harness release",
		"Regenerating it is manual until a later harness release")

	// §2.1: feature PRs into dev may be squash-merged; it is not a rule.
	for _, p := range []string{"practices/workflow.md", "practices/devops_cicd.md"} {
		has(p, "may be squash-merged")
		lacks(p, "squash-merge feature PRs", "squash-merge into")
	}

	has("templates/adr-conflict.md", "- **Status:** Accepted\n")

	// The phase flow templates (design §5, §6, §10.1, §11.3, §12.8.6).
	has("templates/spec.md", "## Summary", "## Questions for the owner", "## Calls I made", "## Features",
		"## UX outline", "## Contract", "## Out of scope", "## Appendix",
		"Delete this section when the phase does not change the API.",
		"**Why:**", "**Use case:**", "**Test script:**")
	if sp := text("templates/spec.md"); strings.Index(sp, "Delete this section when the phase does not change the API.") < strings.Index(sp, "## Contract") {
		t.Errorf("spec.md: the delete line must sit under ## Contract")
	}
	has("templates/plan.md", "## Summary", "## Build approach", "## Reports", "## Tasks", "include load test")
	has("templates/signoff-pr.md", "[HUMAN GATE]", "## Summary", "## Calls I made", "## Review results",
		"## Test reports", "## When you merge", "## Your test scripts")
	has("templates/retro.md", "## Went well", "## Went badly", "## Ideas")
	has("templates/test-summary.md", "1.", "coverage", "p95")

	if _, f := state.ParsePhaseState([]byte(text("templates/phase-state.yaml"))); len(f) != 0 {
		t.Errorf("templates/phase-state.yaml findings: %v", f)
	}
	if _, f := state.ParseFeatures([]byte(text("templates/features.yaml"))); len(f) != 0 {
		t.Errorf("templates/features.yaml findings: %v", f)
	}

	// Size caps (anti-goal: rule growth) and front matter for skills and agents.
	front := func(p string) map[string]string {
		m := map[string]string{}
		ls := strings.Split(text(p), "\n")
		if ls[0] != "---" {
			t.Errorf("%s has no front matter", p)
			return m
		}
		for _, l := range ls[1:] {
			if l == "---" {
				return m
			}
			if k, v, ok := strings.Cut(l, ":"); ok {
				m[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
		t.Errorf("%s front matter is not closed", p)
		return m
	}
	for _, g := range []struct {
		pattern string
		max     int
	}{
		{"content/skills/*/SKILL.md", 250},
		{"content/skills/*/references/*.md", 150},
		{"content/agents/*.md", 120},
	} {
		matches, err := fs.Glob(harness.Content, g.pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range matches {
			if n := lines(p); n > g.max {
				t.Errorf("%s has %d lines, max %d", p, n, g.max)
			}
		}
	}
	skills, _ := fs.Glob(harness.Content, "content/skills/*/SKILL.md")
	for _, p := range skills {
		fm := front(strings.TrimPrefix(p, "content/"))
		if want := path.Base(path.Dir(p)); fm["name"] != want {
			t.Errorf("%s name = %q, want %q", p, fm["name"], want)
		}
		if fm["description"] == "" {
			t.Errorf("%s has no description", p)
		}
	}
	agents, _ := fs.Glob(harness.Content, "content/agents/*.md")
	for _, p := range agents {
		fm := front(strings.TrimPrefix(p, "content/"))
		if want := "pal-" + strings.TrimSuffix(path.Base(p), ".md"); fm["name"] != want {
			t.Errorf("%s name = %q, want %q", p, fm["name"], want)
		}
		if fm["description"] == "" || fm["model"] == "" {
			t.Errorf("%s needs a description and a model", p)
		}
		if !strings.Contains(fm["tools"], "Read") {
			t.Errorf("%s tools = %q, must include Read", p, fm["tools"])
		}
	}
}
