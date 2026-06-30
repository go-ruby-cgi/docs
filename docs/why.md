# Why pure Go

`go-ruby-cgi/cgi` reimplements Ruby's CGI escaping and query-parsing methods in **pure Go,
with cgo disabled**. The slice of Ruby it covers is **deterministic and
interpreter-independent**: URL/form escaping, HTML-entity coding and query parsing
are pure string transforms — given the input string, the result is a pure function
of it, with no live binding and no evaluation of arbitrary Ruby. That is exactly
the part that can — and should — live as a standalone Go library, separate from the
interpreter.

## What it is — and isn't

URL/form escaping, HTML-entity coding and query parsing are pure, deterministic
string transforms that need **no interpreter**, so they live here as pure Go. The
request/response cycle, cookies, multipart form handling and the HTML-generation
DSL — everything that needs a live server or CGI environment — is **out of scope**;
this library is the compute core only.

**Ruby 4.0 note.** In MRI 4.0 the `cgi` library was slimmed to `cgi/escape`:
`CGI.escape`, `CGI.unescape`, `CGI.escapeHTML`, `CGI.unescapeHTML`,
`CGI.escapeURIComponent`, `CGI.unescapeURIComponent`, `CGI.escapeElement` and
`CGI.unescapeElement` are the default surface, while `CGI.parse` now ships in the
separate `cgi` gem. This package provides all of them (`ParseQuery` for
`CGI.parse`).

## Extracted from rbgo, reusable by anyone

This library began life inside [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)'s
`rbgo`. It has been **extracted into a reusable standalone library** so that:

- any Go program can import `github.com/go-ruby-cgi/cgi` directly, with no Ruby runtime;
- the dependency runs the *other* way — `rbgo` binds this module as a native
  module (the same pattern as [go-ruby-yaml](https://github.com/go-ruby-yaml),
  [go-ruby-regexp](https://github.com/go-ruby-regexp) and
  [go-ruby-erb](https://github.com/go-ruby-erb)), rather than this module
  depending on the interpreter;
- the behaviour is pinned by a **differential oracle** against the system
  `ruby`, independent of any one consumer.

## Why pure Go matters here

Because the library is CGO-free and dependency-free, it:

- cross-compiles to every Go target with no C toolchain, and links into a single
  static binary;
- has **no dependency on the Ruby runtime** — the dependency runs the other way;
- can be differentially tested against the `ruby` binary wherever one is on
  `PATH`, while the cross-arch lanes (where `ruby` is absent) still validate the
  library itself.

See [Usage & API](api.md) for the surface and [Roadmap](roadmap.md) for what is
in scope.
