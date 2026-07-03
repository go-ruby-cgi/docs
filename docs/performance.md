# Performance

`go-ruby-cgi/cgi` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's CGI utility
methods. This page describes the **comparative benchmark methodology** used to
measure that module against the reference Ruby runtimes, part of the
ecosystem-wide per-module parity suite.

## What is measured

The **same** Ruby script — a representative mix of `CGI.escape`, `CGI.unescapeHTML`
and `CGI.parse` calls — is run under every runtime. `rbgo`'s number reflects
**this pure-Go library doing the work**; every other column is that interpreter's
own `cgi` stdlib. So the comparison is the **Ruby-visible operation**,
apples-to-apples across interpreters. The script prints a deterministic checksum
and its output is checked **byte-identical to MRI** before timing.

- **Method:** best-of-N wall time (best, not mean, to suppress scheduler noise);
  single-shot processes, no warm-up beyond the script's own loop.
- **Runtimes:** `ruby` (MRI, the oracle) and `ruby --yjit`; `jruby` (on the JVM);
  `truffleruby` (GraalVM). JRuby and TruffleRuby are timed **cold, single-shot**,
  so they carry JVM / Graal startup on every run — read them as one-shot
  `ruby file.rb` costs, the same way `rbgo` and MRI are measured, not as
  steady-state JIT numbers.
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules).

## Result (best of 5, ms)

| Runtime | time | vs MRI |
| --- | ---: | ---: |
| **rbgo** (go-ruby-cgi) | 60 | 1.00× |
| MRI (ruby 4.0.5) | 60 | 1.00× |
| MRI + YJIT | 60 | 1.00× |
| JRuby 10.1.0.0 | 1090 | 18.17× |
| TruffleRuby 34.0.1 | 1950 | 32.50× |

rbgo runs on **go-ruby-cgi** at **parity with MRI** (1.00x) on this HTML/URL escape+unescape workload; both ultimately call native escaping. TruffleRuby pays heavy cold warm-up here (1950 ms).

!!! note "Honest framing"
    JRuby and TruffleRuby are timed **cold, single-shot**, so they carry JVM /
    Graal startup on every run — read them as one-shot `ruby file.rb` costs, the
    same way `rbgo` and MRI are measured, not as steady-state JIT numbers. Rows
    that complete in well under ~200 ms carry the most relative noise; treat
    their ratios as order-of-magnitude. These are **real measured numbers** from
    the 2026-06-30 run (Apple M-series; `ruby 4.0.5 +PRISM`, `jruby 10.1.0.0`,
    `truffleruby 34.0.1`) — nothing is fabricated or cherry-picked.

## Library-level benchmark (Go API vs runtimes) — 2026-07-03

This section measures the **pure-Go library directly, through its Go API** — not
the `rbgo` interpreter path recorded above. It isolates the library primitive
from Ruby-interpreter dispatch, answering the parity question head-on: *is the
pure-Go implementation as fast as the reference runtime's own `cgi`?* The **same
workload, same fixed inputs, same iteration counts** run through the Go library
and through each reference runtime's stdlib; every op's output was checked
**byte-identical to MRI** (the run aborts on any mismatch) before any timing.

