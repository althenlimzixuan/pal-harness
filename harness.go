// Package harness is the single implementation of the AI Workshop shared
// harness: project manifests, generated tables, the pinned-copy writer and
// checker, numbering commands, and the embedded practices/templates content.
package harness

import "embed"

// Version is the harness release. It must equal the version in
// .claude-plugin/plugin.json and the first heading of CHANGELOG.md.
const Version = "1.3.0"

// Content holds the practices and templates shipped in every pinned copy.
// Paths start with "content/". Dot-files are deliberately not embedded.
//
//go:embed content
var Content embed.FS
