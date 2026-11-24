# Test summary — (component), phase <NNN>

Verdict: (pass, or blocked: the reason in one line)

Commit: (sha) · tree: (tree hash) — a staging report carries the staging deploy's tree hash

Every line, and every claim in a note, ends with its source in brackets: a report file
(`[unit_test.html]`), a command and its exit code (`[make reports, exit 0]`), or a file and line.

1. Tests: (passed) of (total) passed, (failed) failed. [unit_test.html]
2. Test coverage: (percent). Minimum is 80%; below that blocks the sign-off. [test_coverage.html]
3. Slowest calls: (name) p95 (ms); (name) p95 (ms). [profiling.html]
4. Speed against the previous code: base commit (sha), (N) pairs, old and new alternating on this
   machine; median p95 old (ms) → new (ms), (+/- percent). (`CPU busy: N%` when `pal-harness cpu`
   was still busy.) A call more than 20% slower blocks the sign-off; so does a web page score 5 or
   more points lower. Coverage since the last report: (+/- points). [profiling.html]
5. Regression: (passed / not run yet). [regression.html]
6. Load test: (result, or "not requested"). [load_test.html]
7. Not applicable: (report and reason, or "none").
