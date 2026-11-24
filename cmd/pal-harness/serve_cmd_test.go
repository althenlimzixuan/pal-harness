package main

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuf is a goroutine-safe writer for a command that runs while the test reads its output.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.b.Write(p) }
func (s *syncBuf) String() string              { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

var serveLineRe = regexp.MustCompile(`^serving (.+) at (http://127\.0\.0\.1:\d+/) until \d{2}:\d{2}\n$`)

// pal-harness 1.12.0: serve shows evidence pages to the browser tools, which refuse file: URLs (design I1).
func TestServe(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "summary.html", "<p>QA summary</p>")
	write(t, dir, ".env", "SECRET=1")
	write(t, dir, ".git/config", "[core]")
	write(t, filepath.Dir(dir), "outside.txt", "outside")

	var out, errs syncBuf
	done := make(chan int, 1)
	go func() { done <- run([]string{"serve", "--for", "3s", dir}, &out, &errs) }()

	var url string
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if m := serveLineRe.FindStringSubmatch(out.String()); m != nil {
			url = m[2]
			break
		}
	}
	if url == "" {
		t.Fatalf("no serving line: out=%q err=%q", out.String(), errs.String())
	}
	get := func(path string) (int, string) {
		t.Helper()
		res, err := http.Get(url + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return res.StatusCode, string(b)
	}
	if code, body := get("summary.html"); code != 200 || !strings.Contains(body, "QA summary") {
		t.Errorf("summary.html: %d %q", code, body)
	}
	for _, p := range []string{".env", ".git/config", "%2e%2e/outside.txt", "sub/%2e%2e/%2e%2e/outside.txt"} {
		if code, body := get(p); code != 404 {
			t.Errorf("%s: %d %q, want 404", p, code, body)
		}
	}
	select {
	case code := <-done:
		if code != 0 {
			t.Fatalf("exit %d err=%q", code, errs.String())
		}
	case <-time.After(10 * time.Second):
		t.Fatal("serve did not stop after --for")
	}
	if !strings.Contains(out.String(), "stopped") {
		t.Errorf("no stopped line: %q", out.String())
	}
}

func TestServeUsage(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct {
		name string
		args []string
	}{
		{"no folder", []string{"serve"}},
		{"missing folder", []string{"serve", filepath.Join(dir, "nope")}},
		{"a file", []string{"serve", filepath.Join(dir, "f.txt")}},
		{"for over 2h", []string{"serve", "--for", "3h", dir}},
		{"for zero", []string{"serve", "--for", "0", dir}},
		{"bad port", []string{"serve", "--port", "70000", dir}},
		{"two folders", []string{"serve", dir, dir}},
	} {
		t.Run(c.name, func(t *testing.T) {
			write(t, dir, "f.txt", "x")
			if code, out, errs := do(t, c.args...); code != 2 {
				t.Fatalf("code %d out=%q err=%q", code, out, errs)
			}
		})
	}
}
