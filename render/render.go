// Package render turns a Markdown spec and its Mermaid diagrams into one
// self-contained HTML page that reads comfortably on a phone.
package render

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const mermaidScript = `<script src="https://cdn.jsdelivr.net/npm/mermaid@11.4.1/dist/mermaid.min.js"></script>
<script>mermaid.initialize({startOnLoad: true, theme: window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'default'});</script>
`

const css = `:root{color-scheme:light dark;--bg:#fff;--fg:#1f2328;--muted:#59636e;--border:#d1d9e0;--code:#f6f8fa;--link:#0969da}
@media (prefers-color-scheme: dark){:root{--bg:#0d1117;--fg:#e6edf3;--muted:#9198a1;--border:#3d444d;--code:#151b23;--link:#4493f8}}
*{box-sizing:border-box}
body{margin:0 auto;max-width:42rem;padding:1rem;background:var(--bg);color:var(--fg);font:1rem/1.6 system-ui,-apple-system,"Segoe UI",sans-serif;overflow-wrap:break-word}
a{color:var(--link)}
h1,h2,h3{line-height:1.25}
code,pre{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em}
pre{background:var(--code);padding:.75rem;border-radius:6px;overflow-x:auto}
code{background:var(--code);padding:.1em .3em;border-radius:4px}
pre code{background:none;padding:0}
table{display:block;overflow-x:auto;border-collapse:collapse;max-width:100%}
th,td{border:1px solid var(--border);padding:.35rem .6rem}
blockquote{margin:1rem 0;padding:0 1rem;border-left:4px solid var(--border);color:var(--muted)}
figure{margin:1rem 0;overflow-x:auto}
figure pre.mermaid{background:none;text-align:center}
figcaption{color:var(--muted);font-size:.9em;text-align:center}
details{margin:1rem 0;border-top:1px solid var(--border)}
summary{cursor:pointer;font-weight:600;font-size:1.25rem;padding:.5rem 0}
img{max-width:100%}
`

// state is one render's bookkeeping, shared by the node renderers.
type state struct {
	readDiagram func(string) ([]byte, error)
	hasDiagram  bool
	inAppendix  bool
	err         error
}

func (s *state) fail(rel string, err error) {
	if s.err != nil {
		return
	}
	if errors.Is(err, fs.ErrNotExist) {
		s.err = fmt.Errorf("%s: file not found", rel)
		return
	}
	s.err = fmt.Errorf("%s: %w", rel, err)
}

func (s *state) mermaid(w util.BufWriter, body, caption string) {
	s.hasDiagram = true
	if caption != "" {
		w.WriteString("<figure>")
	}
	w.WriteString(`<pre class="mermaid">`)
	w.WriteString(html.EscapeString(strings.TrimRight(body, "\n")))
	w.WriteString("</pre>")
	if caption != "" {
		w.WriteString("<figcaption>" + html.EscapeString(caption) + "</figcaption></figure>")
	}
	w.WriteString("\n")
}

// diagramLink returns the link of a paragraph whose only content is a link to
// a .mmd file.
func diagramLink(p ast.Node) *ast.Link {
	l, ok := p.FirstChild().(*ast.Link)
	if !ok || p.FirstChild() != p.LastChild() {
		return nil
	}
	if !strings.HasSuffix(strings.ToLower(string(l.Destination)), ".mmd") {
		return nil
	}
	return l
}

func (s *state) paragraph(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if l := diagramLink(n); l != nil {
		if !entering {
			return ast.WalkContinue, nil
		}
		rel := string(l.Destination)
		body, err := s.readDiagram(rel)
		if err != nil {
			s.fail(rel, err)
			return ast.WalkSkipChildren, nil
		}
		s.mermaid(w, string(body), nodeText(src, l))
		return ast.WalkSkipChildren, nil
	}
	if entering {
		w.WriteString("<p>")
	} else {
		w.WriteString("</p>\n")
	}
	return ast.WalkContinue, nil
}

func (s *state) fenced(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	fb := n.(*ast.FencedCodeBlock)
	var body strings.Builder
	for i := 0; i < fb.Lines().Len(); i++ {
		seg := fb.Lines().At(i)
		body.Write(seg.Value(src))
	}
	lang := string(fb.Language(src))
	if lang == "mermaid" {
		s.mermaid(w, body.String(), "")
		return ast.WalkSkipChildren, nil
	}
	if lang != "" {
		w.WriteString(`<pre><code class="language-` + html.EscapeString(lang) + `">`)
	} else {
		w.WriteString("<pre><code>")
	}
	w.WriteString(html.EscapeString(body.String()))
	w.WriteString("</code></pre>\n")
	return ast.WalkSkipChildren, nil
}

// heading opens the collapsed appendix at "## Appendix"; the summary replaces
// the heading.
func (s *state) heading(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	h := n.(*ast.Heading)
	if h.Level == 2 && !s.inAppendix && nodeText(src, h) == "Appendix" {
		if entering {
			s.inAppendix = true
			w.WriteString("<details><summary>Appendix</summary>\n")
		}
		return ast.WalkSkipChildren, nil
	}
	tag := fmt.Sprintf("h%d", h.Level)
	if entering {
		w.WriteString("<" + tag + ">")
	} else {
		w.WriteString("</" + tag + ">\n")
	}
	return ast.WalkContinue, nil
}

func (s *state) document(w util.BufWriter, _ []byte, _ ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering && s.inAppendix {
		w.WriteString("</details>\n")
	}
	return ast.WalkContinue, nil
}

func (s *state) RegisterFuncs(r renderer.NodeRendererFuncRegisterer) {
	r.Register(ast.KindParagraph, s.paragraph)
	r.Register(ast.KindFencedCodeBlock, s.fenced)
	r.Register(ast.KindHeading, s.heading)
	r.Register(ast.KindDocument, s.document)
}

// nodeText is the plain text under n.
func nodeText(src []byte, n ast.Node) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if t, ok := c.(*ast.Text); ok && entering {
			b.Write(t.Segment.Value(src))
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// Page renders md to a complete HTML page. readDiagram receives each linked
// .mmd path exactly as written; an error wrapping fs.ErrNotExist makes Page
// fail with "<rel>: file not found". Raw HTML in md is not passed through.
func Page(md []byte, readDiagram func(rel string) ([]byte, error)) ([]byte, error) {
	s := &state{readDiagram: readDiagram}
	gm := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(renderer.WithNodeRenderers(util.Prioritized(s, 100))),
	)
	doc := gm.Parser().Parse(text.NewReader(md))
	title := "Document"
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		if h, ok := c.(*ast.Heading); ok && h.Level == 1 {
			title = nodeText(md, h)
			break
		}
	}
	var body bytes.Buffer
	if err := gm.Renderer().Render(&body, md, doc); err != nil {
		return nil, err
	}
	if s.err != nil {
		return nil, s.err
	}
	var out bytes.Buffer
	out.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n<meta charset=\"utf-8\">\n")
	out.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">` + "\n")
	out.WriteString("<title>" + html.EscapeString(title) + "</title>\n<style>\n" + css + "</style>\n</head>\n<body>\n")
	out.Write(body.Bytes())
	if s.hasDiagram {
		out.WriteString(mermaidScript)
	}
	out.WriteString("</body>\n</html>\n")
	return out.Bytes(), nil
}
