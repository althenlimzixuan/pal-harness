package blocks

import (
	"errors"
	"strings"
	"testing"

	"github.com/althenlimzixuan/pal-harness/manifest"
)

const wantComponents = "<!-- pal:begin components (generated from project.yaml; do not edit) -->\n" +
	"| Folder | Role | Rank | Stack | Repo | Status |\n" +
	"|---|---|---|---|---|---|\n" +
	"| `api/` | provider | 1 | go_api | https://github.com/althenlimzixuan/fhph-api | active |\n" +
	"<!-- pal:end components -->"

const wantRegistry = "<!-- pal:begin registry (generated from workspace.yaml; do not edit) -->\n" +
	"| Folder | Key | Summary | Status | SSOT |\n" +
	"|---|---|---|---|---|\n" +
	"| `_IT-Internal-Tools\\PC-Pal Center\\` | pc | Pal Center | active | https://github.com/althenlimzixuan/pal-center |\n" +
	"| `NOSUM\\` | ns | — | paused | https://github.com/althenlimzixuan/ns |\n" +
	"<!-- pal:end registry -->"

func TestComponentsTableExact(t *testing.T) {
	p := &manifest.Project{Components: []manifest.Component{{
		Folder: "api", Repo: "https://github.com/althenlimzixuan/fhph-api",
		Role: "provider", Rank: 1, Stack: []string{"go_api"}, Status: "active"}}}
	if got := ComponentsTable(p); got != wantComponents {
		t.Fatalf("got:\n%s\nwant:\n%s", got, wantComponents)
	}
	multi := &manifest.Project{Components: []manifest.Component{{Folder: "a", Stack: []string{"go_api", "tbd"}}}}
	if !strings.Contains(ComponentsTable(multi), "go_api, tbd") {
		t.Fatal("stacks must join with ', '")
	}
	empty := ComponentsTable(&manifest.Project{})
	if strings.Count(empty, "\n") != 3 {
		t.Fatalf("empty list should render header and separator only: %q", empty)
	}
}

func TestRegistryTableExact(t *testing.T) {
	w := &manifest.Workspace{Projects: []manifest.WorkspaceProject{
		{Key: "pc", Name: "Pal Center", Folder: "_IT-Internal-Tools/PC-Pal Center", SSOT: "https://github.com/althenlimzixuan/pal-center", Status: "active", Summary: "Pal Center"},
		{Key: "ns", Name: "No Summary", Folder: "NOSUM", SSOT: "https://github.com/althenlimzixuan/ns", Status: "paused"},
	}}
	want := wantRegistry
	if got := RegistryTable(w); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestReplaceKeepsProse(t *testing.T) {
	before, after := "# Title\n\nSome prose.\n\n", "\n\nMore prose.\n"
	doc := before + "<!-- pal:begin components (old) -->\nstale\n<!-- pal:end components -->" + after
	got, err := Replace(doc, Components, wantComponents)
	if err != nil {
		t.Fatal(err)
	}
	if got != before+wantComponents+after {
		t.Fatalf("got %q", got)
	}
}

func TestReplaceKeepsCRLF(t *testing.T) {
	doc := "# T\r\n\r\n<!-- pal:begin components (x) -->\r\nstale\r\n<!-- pal:end components -->\r\ntail\r\n"
	got, err := Replace(doc, Components, wantComponents)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ReplaceAll(got, "\r\n", ""), "\n") {
		t.Fatalf("bare LF in %q", got)
	}
	ok, err := Matches(got, Components, wantComponents)
	if err != nil || !ok {
		t.Fatalf("Matches = %v, %v", ok, err)
	}
	if !strings.HasSuffix(got, "<!-- pal:end components -->\r\ntail\r\n") {
		t.Fatalf("tail changed: %q", got)
	}
}

func TestMatchesIgnoresEOL(t *testing.T) {
	doc := strings.ReplaceAll("x\n"+wantComponents+"\ny\n", "\n", "\r\n")
	ok, err := Matches(doc, Components, wantComponents)
	if err != nil || !ok {
		t.Fatalf("Matches = %v, %v", ok, err)
	}
	ok, _ = Matches(strings.Replace(doc, "provider", "platform", 1), Components, wantComponents)
	if ok {
		t.Fatal("different content must not match")
	}
	if _, err := Matches("nothing", Components, wantComponents); !errors.Is(err, ErrNoBlock) {
		t.Fatalf("err = %v", err)
	}
	ex, _ := Extract(doc, Components)
	if ex != wantComponents {
		t.Fatalf("Extract = %q", ex)
	}
}

func TestBadBlocks(t *testing.T) {
	b, e := "<!-- pal:begin components (x) -->", "<!-- pal:end components -->"
	cases := []struct {
		name, doc string
		want      error
	}{
		{"no begin", "prose\n", ErrNoBlock},
		{"only end", e + "\n", ErrBadBlock},
		{"two begins", b + "\n" + b + "\n" + e + "\n", ErrBadBlock},
		{"end before begin", e + "\n" + b + "\n", ErrBadBlock},
		{"begin without end", b + "\nfoo\n", ErrBadBlock},
		{"two ends", b + "\n" + e + "\n" + e + "\n", ErrBadBlock},
		{"other block only", "<!-- pal:begin registry (x) -->\n<!-- pal:end registry -->\n", ErrNoBlock},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Replace(c.doc, Components, wantComponents)
			if got != "" || !errors.Is(err, c.want) {
				t.Fatalf("got %q, %v; want %v", got, err, c.want)
			}
			if _, err := Extract(c.doc, Components); !errors.Is(err, c.want) {
				t.Fatalf("Extract err = %v", err)
			}
		})
	}
}
