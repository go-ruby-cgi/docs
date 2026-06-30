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

!!! note "No fabricated numbers"
    This page documents the **methodology**; the headline figures for this module
    are produced by running the harness above on a pinned host and are reported
    with the run date and exact runtime versions. Numbers are never transcribed
    from memory or estimated — only measured results from a real run are
    published, byte-checked against MRI first.
