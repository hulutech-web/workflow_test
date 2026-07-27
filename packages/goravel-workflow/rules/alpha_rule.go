package rules

import (
	"github.com/goravel/framework/contracts/validation"
	"strings"
)

// AlphaRule 自定义验证规则：校验字段值是否仅包含英文字母和下划线。
type AlphaRule struct {
}

// Signature 返回验证规则的名称。
// The name of the rule.
func (receiver *AlphaRule) Signature() string {
	return "alpha_rule"
}

// Passes 判断验证规则是否通过。
// 允许空值直接通过；非空值时逐字符检查，仅允许小写字母 a-z、大写字母 A-Z 以及下划线 _。
// Determine if the validation rule passes.
func (receiver *AlphaRule) Passes(data validation.Data, val any, options ...any) bool {
	// 空值直接放行，不进行校验
	if val == "" {
		return true
	}
	// 遍历每一个字符，检查是否在允许的字符范围内
	for _, ch := range []rune(val.(string)) {
		if !('a' <= ch && ch <= 'z') && !('A' <= ch && ch <= 'Z') && !(strings.Contains("_", string(ch))) {
			return false
		}
	}
	return true
}

// Message 返回验证失败时的错误消息（中文）。
// Get the validation error message.
func (receiver *AlphaRule) Message() string {
	return "字符 :attribute 必须是英文字符或者包含下划线"
}
