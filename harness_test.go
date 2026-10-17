package harness_test

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	harness "github.com/althenlimzixuan/pal-harness"
)

var semver = regexp.MustCompile(`^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

func TestVersionsAgree(t *testing.T) {
	if !semver.MatchString(harness.Version) {
		t.Fatalf("Version %q is not semver", harness.Version)
	}

	raw, err := os.ReadFile(".claude-plugin/plugin.json")
	if err != nil {
		t.Fatal(err)
	}
	var plugin struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &plugin); err != nil {
		t.Fatal(err)
	}
	if plugin.Version != harness.Version {
		t.Errorf("plugin.json version %q != Version %q", plugin.Version, harness.Version)
	}

	changelog, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	var heading string
	for _, line := range strings.Split(string(changelog), "\n") {
		if strings.HasPrefix(line, "## ") {
			heading = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			break
		}
	}
	if heading == "" {
		t.Fatal("CHANGELOG.md has no '## ' heading")
	}
	got, _, _ := strings.Cut(heading, " — ")
	if got != harness.Version {
		t.Errorf("CHANGELOG first heading version %q != Version %q", got, harness.Version)
	}
}

func TestContentEmbedded(t *testing.T) {
	for _, p := range []string{"content/practices/workflow.md", "content/templates/spec.md"} {
		if _, err := harness.Content.ReadFile(p); err != nil {
			t.Errorf("embedded %s: %v", p, err)
		}
	}
}
