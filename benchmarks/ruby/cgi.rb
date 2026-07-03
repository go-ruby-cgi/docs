# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
#
# Cross-runtime workload for Ruby's CGI utility methods, byte-identical to the
# pure-Go go-ruby-cgi/cgi driver in ../go. Run with no args to print RESULT
# ns/op lines; run with "check" to print CHECK <label> <hex> lines for a
# byte-for-byte parity diff against the Go driver.

# CGI.escape / unescape / escapeHTML / unescapeHTML come from cgi/escape (present
# in every runtime). CGI.parse lives in cgi/core; MRI and TruffleRuby ship it,
# but JRuby 10.1 removed it from stdlib (it directs users to the cgi gem). Load
# it if available and guard the parse op on CGI.respond_to?(:parse) so a runtime
# without it still reports the other four ops instead of aborting.
require "cgi/escape"
begin
  require "cgi/core"
rescue LoadError
  # parse unavailable in this runtime's stdlib; its parse row is omitted.
end
require_relative "_harness"

HAS_PARSE = CGI.respond_to?(:parse)

# Fixed representative inputs, byte-identical to go/main.go.
URL_INPUT   = "Café René & Co. — 100% \"cœur\", <tag>? a=1 b=2 (x/y) #frag"
HTML_INPUT  = "<a href=\"/u?id=42&t=1\">Tom & Jerry's \"Big\" <Adventure></a> 5 > 3 & 2 < 4"
HTML_ESCAPED = "Tom &amp; Jerry&#39;s &quot;Big&quot; &lt;Adventure&gt; &#8212; caf&#233; &#x263A; &apos;end&apos;"
QUERY_INPUT = "name=John+Doe&email=j%40x.com&tags[]=a&tags[]=b&tags[]=c&q=caf%C3%A9+%26+co&page=2&sort=desc&empty="

ESC_URL = CGI.escape(URL_INPUT)

# Canonicalise CGI.parse output the same way go/serialize.go does: sorted keys,
# 0x1f between a key and each of its values, 0x1e terminating each key group.
def parse_string(h)
  out = +""
  h.keys.sort.each do |k|
    out << k
    h[k].each { |v| out << "\x1f" << v }
    out << "\x1e"
  end
  out
end

if ARGV[0] == "check"
  check("escape",       CGI.escape(URL_INPUT))
  check("unescape",     CGI.unescape(ESC_URL))
  check("escapeHTML",   CGI.escapeHTML(HTML_INPUT))
  check("unescapeHTML", CGI.unescapeHTML(HTML_ESCAPED))
  check("parse",        parse_string(CGI.parse(QUERY_INPUT))) if HAS_PARSE
  exit
end

bench("escape", 5000)       { CGI.escape(URL_INPUT) }
bench("unescape", 5000)     { CGI.unescape(ESC_URL) }
bench("escapeHTML", 5000)   { CGI.escapeHTML(HTML_INPUT) }
bench("unescapeHTML", 5000) { CGI.unescapeHTML(HTML_ESCAPED) }
bench("parse", 2000)        { CGI.parse(QUERY_INPUT) } if HAS_PARSE
