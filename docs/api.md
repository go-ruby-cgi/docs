# Usage & API

The public API lives at the module root (`github.com/go-ruby-cgi/cgi`). It is **Ruby-shaped but Go-idiomatic**: `Escape` / `Unescape` / `ParseQuery` mirror `CGI.escape` / `CGI.unescape` / `CGI.parse`, while the surface follows Go conventions — plain `string` and `map[string][]string` values, no global state.

!!! success "Status: implemented"
    The library is built and importable as `github.com/go-ruby-cgi/cgi`, bound into
    `rbgo` as a native module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-cgi/cgi
```

## Worked example

```go
package main

import (
	"fmt"

	"github.com/go-ruby-cgi/cgi"
)

func main() {
	fmt.Println(cgi.Escape("a b&c"))                 // a+b%26c
	fmt.Println(cgi.Unescape("a+b%26c"))             // a b&c
	fmt.Println(cgi.EscapeURIComponent("a b"))       // a%20b
	fmt.Println(cgi.EscapeHTML(`<a href="x">&'`))    // &lt;a href=&quot;x&quot;&gt;&amp;&#39;
	fmt.Println(cgi.UnescapeHTML("&#9731; &amp;"))   // ☃ &
	fmt.Println(cgi.ParseQuery("a=1&b=2&a=3"))       // map[a:[1 3] b:[2]]
}
```

## Shape

```go
// Form (application/x-www-form-urlencoded) encoding — CGI.escape / CGI.unescape.
func Escape(s string) string
func Unescape(s string) string

// URI-component encoding — CGI.escapeURIComponent / CGI.unescapeURIComponent.
func EscapeURIComponent(s string) string
func UnescapeURIComponent(s string) string

// HTML-entity coding — CGI.escapeHTML / CGI.unescapeHTML.
func EscapeHTML(s string) string
func UnescapeHTML(s string) string

// Element-tag (un)escaping — CGI.escapeElement / CGI.unescapeElement.
func EscapeElement(s string, elements ...string) string
func UnescapeElement(s string, elements ...string) string

// Query parsing — CGI.parse.
func ParseQuery(query string) map[string][]string
```

## Behaviour notes

- **`Escape` / `Unescape`** — `application/x-www-form-urlencoded`: every byte
  outside the unreserved set `A–Z a–z 0–9 - . _ ~` is percent-encoded (uppercase
  hex), a space becomes `+`, and `Unescape` decodes `+`→space and `%XX`, leaving
  malformed escapes verbatim (it never raises).
- **`EscapeURIComponent` / `UnescapeURIComponent`** — like the above but a space
  is `%20` (and `Unescape` does not treat `+` as a space).
- **`EscapeHTML` / `UnescapeHTML`** — encodes `& < > " '` (the apostrophe as
  `&#39;`); decoding recognises those five names plus `&apos;`, decimal `&#NN;`
  and hex `&#xHH;` / `&#XHH;` numeric entities (emitting the raw UTF-8 bytes of
  the code point, surrogates included, exactly as MRI does), and rejects unknown
  or overflowing entities verbatim.
- **`EscapeElement` / `UnescapeElement`** — HTML-(un)escape only the start/end
  tags of named elements, matching the element name case-insensitively at a word
  boundary.
- **`ParseQuery`** — `CGI.parse`: split on `&` and `;`, form-decode keys and
  values, accumulate repeated keys, and return an empty slice for a bare key.

## MRI conformance

Correctness is defined by reference Ruby. A **differential oracle** runs a corpus
through both the system `ruby` (`CGI.escape`, `CGI.unescapeHTML`, `CGI.parse`, …)
and this library and compares the results **byte-for-byte** — not approximated
from memory. The oracle tests skip themselves where `ruby` is not on `PATH` (e.g.
the qemu arch lanes), and where the `cgi` gem is not installed for the
`CGI.parse` case, so the cross-arch builds still validate the library.

## Relationship to Ruby

`go-ruby-cgi/cgi` is **standalone and reusable**, and is the CGI backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by `rbgo` as a native
module — the same way [go-ruby-yaml](https://github.com/go-ruby-yaml),
[go-ruby-regexp](https://github.com/go-ruby-regexp) and
[go-ruby-erb](https://github.com/go-ruby-erb) are bound. The dependency runs the
other way: this library has no dependency on the Ruby runtime.
