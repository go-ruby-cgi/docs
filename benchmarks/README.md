<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-cgi` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go `go-ruby-cgi/cgi` library**
against the reference Ruby runtimes (MRI, MRI + YJIT, JRuby, TruffleRuby). It
measures the **library primitives** through their Go API, isolated from the rbgo
interpreter, so the numbers answer: *is the pure-Go implementation as fast as the
reference runtime's own `cgi`?*

## Layout

- `go/`          — self-contained Go driver; `go.mod` pins the published library
  by pseudo-version (no `replace`).
- `ruby/cgi.rb`  — the equivalent workload; `ruby/_harness.rb` is the shared timer.
- `run.sh`       — verifies Go output is byte-identical to MRI, then runs every
  available runtime and prints one Markdown table per sub-benchmark
  (ns/op + ratio vs MRI).

## Run

```sh
bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.

## Operations

Five representative CGI primitives over fixed inputs:

- `escape` / `unescape` — `application/x-www-form-urlencoded` URL escaping of a
  realistic form value (spaces, reserved bytes, multi-byte UTF-8).
- `escapeHTML` / `unescapeHTML` — the five-character HTML-entity escape and its
  inverse (named + decimal + hex numeric entities).
- `parse` — `CGI.parse`, a query string with repeated (array) keys and
  `%`/`+`-encoded values, into a `key → [values]` map.

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as **ns/op**. Interpreter start-up is outside the
timed region. The Go driver and the Ruby script build **identical inputs** and
each op's output is checked **byte-identical to MRI** (via the `check` mode's
hex dump) before any timing is recorded. Results are published, dated, in
[`../docs/performance.md`](../docs/performance.md).
