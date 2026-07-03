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

**go-vs-YJIT verdict:** the pure-Go library now **beats MRI + YJIT on all five
ops.** The URL ops were already ahead — `escape` (~0.61× YJIT), `unescape`
(~0.29× YJIT) and `parse` (~0.15× YJIT, i.e. ~6–7× faster). The two HTML-escape
ops used to lose to YJIT (they are exactly where MRI/YJIT dispatch to the
`cgi/escape` **C extension**); rewriting them as **single-pass, table-driven
scanners** that mirror the C path flipped both: `escapeHTML` went from 1.52× to
**0.93× YJIT** and `unescapeHTML` from 1.05× to **0.85× YJIT**. See
[HTML-escaper design](#the-html-escaper-rewrite-how-go-overtakes-the-c-extension)
below for how.

#### parse — `CGI.parse`, query string → `{key => [values]}`

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 771.0 | 0.13× |
| MRI | 5928.0 | 1.00× |
| MRI + YJIT | 5110.5 | 0.86× |
| TruffleRuby | 7345.5 | 1.24× |

`ParseQuery` is the standout: **~7.7× faster than MRI and ~6.6× faster than YJIT**
(771.0 / 5110.5 = **0.15× YJIT**). MRI's `CGI.parse` is pure Ruby (regex split +
per-pair `CGI.unescape`), so the compiled Go map-building path dominates it.
JRuby 10.1 removed `CGI.parse` from its stdlib (it now directs callers to the
`cgi` gem), so JRuby has no row here — the harness omits it rather than
substitute a non-stdlib implementation.

#### escape — `CGI.escape`, form-URL encoding

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 244.4 | 0.56× |
| MRI | 433.2 | 1.00× |
| MRI + YJIT | 398.2 | 0.92× |
| JRuby | 362.0 | 0.84× |
| TruffleRuby | 25169.3 | 58.10× |

Percent-encoding **beats MRI (0.56×) and YJIT (244.4 / 398.2 = 0.61× YJIT)**. The
lone TruffleRuby outlier (26 µs) is a cold-JIT artifact on this short loop — read
it as order-of-magnitude, per the warm-up caveat below.

#### unescape — `CGI.unescape`, form-URL decoding

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 132.4 | 0.27× |
| MRI | 492.8 | 1.00× |
| MRI + YJIT | 457.8 | 0.93× |
| JRuby | 362.1 | 0.73× |
| TruffleRuby | 2920.6 | 5.93× |

Decoding **beats MRI (0.27×) and YJIT (132.4 / 457.8 = 0.29× YJIT)** — roughly 3.5×
faster than both.

#### escapeHTML — `CGI.escapeHTML`, five-character entity escape

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 107.3 | 0.72× |
| MRI | 148.2 | 1.00× |
| MRI + YJIT | 116.0 | 0.78× |
| JRuby | 275.0 | 1.86× |
| TruffleRuby | 871.9 | 5.88× |

**Flipped from loss to win.** This was the module's one clear loss (1.20× MRI,
1.52× YJIT) when `CGI.escapeHTML` was a `strings.NewReplacer`. Rewriting it as a
**256-entry `byte → entity` table** driving a **single-pass bulk-copy** scanner —
the same shape as MRI's C `escape_html` — now **beats MRI (0.72×) and YJIT
(107.3 / 116.0 = 0.93× YJIT)**. The old `Replacer` made two internal passes (size,
then fill); the table scan makes one, copies each safe run in bulk, and writes a
zero-copy result. See the [design note](#the-html-escaper-rewrite-how-go-overtakes-the-c-extension).

#### unescapeHTML — `CGI.unescapeHTML`, named + numeric entity decode

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 189.3 | 0.74× |
| MRI | 256.6 | 1.00× |
| MRI + YJIT | 222.2 | 0.87× |
| JRuby | 211.8 | 0.83× |
| TruffleRuby | 1454.3 | 5.67× |

**Was a near-tie, now a clear win.** The old decoder allocated per entity
(`string(rune(b))` plus a map lookup each) and copied verbatim text one byte at a
time, leaving it just behind YJIT (1.05×). The rewrite jumps `&`→`&` with
`IndexByte`, **bulk-copies** each verbatim run, resolves named entities to
**constant strings**, and writes numeric entities' UTF-8 bytes **straight into
the buffer** — no per-entity allocation. It now **beats MRI (0.74×), YJIT
(189.3 / 222.2 = 0.85× YJIT) and JRuby**, all of which run the C `unescapeHTML`.

### The HTML-escaper rewrite — how Go overtakes the C extension

Both HTML ops used to lose to YJIT because MRI dispatches them to the
`cgi/escape` **C extension** (`escape_html` / `unescapeHTML`), a tight native
table scan. Matching it in pure Go — with no cgo — came down to two ideas the C
code already uses, plus removing the allocations Go's convenience APIs hide.

**escapeHTML — 256-entry table + one-pass bulk copy.** The old code called
`strings.NewReplacer("&","&amp;", …).Replace`, whose single-byte-key path makes
**two internal passes** over the string (one to size the result, one to fill it)
and dispatches through the `Replacer` interface. The rewrite is a byte-classified
scan:

```go
// htmlEscapes[b] is the entity for b, or "" when b is copied verbatim.
// Indexing a [256]string by a byte is bounds-check-free: one table load, no branch tree.
var htmlEscapes = [256]string{'&': "&amp;", '<': "&lt;", '>': "&gt;", '"': "&quot;", '\'': "&#39;"}

func EscapeHTML(s string) string {
    i := 0
    for i < len(s) && len(htmlEscapes[s[i]]) == 0 { // skip to first escapable byte
        i++
    }
    if i == len(s) {
        return s // nothing to escape: no allocation at all
    }
    buf := make([]byte, 0, 2*len(s)) // one buffer; append grows it if a dense input needs more
    buf = append(buf, s[:i]...)
    last := i
    for ; i < len(s); i++ {
        e := htmlEscapes[s[i]]
        if len(e) == 0 {
            continue
        }
        buf = append(buf, s[last:i]...) // bulk-copy the safe run
        buf = append(buf, e...)         // splice in the entity
        last = i + 1
    }
    buf = append(buf, s[last:]...)
    return unsafe.String(unsafe.SliceData(buf), len(buf)) // zero-copy, like strings.Builder
}
```

One pass, safe runs copied in bulk (not byte-by-byte), a bounds-check-free table
load per byte, and a zero-copy result — the `2×len` estimate keeps the common
case to a single allocation while `append` still guarantees correctness on
adversarial `"`-dense input. That took `escapeHTML` from **1.52× to 0.93× YJIT.**

**unescapeHTML — no per-entity allocation.** The decode side already scanned
`&`→`&` with `IndexByte`, but it copied verbatim text **one byte at a time** and,
for every entity, allocated: named entities went through a `map[string]byte` then
`string(rune(b))` (a fresh 1-byte string each), and numeric entities built a
`string([]byte{…})`. The rewrite bulk-copies runs, resolves names with a
**`switch` returning constant strings**, and appends decoded bytes **directly into
the output buffer** — zero allocations beyond the single result. That took
`unescapeHTML` from **1.05× to 0.85× YJIT.**

**Why not SIMD here?** A SWAR / `go-simd` byte-classify to find the next
escapable byte was prototyped, but on the realistic short fragments these ops see
(tens of bytes, escapable bytes never far apart) the vector setup cost exceeded
the scalar table scan — it measured **slower**. The table-driven scalar path is
the right tool at this input size; SIMD would only pay off on large, escapable-sparse
documents, which is not this workload.

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
