package controllers

import "strings"

// fixJSONUnicodeEscapes reverses Go's json.Marshal HTML escaping so that
// < and > look like literal < > inside the stored JSON string. Without this
// MySQL doubly-escapes the backslashes, leaving the runtime evaluator with
// > / < literals instead of real operators.
// fixJSONUnicodeEscapes 反转 Go 的 json.Marshal 产生的 HTML 转义，
// 使得 < 和 > 在存储的 JSON 字符串中保持为字面量 < >。
// 如果不做此处理，MySQL 会二次转义反斜杠，导致运行时求值器拿到的
// 是 > / < 字面量而非真正的运算符。
func fixJSONUnicodeEscapes(b []byte) []byte {
	s := string(b)

	// 还原 > 被转义后的 Unicode 转义序列 > 为字面量字符 >
	s = strings.ReplaceAll(s, `>`, `>`)
	// 还原 >= 组合
	s = strings.ReplaceAll(s, `>=`, `>=`)
	// 还原 < 被转义后的 Unicode 转义序列 < 为字面量字符 <
	s = strings.ReplaceAll(s, `<`, `<`)
	// 还原 <= 组合
	s = strings.ReplaceAll(s, `<=`, `<=`)

	return []byte(s)
}
