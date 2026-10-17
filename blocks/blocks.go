// Package blocks renders the generated tables of the SSOT standard and swaps
// them into documents between their marker lines. The pal-harness commands
// and the Pal Center daemon both use it, so every writer produces identical text.
package blocks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/althenlimzixuan/pal-harness/manifest"
)

// Block names, as they appear in the marker lines.
const Components, Registry = "components", "registry"

var (
	ErrNoBlock  = errors.New("generated block not found")
	ErrBadBlock = errors.New("generated block markers are duplicated, unterminated or out of order")
)

// A block's text is the begin-marker line, header, separator, the rows, then
// the end-marker line. Lines are joined with LF and there is NO trailing
// newline after the end marker: the line ending that follows the end marker
// belongs to the surrounding document. Replace swaps the begin..end lines
// inclusive and leaves whatever followed the end line untouched.

func beginMarker(name, source string) string {
	return "<!-- pal:begin " + name + " (generated from " + source + "; do not edit) -->"
}

func endMarker(name string) string { return "<!-- pal:end " + name + " -->" }

func render(name, source string, lines []string) string {
	all := append([]string{beginMarker(name, source)}, lines...)
	all = append(all, endMarker(name))
	return strings.Join(all, "\n")
}

// ComponentsTable renders the components block for a project manifest.
func ComponentsTable(p *manifest.Project) string {
	lines := []string{"| Folder | Role | Rank | Stack | Repo | Status |", "|---|---|---|---|---|---|"}
	for _, c := range p.Components {
		lines = append(lines, fmt.Sprintf("| `%s/` | %s | %d | %s | %s | %s |",
			c.Folder, c.Role, c.Rank, strings.Join(c.Stack, ", "), c.Repo, c.Status))
	}
	return render(Components, "project.yaml", lines)
}

// RegistryTable renders the registry block for the workspace manifest. Folders
// use backslashes and a trailing backslash so the daemon's parsers keep working.
func RegistryTable(w *manifest.Workspace) string {
	lines := []string{"| Folder | Key | Summary | Status | SSOT |", "|---|---|---|---|---|"}
	for _, e := range w.Projects {
		summary := e.Summary
		if summary == "" {
			summary = "—"
		}
		folder := strings.ReplaceAll(e.Folder, "/", `\`) + `\`
		lines = append(lines, fmt.Sprintf("| `%s` | %s | %s | %s | %s |", folder, e.Key, summary, e.Status, e.SSOT))
	}
	return render(Registry, "workspace.yaml", lines)
}

// isBegin reports whether line is a begin marker for the name in prefix: the
// prefix must end the line or be followed by a space, "(" or "-->", so a
// marker for "components2" is not one for "components".
func isBegin(line, prefix string) bool {
	rest, ok := strings.CutPrefix(line, prefix)
	return ok && (rest == "" || rest[0] == ' ' || rest[0] == '(' || strings.HasPrefix(rest, "-->"))
}

// locate returns the byte range [start, end) of the block in doc: start is the
// beginning of the begin-marker line, end is the end of the end-marker line's
// content (before its CR/LF).
func locate(doc, name string) (int, int, error) {
	beginPrefix := "<!-- pal:begin " + name
	endLine := endMarker(name)
	var begins, ends []int // begin: line start; end: content end
	pos := 0
	for pos <= len(doc) {
		nl := strings.IndexByte(doc[pos:], '\n')
		lineEnd := len(doc)
		if nl >= 0 {
			lineEnd = pos + nl
		}
		content := strings.TrimSuffix(doc[pos:lineEnd], "\r")
		trimmed := strings.TrimSpace(content)
		switch {
		case isBegin(trimmed, beginPrefix):
			begins = append(begins, pos)
		case trimmed == endLine:
			ends = append(ends, pos+len(content))
		}
		if nl < 0 {
			break
		}
		pos = lineEnd + 1
	}
	switch {
	case len(begins) == 0 && len(ends) == 0:
		return 0, 0, fmt.Errorf("%w: %s", ErrNoBlock, name)
	case len(begins) != 1 || len(ends) != 1:
		return 0, 0, fmt.Errorf("%w: %s (%d begin, %d end markers)", ErrBadBlock, name, len(begins), len(ends))
	case ends[0] < begins[0]:
		return 0, 0, fmt.Errorf("%w: %s (end marker before begin marker)", ErrBadBlock, name)
	}
	return begins[0], ends[0], nil
}

// Replace swaps the named block in doc for block. If doc uses CRLF, the
// inserted lines do too. On any error it returns "" and the error.
func Replace(doc, name, block string) (string, error) {
	start, end, err := locate(doc, name)
	if err != nil {
		return "", err
	}
	block = strings.ReplaceAll(block, "\r\n", "\n")
	if strings.Contains(doc, "\r\n") {
		block = strings.ReplaceAll(block, "\n", "\r\n")
	}
	return doc[:start] + block + doc[end:], nil
}

// Extract returns the named block, markers included, with LF line endings.
func Extract(doc, name string) (string, error) {
	start, end, err := locate(doc, name)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(doc[start:end], "\r\n", "\n"), nil
}

// Matches reports whether the block in doc equals block, ignoring line endings.
func Matches(doc, name, block string) (bool, error) {
	got, err := Extract(doc, name)
	if err != nil {
		return false, err
	}
	return got == strings.ReplaceAll(block, "\r\n", "\n"), nil
}
