// Package manifest is the single implementation of the three strict file
// formats of the SSOT standard: project.yaml, workspace.yaml and ssot.yaml.
package manifest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// FormatVersion is the only supported value of every manifest's `version`.
const FormatVersion = 1

var (
	Roles    = []string{"provider", "platform", "consumer"}
	Stacks   = []string{"go_api", "nextjs_web", "react_native", "django_admin", "tbd"}
	Statuses = []string{"active", "paused", "dormant", "archived"}
)

// RankFor returns the rank implied by a role: provider and platform are 1,
// consumer is 2, anything else 0.
func RankFor(role string) int {
	switch role {
	case "provider", "platform":
		return 1
	case "consumer":
		return 2
	}
	return 0
}

type Project struct {
	Version    int         `yaml:"version"`
	Key        string      `yaml:"key"`
	Name       string      `yaml:"name"`
	Summary    string      `yaml:"summary"`
	Status     string      `yaml:"status"`
	SSOT       SSOT        `yaml:"ssot"`
	Components []Component `yaml:"components"`
	Boards     Boards      `yaml:"boards"`
	Postman    Postman     `yaml:"postman"`
	Inbox      Inbox       `yaml:"inbox"`
	Harness    Harness     `yaml:"harness"`
}

type SSOT struct {
	Repo string `yaml:"repo"`
}

type Component struct {
	Folder string   `yaml:"folder"`
	Repo   string   `yaml:"repo"`
	Role   string   `yaml:"role"`
	Rank   int      `yaml:"rank"`
	Stack  []string `yaml:"stack,flow"`
	Status string   `yaml:"status"`
}

type Boards struct {
	Backlog  BoardRef `yaml:"backlog,flow"`
	Roadmap  BoardRef `yaml:"roadmap,flow"`
	Feedback BoardRef `yaml:"feedback,flow"`
}

type BoardRef struct {
	Owner  string `yaml:"owner"`
	Number int    `yaml:"number"`
}

type Postman struct {
	Workspace *string `yaml:"workspace"`
}

type Inbox struct {
	SlackPrefix string `yaml:"slack_prefix"`
}

type Harness struct {
	Version   string     `yaml:"version"`
	Overrides []Override `yaml:"overrides"`
}

type Override struct {
	Item     string `yaml:"item"`
	Source   string `yaml:"source"`
	Reason   string `yaml:"reason"`
	Since    string `yaml:"since"`
	Upstream string `yaml:"upstream,omitempty"`
}

type Workspace struct {
	Version  int                `yaml:"version"`
	Projects []WorkspaceProject `yaml:"projects"`
}

type WorkspaceProject struct {
	Key     string `yaml:"key"`
	Name    string `yaml:"name"`
	Folder  string `yaml:"folder"`
	SSOT    string `yaml:"ssot"`
	Status  string `yaml:"status"`
	Summary string `yaml:"summary,omitempty"`
}

type Pointer struct {
	Version int    `yaml:"version"`
	Project string `yaml:"project"`
	SSOT    string `yaml:"ssot"`
	Role    string `yaml:"role"`
	Rank    int    `yaml:"rank"`
}

// ValidationError lists every problem found, each as "<path>: <message>".
type ValidationError struct{ Problems []string }

func (e *ValidationError) Error() string {
	return "invalid manifest: " + strings.Join(e.Problems, "; ")
}

// ---- rules ----

var (
	keyRe      = regexp.MustCompile(`^[a-z0-9]{2,6}$`)
	folderRe   = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,29}$`)
	repoRe     = regexp.MustCompile(`^https://github\.com/([A-Za-z0-9][A-Za-z0-9-]*)/([A-Za-z0-9._-]+?)(\.git)?$`)
	semverRe   = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)
	dateRe     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	driveRe    = regexp.MustCompile(`^[A-Za-z]:`)
	reserved   = []string{"doc", "specs", "context", "postman", "_feedbacks"}
	msgRepoURL = "must be https://github.com/<owner>/<name>"
)

// RepoFullName turns a GitHub repo URL into "owner/name". A trailing .git is accepted.
func RepoFullName(url string) (string, error) {
	m := repoRe.FindStringSubmatch(url)
	if m == nil || m[2] == "." || m[2] == ".." {
		return "", fmt.Errorf("%q is not a GitHub repo URL (%s)", url, msgRepoURL)
	}
	return m[1] + "/" + m[2], nil
}

