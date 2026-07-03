// SPDX-License-Identifier: BSD-3-Clause
package main

import (
	"sort"
	"strings"
)

// parseString canonicalises a CGI.parse result into a stable, order-independent
// string so the Go and Ruby drivers can be diffed byte-for-byte. Keys are sorted
// and each key's value list is joined verbatim: "key\x1fv1\x1fv2\x1e" per key.
// The 0x1f/0x1e separators cannot appear in the fixed inputs, so the encoding is
// unambiguous.
func parseString(m map[string][]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		for _, v := range m[k] {
			sb.WriteByte(0x1f)
			sb.WriteString(v)
		}
		sb.WriteByte(0x1e)
	}
	return sb.String()
}
