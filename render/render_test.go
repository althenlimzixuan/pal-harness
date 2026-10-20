package render

import (
	"io/fs"
	"strings"
	"testing"
)

func noDiagrams(rel string) ([]byte, error) { panic("unexpected diagram read: " + rel) }

func page(t *testing.T, md string, rd func(string) ([]byte, error)) string {
	t.Helper()
	b, err := Page([]byte(md), rd)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func wantContains(t *testing.T, got string, subs ...string) {
	t.Helper()
	for _, s := range subs {
		if !strings.Contains(got, s) {
			t.Errorf("missing %q in:\n%s", s, got)
		}
	}
}

func TestPageTitleAndViewport(t *testing.T) {
	got := page(t, "# 096 — Trip budget\n\nbody\n", noDiagrams)
	wantContains(t, got, "<title>096 — Trip budget</title>",
		`<meta name="viewport" content="width=device-width, initial-scale=1">`)
}

func TestPageDiagramLink(t *testing.T) {
	var asked string
	rd := func(rel string) ([]byte, error) { asked = rel; return []byte("graph TD; A-->B"), nil }
	got := page(t, "# T\n\n[Flow](diagrams/flow.mmd)\n", rd)
	if asked != "diagrams/flow.mmd" {
		t.Errorf("asked for %q", asked)
	}
	wantContains(t, got, `<pre class="mermaid">graph TD; A--&gt;B</pre>`, "<figcaption>Flow</figcaption>",
		`<script src="https://cdn.jsdelivr.net/npm/mermaid@11.4.1/dist/mermaid.min.js"></script>`,
		`<script>mermaid.initialize({startOnLoad: true});</script>`, "<figure>")
}

func TestPageFencedMermaid(t *testing.T) {
	got := page(t, "# T\n\n```mermaid\ngraph TD; A-->B\n```\n", noDiagrams)
	wantContains(t, got, `<pre class="mermaid">graph TD; A--&gt;B`, "mermaid.min.js")
}

func TestPageNoDiagramNoScript(t *testing.T) {
	got := page(t, "# T\n\nplain\n\n```go\nx := 1\n```\n", noDiagrams)
	if strings.Contains(got, "mermaid.min.js") {
		t.Error("script present without a diagram")
	}
	wantContains(t, got, "x := 1")
}

func TestPageAppendixCollapsed(t *testing.T) {
	got := page(t, "# T\n\nmain\n\n## Appendix\n\nextra detail\n", noDiagrams)
	wantContains(t, got, "<details><summary>Appendix</summary>", "</details>")
	i, j, k := strings.Index(got, "<details>"), strings.Index(got, "extra detail"), strings.Index(got, "</details>")
	if !(i < j && j < k) {
		t.Errorf("appendix text not inside details: %d %d %d", i, j, k)
	}
	if strings.Index(got, "main") > i {
		t.Error("main text should be before details")
	}
}

func TestPageEscapesRawHTML(t *testing.T) {
	got := page(t, "# T\n\n<script>alert(1)</script>\n\ninline <b>x</b>\n", noDiagrams)
	if strings.Contains(got, "<script>alert(1)") || strings.Contains(got, "<b>x") {
		t.Errorf("raw HTML passed through:\n%s", got)
	}
}

func TestPageGFM(t *testing.T) {
	got := page(t, "# T\n\n| a | b |\n|---|---|\n| 1 | 2 |\n\n~~gone~~\n\n- [x] done\n", noDiagrams)
	wantContains(t, got, "<table>", "<del>gone</del>", `type="checkbox"`)
}

func TestPageMissingDiagram(t *testing.T) {
	rd := func(rel string) ([]byte, error) { return nil, &fsErr{} }
	_, err := Page([]byte("# T\n\n[X](diagrams/none.mmd)\n"), rd)
	if err == nil || err.Error() != "diagrams/none.mmd: file not found" {
		t.Errorf("err = %v", err)
	}
}

type fsErr struct{}

func (*fsErr) Error() string { return "nope" }
func (*fsErr) Unwrap() error { return fs.ErrNotExist }