type problems struct{ list []string }

func (p *problems) add(path, format string, a ...any) {
	p.list = append(p.list, path+": "+fmt.Sprintf(format, a...))
}

func (p *problems) err() error {
	if len(p.list) == 0 {
		return nil
	}
	return &ValidationError{Problems: p.list}
}

func oneOf(v string, set []string) bool {
	for _, s := range set {
		if v == s {
			return true
		}
	}
	return false
}

func oneLine(s string, max int) bool {
	n := utf8.RuneCountInString(s)
	return n >= 1 && n <= max && !strings.ContainsAny(s, "\r\n")
}

func (p *problems) version(v int) {
	if v != FormatVersion {
		p.add("version", "must be %d", FormatVersion)
	}
}

func (p *problems) key(path, v string) {
	if !keyRe.MatchString(v) {
		p.add(path, "must match ^[a-z0-9]{2,6}$")
	}
}

func (p *problems) name(path, v string) {
	if !oneLine(v, 60) {
		p.add(path, "must be one line of 1–60 characters")
	}
}

func (p *problems) summary(path, v string) {
	if !oneLine(v, 120) || strings.Contains(v, "|") {
		p.add(path, "must be one line of 1–120 characters without |")
	}
}

func (p *problems) status(path, v string) {
	if !oneOf(v, Statuses) {
		p.add(path, "must be one of %s", strings.Join(Statuses, ", "))
	}
}

func (p *problems) repo(path, v string) {
	if _, err := RepoFullName(v); err != nil {
		p.add(path, "%s", msgRepoURL)
	}
}

func (p *problems) roleRank(rolePath, rankPath, role string, rank int) {
	if !oneOf(role, Roles) {
		p.add(rolePath, "must be one of %s", strings.Join(Roles, ", "))
		return
	}
	if rank != RankFor(role) {
		p.add(rankPath, "must be %d for role %s", RankFor(role), role)
	}
}

// Validate reports every problem in the project manifest.
func (pr *Project) Validate() error {
	p := &problems{}
	p.version(pr.Version)
	p.key("key", pr.Key)
	p.name("name", pr.Name)
	p.summary("summary", pr.Summary)
	p.status("status", pr.Status)
	p.repo("ssot.repo", pr.SSOT.Repo)

	if len(pr.Components) == 0 {
		p.add("components", "at least one component is required")
	}
	seen := map[string]bool{}
	for i, c := range pr.Components {
		path := fmt.Sprintf("components[%d]", i)
		switch {
		case oneOf(c.Folder, reserved) || strings.HasPrefix(c.Folder, ".") || strings.HasPrefix(c.Folder, "_"):
			p.add(path+".folder", "%q is reserved", c.Folder)
		case !folderRe.MatchString(c.Folder):
			p.add(path+".folder", "must match ^[a-z][a-z0-9_-]{0,29}$")
		case seen[c.Folder]:
			p.add(path+".folder", "duplicate %q", c.Folder)
		}
		seen[c.Folder] = true
		p.repo(path+".repo", c.Repo)
		p.roleRank(path+".role", path+".rank", c.Role, c.Rank)
		if len(c.Stack) == 0 {
			p.add(path+".stack", "at least one item is required")
		}
		for j, s := range c.Stack {
			if !oneOf(s, Stacks) {
				p.add(fmt.Sprintf("%s.stack[%d]", path, j), "must be one of %s", strings.Join(Stacks, ", "))
			}
		}
		p.status(path+".status", c.Status)
	}

	for _, b := range []struct {
		name string
		ref  BoardRef
	}{{"backlog", pr.Boards.Backlog}, {"roadmap", pr.Boards.Roadmap}, {"feedback", pr.Boards.Feedback}} {
		path := "boards." + b.name
		if b.ref.Owner == "" {
			p.add(path+".owner", "must not be empty")
		}
		if b.ref.Number < 0 {
			p.add(path+".number", "must be >= 0")
		}
	}

	if pr.Postman.Workspace != nil && *pr.Postman.Workspace == "" {
		p.add("postman.workspace", "must be null or a non-empty string")
	}

	switch {
	case !keyRe.MatchString(pr.Inbox.SlackPrefix):
		p.add("inbox.slack_prefix", "must match ^[a-z0-9]{2,6}$")
	case pr.Inbox.SlackPrefix == "pal":
		p.add("inbox.slack_prefix", `must not be "pal"`)
	}

	if !semverRe.MatchString(pr.Harness.Version) {
		p.add("harness.version", "must be semver")
	}
	for i, o := range pr.Harness.Overrides {
		path := fmt.Sprintf("harness.overrides[%d]", i)
		for _, f := range []struct{ n, v string }{{"item", o.Item}, {"source", o.Source}, {"reason", o.Reason}} {
			if strings.TrimSpace(f.v) == "" {
				p.add(path+"."+f.n, "must not be empty")
			}
		}
		if _, err := time.Parse("2006-01-02", o.Since); err != nil || !dateRe.MatchString(o.Since) {
			p.add(path+".since", "must be YYYY-MM-DD")
		}
		if o.Upstream != "" && !strings.HasPrefix(o.Upstream, "https://") {
			p.add(path+".upstream", "must be an https URL")
		}
	}
	return p.err()
}

