# Roadmap

`go-ruby-cgi/cgi` is grown **test-first**, each capability differential-tested against MRI
rather than built in isolation. The escaping and query-parsing surface of Ruby's
CGI — the deterministic, interpreter-independent core extracted from rbgo's
internals — is **complete**.

| Stage | What | Status |
| --- | --- | --- |
| `Escape` / `Unescape` | `application/x-www-form-urlencoded`: every byte outside `A–Z a–z 0–9 - . _ ~` is percent-encoded (uppercase hex), space → `+`, and `Unescape` decodes `+`→space and `%XX`, leaving malformed escapes verbatim (never raises). | **Done** |
| `EscapeURIComponent` / `UnescapeURIComponent` | Like the form encoders but space is `%20` (and `Unescape` does not treat `+` as a space). | **Done** |
| `EscapeHTML` / `UnescapeHTML` | Encodes `& < > " '` (apostrophe as `&#39;`); decoding recognises those five names plus `&apos;`, decimal `&#NN;` and hex `&#xHH;` / `&#XHH;` numeric entities (raw UTF-8 bytes, surrogates included), rejecting unknown or overflowing entities verbatim. | **Done** |
| `EscapeElement` / `UnescapeElement` | HTML-(un)escape only the start/end tags of named elements, matching the element name case-insensitively at a word boundary. | **Done** |
| `ParseQuery` | `CGI.parse`: split on `&` and `;`, form-decode keys and values, accumulate repeated keys, and return an empty slice for a bare key. | **Done** |
| Differential oracle & coverage | A corpus run through both the system `ruby` (`CGI.escape`, `CGI.unescapeHTML`, `CGI.parse`, …) and this package, compared byte-for-byte; 100% coverage, gofmt + go vet clean, green across all six 64-bit Go arches and three OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **Compute core only.** The request/response cycle, cookies, multipart form
  handling and the HTML-generation DSL — everything that needs a live server or
  CGI environment — is out of scope. This library is the deterministic string
  transforms.
- **No interpreter.** The library never runs arbitrary Ruby; anything that needs
  a live binding is the consumer's job — that is why `rbgo` binds this module
  rather than the reverse.
- **Reference is reference Ruby (MRI).** Byte-for-byte conformance targets MRI's
  CGI behaviour; differences across MRI releases are matched to the reference
  used by the differential oracle.
- **Standalone & reusable.** The module has no dependency on the Ruby runtime;
  the dependency runs the other way.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
deterministic/interpreter split.
