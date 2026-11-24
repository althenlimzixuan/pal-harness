package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

// serveMaxFor caps how long serve may run, so a forgotten server always ends.
const serveMaxFor = 2 * time.Hour

// runServe serves a folder read-only on loopback, for evidence pages the
// browser tools refuse to open as file: URLs. It stops after --for or on an
// interrupt: 0 stopped, 2 usage, a missing folder or a port it cannot open.
func runServe(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	port := fs.Int("port", 0, "")
	forDur := fs.Duration("for", 30*time.Minute, "")
	if err := fs.Parse(args); err != nil {
		return usageErr(stderr, "serve: "+err.Error())
	}
	if fs.NArg() != 1 {
		return usageErr(stderr, "serve takes one folder")
	}
	if *port < 0 || *port > 65535 {
		return usageErr(stderr, "serve: --port must be 0 to 65535")
	}
	if *forDur <= 0 || *forDur > serveMaxFor {
		return usageErr(stderr, "serve: --for must be above 0 and at most 2h")
	}
	dir := fs.Arg(0)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return fail(stderr, "serve: %s is not a folder", dir)
	}
	ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(*port)))
	if err != nil {
		return fail(stderr, "serve: %v", err)
	}
	srv := &http.Server{Handler: noDotPaths(http.FileServer(http.Dir(dir))), ReadHeaderTimeout: 10 * time.Second}
	until := now().Add(*forDur)
	fmt.Fprintf(stdout, "serving %s at http://%s/ until %s\n", dir, ln.Addr(), until.Format("15:04"))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, *forDur)
	defer cancel()
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	select {
	case <-ctx.Done():
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return fail(stderr, "serve: %v", err)
		}
	}
	shut, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()
	_ = srv.Shutdown(shut)
	fmt.Fprintln(stdout, "serve: stopped")
	return 0
}

// noDotPaths answers 404 for any path with a segment that starts with a dot
// (.git, .env, ..), so a served folder never shows hidden files.
func noDotPaths(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, seg := range strings.Split(r.URL.Path, "/") {
			if strings.HasPrefix(seg, ".") {
				http.NotFound(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}
