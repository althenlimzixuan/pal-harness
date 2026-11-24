package state

import (
	"reflect"
	"strings"
	"testing"
)

const retroOK = "---\nexpected:\n" +
	"  - {line: \"You can see the tally\", source: \"summary 3\", result: met, evidence: \"qa/summary.md row 4\"}\n" +
	"  - {line: \"The card moves\", source: \"F-039 step 2\", result: partly, evidence: \"TestCardMoves\", why: \"only on refresh\"}\n" +
	"  - {line: \"A banner shows\", source: \"F-040 expect 1\", result: missed, evidence: \"none\", why: \"no evidence\"}\n" +
	"---\n# Phase 026 — x: retro\n"

// pal-harness 1.12.0: the retro's expected list (design D).
func TestParseRetro(t *testing.T) {
	es, findings := ParseRetro([]byte(retroOK))
	if len(findings) != 0 {
		t.Fatalf("valid retro: findings %v", findings)
	}
	if len(es) != 3 || es[0].Result != "met" || es[1].Why != "only on refresh" || es[2].Source != "F-040 expect 1" {
		t.Fatalf("expected = %+v", es)
	}

	t.Run("no front matter is not checked", func(t *testing.T) {
		if es, f := ParseRetro([]byte("# Phase 001 — x: retro\n\n## Went well\n")); es != nil || f != nil {
			t.Fatalf("got %v %v", es, f)
		}
	})

	rep := func(old, new string) string {
		if !strings.Contains(retroOK, old) {
			t.Fatalf("fixture lacks %q", old)
		}
		return strings.Replace(retroOK, old, new, 1)
	}
	for _, c := range []struct {
		name, in string
		want     []string
	}{
		{"result", rep("result: met", "result: done"), []string{`expected 1: result "done" is not one of met, partly, missed`}},
		{"line", rep(`line: "You can see the tally", `, ""), []string{"expected 1: line is required"}},
		{"evidence", rep(`, evidence: "qa/summary.md row 4"`, ""), []string{"expected 1: evidence is required"}},
		{"source", rep(`source: "summary 3"`, `source: "the summary"`),
			[]string{`expected 1: source "the summary" is not summary N, F-NNN step N or F-NNN expect N`}},
		{"why for partly", rep(`, why: "only on refresh"`, ""), []string{"expected 2: partly needs why"}},
		{"why for missed", rep(`, why: "no evidence"`, ""), []string{"expected 3: missed needs why"}},
		{"unknown field", rep("result: met", "result: met, score: 3"), []string{"field score not found"}},
		{"unclosed", strings.Replace(retroOK, "---\n# Phase", "# Phase", 1), []string{"front matter is not closed"}},
		{"empty list", "---\nexpected: []\n---\n", []string{"expected is empty"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, got := ParseRetro([]byte(c.in))
			if len(got) != len(c.want) {
				t.Fatalf("findings %v, want %v", got, c.want)
			}
			for i, w := range c.want {
				if !strings.Contains(got[i], w) {
					t.Errorf("finding %q lacks %q", got[i], w)
				}
			}
		})
	}

	t.Run("CRLF", func(t *testing.T) {
		es, f := ParseRetro([]byte(strings.ReplaceAll(retroOK, "\n", "\r\n")))
		if len(f) != 0 || !reflect.DeepEqual(len(es), 3) {
			t.Fatalf("got %v %v", es, f)
		}
	})
}
