package harness_test

import (
	"io/fs"
	"sort"
	"strings"
	"testing"

	harness "github.com/althenlimzixuan/pal-harness"
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
		"content/templates/spec.md",
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

	spec, _ := harness.Content.ReadFile("content/templates/spec.md")
	for _, h := range []string{"## User Scenarios", "## UX Outline", "## Requirements", "## Success Criteria", "## Key Entities", "## Assumptions"} {
		if !strings.Contains(string(spec), h) {
			t.Errorf("spec.md lacks heading %q", h)
		}
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
}
