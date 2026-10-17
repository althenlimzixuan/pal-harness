package seq

import "testing"

func TestMaxSpec(t *testing.T) {
	if got := MaxSpec(nil); got != 0 {
		t.Fatalf("nil: %d", got)
	}
	names := []string{"specs/001-a", "specs/012-b", "specs/notes", "specs/7-x", "docs/004-c", "origin/docs/013-d", "upstream/docs/0140-e"}
	if got := MaxSpec(names); got != 13 {
		t.Fatalf("got %d, want 13", got)
	}
}

func TestMaxADR(t *testing.T) {
	if got := MaxADR(nil); got != 0 {
		t.Fatalf("nil: %d", got)
	}
	names := []string{"doc/ADR/0001-a.md", "doc/ADR/0007-b.md", "doc/ADR/README.md", "doc/ADR/0009-c.txt", "0003-d.md"}
	if got := MaxADR(names); got != 7 {
		t.Fatalf("got %d, want 7", got)
	}
}

func TestSpecAndADRNumber(t *testing.T) {
	if n, ok := Spec("042-x"); !ok || n != 42 {
		t.Fatalf("Spec = %d, %v", n, ok)
	}
	if _, ok := Spec("42-x"); ok {
		t.Fatal("two digits must not match")
	}
	if n, ok := ADR("0042-x.md"); !ok || n != 42 {
		t.Fatalf("ADR = %d, %v", n, ok)
	}
	if _, ok := ADR("0042-x.txt"); ok {
		t.Fatal(".txt must not match")
	}
}
