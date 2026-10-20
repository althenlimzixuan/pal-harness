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
		"content/agents/retro.md",
		"content/agents/review-code.md",
		"content/agents/review-contract.md",
		"content/agents/review-errors.md",
		"content/agents/review-plan.md",
		"content/agents/review-security.md",
		"content/agents/review-tests.md",
		"content/agents/review-ui.md",
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
		"content/skills/phase/SKILL.md",
		"content/skills/phase/references/build.md",
		"content/skills/phase/references/plan.md",
		"content/skills/phase/references/review.md",
		"content/skills/phase/references/ship.md",
		"content/skills/phase/references/signoff.md",
		"content/skills/phase/references/spec.md",
		"content/skills/phase/references/staging-test.md",
		"content/skills/pm/SKILL.md",
		"content/skills/quick-fix/SKILL.md",
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

	// The phase flow practices (design §3-§11, §12.8).
	has("practices/workflow.md", ".pal-harness/skills/phase/SKILL.md", "before production", "[HUMAN GATE]",
		"spec", "plan", "sign-off", "staging test", "retro", "10 minutes", "Code patterns", "pal-review", "accept all")
	lacks("practices/workflow.md", "Four phases, each with a gate")
	has("practices/testing.md", "_test_reports/<component>/", "summary.md", "tree hash", "80%", "20%",
		"5 points", "include load test", "n/a")
	has("practices/ssot_standard.md", "doc/features.yaml", "doc/north-star.md", "doc/decisions.md",
		"doc/later.md", "_test_reports/", "phase-state.yaml")
	lacks("practices/ssot_standard.md", "superpowers writes its design and plan into that folder")
	has("practices/devops_cicd.md", "before production", "staging → main")

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

	// The pm skill: choosing the work (design §3, §4, §12.8.1/5/9).
	has("skills/pm/SKILL.md", "name: pm", "up to 6", "at least one roadmap feature", "bundle",
		"about one working day", "quick-fix", "owner-task", "doc/later.md", "every 5 phases",
		"doc/north-star.md", "doc/features.yaml", "Approved to run", "Recommended next", "Todo",
		"/pm pickup", "retest", "deferred", "seed")

	// The phase and quick-fix skills: running the work (design §5-§11, §12.3, §12.5, §12.8.2-4/7/8).
	has("skills/phase/SKILL.md", "name: phase",
		"references/spec.md", "references/plan.md", "references/build.md", "references/review.md",
		"references/signoff.md", "references/staging-test.md", "references/ship.md", "retro",
		"`spec`", "`spec-review`", "`planning`", "`plan-review`", "`building`", "`code-review`",
		"`signoff`", "`staging-test`", "`shipping`", "`done`",
		"phase-state.yaml", "last_progress_at", "waiting_on_owner", "10 minutes", "pal-harness render",
		"Remote Control", "PushNotification", "superpowers:writing-plans",
		"superpowers:subagent-driven-development", "not available", "_feedbacks/", "local stack",
		".pal-worktrees/<KEY>-<NNN>/", "session:", "docs/NNN-slug", "needs you", "pal-work")
	has("skills/phase/references/spec.md", "Calls I made", "recommended answer", "accept all", "skip Codex",
		"--kind doc", "doc/decisions.md", "test script", "contract", "split the phase")
	has("skills/phase/references/plan.md", "test file", "fixture", "pal-review-plan", "include load test",
		"Building starts only after the owner approves the plan")
	has("skills/phase/references/build.md", "quick fix", "up to ~4 tasks", "5 or more tasks", "feature/NNN-slug",
		"needs you", "stuck", "1 hour")
	has("skills/phase/references/review.md", "pal-review-code", "pal-review-tests", "pal-review-errors",
		"pal-review-security", "pal-review-contract", "pal-review-ui", "--kind code", "--kind uiux",
		"minor", "major", "critical", "3 rounds", "5 rounds", "doc/later.md")
	has("skills/phase/references/signoff.md", "[HUMAN GATE]", "templates/signoff-pr.md", "docs/NNN-slug",
		"dev → staging")
	has("skills/phase/references/staging-test.md", "fresh seed data", "fix now", "defer", "branches from staging",
		"feedback", "retest", "Expo Go")
	has("skills/phase/references/ship.md", "torn down", "48 hours", "smoke check", "retro", "recommended next",
		"harness capture")
	has("skills/quick-fix/SKILL.md", "name: quick-fix", "one review", "test step")

	// The agents (design §6.3, §8, §9.5, §11.3).
	for _, a := range []string{"review-code", "review-tests", "review-errors", "review-security",
		"review-contract", "review-ui", "review-plan"} {
		has("agents/"+a+".md", "none", "minor", "major", "critical", "specs/NNN-slug/reviews/")
	}
	has("agents/review-code.md", "TDD", "clean layers", "SOLID", "DRY")
	has("agents/review-tests.md", "80%", "fixture", "stale", "_test_reports/")
	has("agents/review-errors.md", "silent")
	has("agents/review-security.md", "secrets")
	has("agents/review-contract.md", "openapi.yaml", "API.md")
	has("agents/review-ui.md", "contrast", "screen-reader")
	has("agents/review-plan.md", "every spec feature", "code patterns", "contract", "one working day", "test file")
	has("agents/retro.md", "Went well", "Went badly", "harness capture", "_feedbacks/")

	// Pal Center 0.4.0 CLIs (pal-work, pal-capture, doc kind, worktree-set reviews).
	has("skills/phase/references/staging-test.md", "wait-results --phase NNN", "Monitor", "results --phase NNN",
		"Pal Center version:", "gh pr merge", "--merge")
	has("skills/phase/SKILL.md", "wait-results --phase NNN", "Monitor", "results --phase NNN",
		"Pal Center version:", "status --session", "older than 0.4.0", "exits 5", "gh pr merge --merge")
	has("skills/phase/references/spec.md", "--spec specs/NNN-slug")
	has("agents/retro.md", "Capture CLI:")
	has("skills/phase/references/ship.md", "Capture CLI:")
	has("skills/pm/SKILL.md", "Work CLI:")
	has("skills/quick-fix/SKILL.md", "gh pr merge --merge", "Pal Center version:")
	has("skills/phase/references/review.md", "gh pr merge", "--merge", "Pal Center version:")
	for _, p := range []string{"skills/phase/SKILL.md", "skills/phase/references/review.md",
		"skills/phase/references/spec.md", "skills/quick-fix/SKILL.md"} {
		s := text(p)
		const phrase = "not available (worktree set)"
		for i := 0; ; {
			j := strings.Index(s[i:], phrase)
			if j < 0 {
				break
			}
			j += i
			lo, hi := max(0, j-300), min(len(s), j+300)
			if !strings.Contains(s[lo:hi], "older than 0.4.0") {
				t.Errorf("%s: %q is not beside \"older than 0.4.0\"", p, phrase)
			}
			i = j + len(phrase)
		}
	}
	for _, p := range []string{"skills/phase/references/review.md", "skills/quick-fix/SKILL.md",
		"skills/phase/references/staging-test.md"} {
		if strings.Contains(text(p), "--squash") {
			t.Errorf("%s: session PRs merge with --merge, never --squash", p)
		}
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
