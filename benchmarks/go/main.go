// SPDX-License-Identifier: BSD-3-Clause
package main

import (
	"os"

	"github.com/go-ruby-cgi/cgi"
)

// Fixed representative inputs, byte-identical to ruby/cgi.rb.
const (
	// A realistic form value: spaces, reserved bytes, and multi-byte UTF-8
	// (café, René, em-dash, cœur) so percent-encoding covers >1-byte runes.
	urlInput = "Café René & Co. — 100% \"cœur\", <tag>? a=1 b=2 (x/y) #frag"

	// An HTML fragment exercising all five CGI.escapeHTML characters.
	htmlInput = "<a href=\"/u?id=42&t=1\">Tom & Jerry's \"Big\" <Adventure></a> 5 > 3 & 2 < 4"

	// A pre-escaped fragment mixing named, decimal and hex numeric entities —
	// the CGI.unescapeHTML hot path (amp, #39, quot, lt, gt, decimal #8212,
	// é as #233, hex #x263A, apos).
	htmlEscaped = "Tom &amp; Jerry&#39;s &quot;Big&quot; &lt;Adventure&gt; &#8212; caf&#233; &#x263A; &apos;end&apos;"

	// A realistic query string: repeated (array) keys, a %-encoded value with a
	// multi-byte rune, a '+'-as-space value, and a valueless trailing key.
	queryInput = "name=John+Doe&email=j%40x.com&tags[]=a&tags[]=b&tags[]=c&q=caf%C3%A9+%26+co&page=2&sort=desc&empty="
)

func main() {
	escURL := cgi.Escape(urlInput)      // hot input for Unescape
	escHTML := cgi.EscapeHTML(htmlInput) // referenced for parity via CHECK

	if len(os.Args) > 1 && os.Args[1] == "check" {
		check("escape", cgi.Escape(urlInput))
		check("unescape", cgi.Unescape(escURL))
		check("escapeHTML", escHTML)
		check("unescapeHTML", cgi.UnescapeHTML(htmlEscaped))
		check("parse", parseString(cgi.ParseQuery(queryInput)))
		return
	}

	bench("escape", 5000, func() { sink = cgi.Escape(urlInput) })
	bench("unescape", 5000, func() { sink = cgi.Unescape(escURL) })
	bench("escapeHTML", 5000, func() { sink = cgi.EscapeHTML(htmlInput) })
	bench("unescapeHTML", 5000, func() { sink = cgi.UnescapeHTML(htmlEscaped) })
	bench("parse", 2000, func() { sink = cgi.ParseQuery(queryInput) })
}