// Validate reports every problem in the workspace registry.
func (w *Workspace) Validate() error {
	p := &problems{}
	p.version(w.Version)
	keys, folders := map[string]bool{}, map[string]bool{}
	for i, e := range w.Projects {
		path := fmt.Sprintf("projects[%d]", i)
		p.key(path+".key", e.Key)
		if keys[e.Key] {
			p.add(path+".key", "duplicate %q", e.Key)
		}
		keys[e.Key] = true
		p.name(path+".name", e.Name)

		switch {
		case e.Folder == "":
			p.add(path+".folder", "must not be empty")
		case strings.HasPrefix(e.Folder, "/") || driveRe.MatchString(e.Folder):
			p.add(path+".folder", "must be relative")
		case strings.Contains(e.Folder, `\`):
			p.add(path+".folder", "must use forward slashes")
		case oneOf("..", strings.Split(e.Folder, "/")):
			p.add(path+".folder", `must not contain ".."`)
		case folders[e.Folder]:
			p.add(path+".folder", "duplicate %q", e.Folder)
		}
		folders[e.Folder] = true

		p.repo(path+".ssot", e.SSOT)
		p.status(path+".status", e.Status)
		if e.Summary != "" {
			p.summary(path+".summary", e.Summary)
		}
	}
	return p.err()
}

// Validate reports every problem in the component pointer.
func (pt *Pointer) Validate() error {
	p := &problems{}
	p.version(pt.Version)
	p.key("project", pt.Project)
	p.repo("ssot", pt.SSOT)
	p.roleRank("role", "rank", pt.Role, pt.Rank)
	return p.err()
}

// ---- parse ----

func decodeStrict(data []byte, out any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("manifest is empty")
		}
		return err
	}
	return nil
}

// ParseProject decodes project.yaml (unknown fields are errors) and validates it.
func ParseProject(data []byte) (*Project, error) {
	var p Project
	if err := decodeStrict(data, &p); err != nil {
		return nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

// ParseWorkspace decodes workspace.yaml (unknown fields are errors) and validates it.
func ParseWorkspace(data []byte) (*Workspace, error) {
	var w Workspace
	if err := decodeStrict(data, &w); err != nil {
		return nil, err
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	return &w, nil
}

// ParsePointer decodes ssot.yaml (unknown fields are errors) and validates it.
func ParsePointer(data []byte) (*Pointer, error) {
	var pt Pointer
	if err := decodeStrict(data, &pt); err != nil {
		return nil, err
	}
	if err := pt.Validate(); err != nil {
		return nil, err
	}
	return &pt, nil
}

// ---- marshal ----

var flowMapRe = regexp.MustCompile(`(?m)^(\s*[a-z_]+: )\{(.+)\}$`)

func encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	// yaml.v3 writes flow maps as {a: 1}; the standard writes { a: 1 }.
	return flowMapRe.ReplaceAll(buf.Bytes(), []byte("$1{ $2 }")), nil
}

// MarshalProject writes project.yaml in struct field order, 2-space indent, LF.
func MarshalProject(p *Project) ([]byte, error) {
	c := *p
	if c.Harness.Overrides == nil {
		c.Harness.Overrides = []Override{}
	}
	return encode(&c)
}

func MarshalWorkspace(w *Workspace) ([]byte, error) { return encode(w) }

func MarshalPointer(p *Pointer) ([]byte, error) { return encode(p) }
