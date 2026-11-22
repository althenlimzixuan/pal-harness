package state

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Version is a harness release, major.minor.patch.
type Version [3]int

// ParseVersion reads "1.8.0" or "v1.8.0".
func ParseVersion(s string) (Version, error) {
	var v Version
	parts := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(parts) != 3 {
		return v, fmt.Errorf("version %q must look like 1.8.0", s)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return v, fmt.Errorf("version %q must look like 1.8.0", s)
		}
		v[i] = n
	}
	return v, nil
}

// Less reports whether v is an earlier release than o.
func (v Version) Less(o Version) bool {
	for i := range v {
		if v[i] != o[i] {
			return v[i] < o[i]
		}
	}
	return false
}

func (v Version) String() string { return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2]) }

// FieldSince is the harness release that added each field after the state
// package arrived in 1.3.0, keyed by file kind and field path. A field not
// listed is from 1.3.0, which every reader knows.
var FieldSince = map[string]string{
	"features.results.check":  "1.8.0",
	"phase-state.qa":          "1.8.0",
	"phase-state.owner_stops": "1.9.0",
}

// Newer is a field a file uses that a reader on an older release cannot read.
type Newer struct {
	Field string // path within the file, like owner_stops or results.check
	Since string // the release that added it
}

// NewerFields lists, sorted by field, each FieldSince field that the file
// actually uses and that is newer than reader. kind is "phase-state" or
// "features". A file that does not decode gives nil: its findings say why.
func NewerFields(kind string, b []byte, reader Version) []Newer {
	var doc any
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return nil
	}
	used := map[string]bool{}
	var walk func(path string, n any)
	walk = func(path string, n any) {
		switch t := n.(type) {
		case map[string]any:
			for k, c := range t {
				p := k
				if path != "" {
					p = path + "." + k
				}
				used[p] = true
				walk(p, c)
			}
		case []any:
			for _, c := range t {
				walk(path, c)
			}
		}
	}
	walk("", doc)
	var out []Newer
	for key, since := range FieldSince {
		field, ok := strings.CutPrefix(key, kind+".")
		if !ok || !used[field] {
			continue
		}
		if v, err := ParseVersion(since); err == nil && reader.Less(v) {
			out = append(out, Newer{field, since})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Field < out[j].Field })
	return out
}
