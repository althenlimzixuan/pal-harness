package mem

import "testing"

// TestAvailable reads the real machine: it must report something above zero.
func TestAvailable(t *testing.T) {
	v, err := Available()
	if err != nil {
		t.Fatal(err)
	}
	if v == 0 {
		t.Fatal("no available memory reported")
	}
}
