# go-ruby-cgi documentation

**Ruby's CGI escaping and query-parsing methods in pure Go — MRI-compatible, no cgo.**

`go-ruby-cgi/cgi` is a faithful, pure-Go (zero cgo) reimplementation of the escaping and
query-parsing surface of Ruby's CGI utility methods — the deterministic,
interpreter-independent core of MRI 4.0.5's `CGI.escape` / `CGI.unescape`, the
HTML-entity helpers, the URI-component helpers, the element helpers and
`CGI.parse`. Every method matches the system `ruby` byte-for-byte, **without any
Ruby runtime**. The module path is `github.com/go-ruby-cgi/cgi`.

In MRI 4.0 the `cgi` library was slimmed to `cgi/escape`: `CGI.escape`,
`CGI.unescape`, `CGI.escapeHTML`, `CGI.unescapeHTML`, `CGI.escapeURIComponent`,
`CGI.unescapeURIComponent`, `CGI.escapeElement` and `CGI.unescapeElement` are the
default surface, while `CGI.parse` now ships in the separate `cgi` gem. This
package provides all of them (`ParseQuery` for `CGI.parse`).

It was **extracted from rbgo into a reusable standalone
library**: the module is standalone and importable by any Go program, and it is
the CGI backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)
by `rbgo` as a native module — just like
[go-ruby-yaml](https://github.com/go-ruby-yaml),
[go-ruby-regexp](https://github.com/go-ruby-regexp) and
[go-ruby-erb](https://github.com/go-ruby-erb). The dependency runs the other
way: this library has **no dependency on the Ruby runtime**.

!!! success "Status: complete — MRI byte-exact"
    Faithful port of the CGI utility methods: **`Escape` / `Unescape`** (form encoding), **`EscapeURIComponent` / `UnescapeURIComponent`**, **`EscapeHTML` / `UnescapeHTML`** (named, decimal and hex numeric entities), **`EscapeElement` / `UnescapeElement`** (per-element tag escaping), and **`ParseQuery`** (`CGI.parse`). Validated by a **differential oracle** against the system `ruby` / `CGI` — compared byte-for-byte — at 100% coverage, `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets and three OSes.

## Quick taste

```go
fmt.Println(cgi.Escape("a b&c"))                 // a+b%26c
fmt.Println(cgi.Unescape("a+b%26c"))             // a b&c
fmt.Println(cgi.EscapeURIComponent("a b"))       // a%20b
fmt.Println(cgi.EscapeHTML(`<a href="x">&'`))    // &lt;a href=&quot;x&quot;&gt;&amp;&#39;
fmt.Println(cgi.UnescapeHTML("&#9731; &amp;"))   // ☃ &
fmt.Println(cgi.ParseQuery("a=1&b=2&a=3"))       // map[a:[1 3] b:[2]]
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`cgi`](https://github.com/go-ruby-cgi/cgi) | the library — Ruby's CGI escaping & query parsing in pure Go |
| [`docs`](https://github.com/go-ruby-cgi/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-cgi.github.io`](https://github.com/go-ruby-cgi/go-ruby-cgi.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-cgi/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain; dependency-free.
- **MRI byte-exact.** Percent-encoding, HTML-entity coding and query parsing
  match reference Ruby exactly, not approximately, validated by a differential
  oracle against the `ruby` binary.
- **Compute core only.** Pure, deterministic string transforms; anything needing
  a live server or CGI environment is out of scope.
- **Standalone & reusable.** Extracted from rbgo's internals; no dependency on
  the Ruby runtime — the dependency runs the other way.
- **100% test coverage** is the target, enforced as a CI gate, across 6 arches
  and 3 OSes.

## Where to go next

- [Why pure Go](why.md) — why CGI escaping and query parsing are deterministic
  enough to live as a standalone, interpreter-independent Go library.
- [Usage & API](api.md) — the public surface and worked examples.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at [github.com/go-ruby-cgi/cgi](https://github.com/go-ruby-cgi/cgi).
