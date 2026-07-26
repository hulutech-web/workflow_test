package controllers

import "strings"

// fixJSONUnicodeEscapes reverses Go's json.Marshal HTML escaping so that
// < and > look like literal < > inside the stored JSON string. Without this
// MySQL doubly-escapes the backslashes, leaving the runtime evaluator with
// > / < literals instead of real operators.
func fixJSONUnicodeEscapes(b []byte) []byte {
	s := string(b)
	s = strings.ReplaceAll(s, `>`, `>`)
	s = strings.ReplaceAll(s, `>=`, `>=`)
	s = strings.ReplaceAll(s, `<`, `<`)
	s = strings.ReplaceAll(s, `<=`, `<=`)
	return []byte(s)
}