- **Host:** Apple M4 Max (`Mac16,5`, arm64), macOS 26.5 — **date 2026-07-03**.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` · MRI + YJIT · JRuby 10.1.0.0
  (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native).
- **Method:** each process runs 3 untimed warm-up passes, then 25 timed passes of
  a fixed inner loop, timed with a monotonic clock; the **best** pass is reported
  as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than MRI*.
  Interpreter start-up is outside the timed region, so these are operation costs,
  not `ruby file.rb` process costs.
- **Inputs (fixed, byte-identical across drivers):** a realistic UTF-8 form value
  for `escape`/`unescape`; an HTML fragment exercising all five escaped
  characters for `escapeHTML`; a mixed named/decimal/hex-entity fragment for
  `unescapeHTML`; and a query string with repeated (array) keys and `%`/`+`
  values for `parse`.

**go-vs-YJIT verdict:** the pure-Go library **beats MRI + YJIT on three of the
five ops** — `escape` (~0.58× YJIT), `unescape` (~0.34× YJIT) and `parse`
(~0.15× YJIT, i.e. ~6–7× faster). It **loses the two HTML-escape ops to YJIT**:
`escapeHTML` (~1.5× YJIT) and `unescapeHTML` (~1.05× YJIT, a near-tie). Those two
are exactly where MRI/YJIT call the `cgi/escape` **C extension** (`escape_html`),
a tight table-driven native path that the Go `strings.NewReplacer` cannot yet
match; they are this module's remaining optimization targets.

#### parse — `CGI.parse`, query string → `{key => [values]}`

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 786.0 | 0.14× |
| MRI | 5813.5 | 1.00× |
| MRI + YJIT | 5223.5 | 0.90× |
| TruffleRuby | 14249.2 | 2.45× |

`ParseQuery` is the standout: **~7× faster than MRI and ~6.6× faster than YJIT**
(786.0 / 5223.5 = **0.15× YJIT**). MRI's `CGI.parse` is pure Ruby (regex split +
per-pair `CGI.unescape`), so the compiled Go map-building path dominates it.
JRuby 10.1 removed `CGI.parse` from its stdlib (it now directs callers to the
`cgi` gem), so JRuby has no row here — the harness omits it rather than
substitute a non-stdlib implementation.

#### escape — `CGI.escape`, form-URL encoding

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 246.8 | 0.56× |
| MRI | 444.6 | 1.00× |
| MRI + YJIT | 428.4 | 0.96× |
| JRuby | 460.1 | 1.03× |
| TruffleRuby | 26123.8 | 58.76× |

Percent-encoding **beats MRI (0.56×) and YJIT (246.8 / 428.4 = 0.58× YJIT)**. The
lone TruffleRuby outlier (26 µs) is a cold-JIT artifact on this short loop — read
it as order-of-magnitude, per the warm-up caveat below.

#### unescape — `CGI.unescape`, form-URL decoding

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 154.9 | 0.32× |
| MRI | 488.6 | 1.00× |
| MRI + YJIT | 452.6 | 0.93× |
| JRuby | 363.8 | 0.74× |
| TruffleRuby | 3284.4 | 6.72× |

Decoding **beats MRI (0.32×) and YJIT (154.9 / 452.6 = 0.34× YJIT)** — roughly 3×
faster than both.

#### escapeHTML — `CGI.escapeHTML`, five-character entity escape

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 177.3 | 1.20× |
| MRI | 148.0 | 1.00× |
| MRI + YJIT | 116.4 | 0.79× |
| JRuby | 245.0 | 1.66× |
| TruffleRuby | 1386.4 | 9.37× |

The one clear loss: **1.20× MRI and 177.3 / 116.4 = 1.52× YJIT**. MRI's
`CGI.escapeHTML` is the C `escape_html` extension — a native table scan the Go
`strings.NewReplacer` does not beat. This is the module's top optimization
target (a byte-class table + single-pass builder would close it).

#### unescapeHTML — `CGI.unescapeHTML`, named + numeric entity decode

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 242.4 | 0.92× |
| MRI | 262.8 | 1.00× |
| MRI + YJIT | 231.4 | 0.88× |
| JRuby | 219.4 | 0.83× |
| TruffleRuby | 2659.1 | 10.12× |

A near-tie: the Go decoder **beats MRI (0.92×)** but sits **just behind YJIT**
(242.4 / 231.4 = **1.05× YJIT**) and JRuby, which run the C `unescapeHTML` path.
The gap is within run-to-run noise (~5%); a small tightening of the entity scan
would flip it.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-cgi/docs/tree/main/benchmarks): a
    self-contained Go driver (`go/`, pins the published library via `go.mod`
    pseudo-version — no `replace`), the equivalent `ruby/cgi.rb` workload, and
    `run.sh` (which first verifies Go output is byte-identical to MRI, then times
    every runtime). Run `bash benchmarks/run.sh`; env `OUTER`/`WARM` tune the
    pass budget and `RUBY`/`JRUBY`/`TRUFFLERUBY` select the runtime binaries.

!!! warning "Warm-up budget & noise — honest framing"
    Numbers reflect a **fixed warm-process budget** (3 warm-up + 25 timed passes
    in one process). The JVM/GraalVM JITs (JRuby, TruffleRuby) may need a larger
    warm-up to reach steady state, so their columns can **understate** peak
    throughput — most visibly TruffleRuby on the shortest loops (the `escape`
    row's 26 µs is one such cold-JIT outlier). Sub-microsecond rows carry the most
    relative noise; treat those ratios as order-of-magnitude. Every number here is
    a **real measured value** from the dated run above — nothing is fabricated,
    estimated, or cherry-picked. The go-ruby column is the pure-Go library; every
    other column is that interpreter's own stdlib doing the equivalent work.
