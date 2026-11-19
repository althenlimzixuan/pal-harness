// Package fingerprint hashes the git objects a CI job reads, so a job whose
// files already passed can be skipped (the CI gate in devops_cicd.md).
package fingerprint

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// Job is one CI job and the paths it reads.
type Job struct {
	Name  string
	Paths []string
}

// Sum is a job's fingerprint.
type Sum struct {
	Name string
	Sum  string
}

// Lookup returns the git object id of a path, or "" when the path does not exist.
type Lookup func(path string) (string, error)

// jobName is safe in a ref name and as a JSON key in a workflow expression.
var jobName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// ParseSpec reads one "<job>: <path> <path> ..." line per job; "#" starts a comment.
func ParseSpec(r io.Reader) ([]Job, error) {
	var jobs []Job
	seen := map[string]bool{}
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line, _, _ := strings.Cut(sc.Text(), "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, rest, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("line %d: want <job>: <path> ...", n)
		}
		name = strings.TrimSpace(name)
		if !jobName.MatchString(name) {
			return nil, fmt.Errorf("line %d: job name %q must match %s", n, name, jobName)
		}
		if seen[name] {
			return nil, fmt.Errorf("line %d: job %q is listed twice", n, name)
		}
		seen[name] = true
		job := Job{Name: name}
		for f := range strings.FieldsSeq(rest) {
			p, err := CheckPath(f)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", n, err)
			}
			job.Paths = append(job.Paths, p)
		}
		if len(job.Paths) == 0 {
			return nil, fmt.Errorf("line %d: job %q has no paths", n, name)
		}
		jobs = append(jobs, job)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, errors.New("the spec lists no jobs")
	}
	return jobs, nil
}

// CheckPath accepts a repo-relative path with forward slashes ("." is the whole
// tree) and returns it without a trailing slash.
func CheckPath(p string) (string, error) {
	if strings.Contains(p, `\`) {
		return "", fmt.Errorf("path %q: use forward slashes", p)
	}
	trimmed := strings.TrimRight(p, "/")
	if trimmed == "" || strings.HasPrefix(p, "/") || strings.Contains(trimmed, ":") {
		return "", fmt.Errorf("path %q must be relative to the repo root", p)
	}
	if slices.Contains(strings.Split(trimmed, "/"), "..") {
		return "", fmt.Errorf("path %q must be relative to the repo root, with no ..", p)
	}
	return trimmed, nil
}

// Compute hashes, per job, its name and one "<path> <object id or missing>" line
// for each of its paths and then each shared path.
func Compute(jobs []Job, shared []string, lookup Lookup) ([]Sum, error) {
	var out []Sum
	for _, j := range jobs {
		lines := []string{j.Name}
		for _, p := range append(append([]string{}, j.Paths...), shared...) {
			id, err := lookup(p)
			if err != nil {
				return nil, fmt.Errorf("%s: %v", p, err)
			}
			if id == "" {
				id = "missing"
			}
			lines = append(lines, p+" "+id)
		}
		h := sha1.Sum([]byte(strings.Join(lines, "\n")))
		out = append(out, Sum{Name: j.Name, Sum: hex.EncodeToString(h[:])})
	}
	return out, nil
}

// NewGit returns a Lookup reading the objects of commit rev in the repo at dir.
func NewGit(dir, rev string) (Lookup, error) {
	if _, ok, err := revParse(dir, rev+"^{commit}"); err != nil || !ok {
		if err == nil {
			err = fmt.Errorf("%s is not a commit", rev)
		}
		return nil, fmt.Errorf("%s: not a git repository with commit %s: %v", dir, rev, err)
	}
	return func(path string) (string, error) {
		spec := rev + ":" + path
		if path == "." {
			spec = rev + "^{tree}"
		}
		id, ok, err := revParse(dir, spec)
		if err != nil || !ok {
			return "", err
		}
		return id, nil
	}, nil
}

// revParse runs git rev-parse --verify --quiet: ok is false when the object does
// not exist (exit 1 with no output); err is any other failure.
func revParse(dir, spec string) (string, bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", spec)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == 1 && strings.TrimSpace(stderr.String()) == "" {
			return "", false, nil
		}
		return "", false, fmt.Errorf("%v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), true, nil
}
