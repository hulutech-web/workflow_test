package workflow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"goravel/packages/goravel-workflow/controllers/common"
	"goravel/packages/goravel-workflow/models"
)

// ValidateConditionCoherence checks a single flowlink's condition group for logical
// contradictions that make the expression unsatisfiable.
//
// Examples of contradictions detected:
//   - amount > 1000 AND amount < 500  (lower > upper)
//   - amount = 5 AND amount != 5       (direct conflict)
//   - between 100 AND 50               (lower > upper)
func ValidateConditionCoherence(conditions []common.ProcessCondition) error {
	if len(conditions) == 0 {
		return nil
	}

	desc := describeConditions(conditions)

	// Check adjacent same-field AND-connected pairs for range contradictions
	for i := 0; i < len(conditions)-1; i++ {
		a, b := conditions[i], conditions[i+1]
		if a.Field != b.Field {
			continue
		}
		if strings.TrimSpace(strings.ToUpper(a.Extra)) != "AND" {
			continue
		}

		// = X AND = Y with X != Y
		if a.Operator == "=" && b.Operator == "=" && a.Value != b.Value {
			return fmt.Errorf("条件矛盾：字段 \"%s\" 无法同时等于 \"%s\" 且等于 \"%s\"\n  完整条件：%s",
				a.Field, a.Value, b.Value, desc)
		}
		// = X AND != X
		if a.Operator == "=" && b.Operator == "!=" && a.Value == b.Value {
			return fmt.Errorf("条件矛盾：字段 \"%s\" 无法同时等于 \"%s\" 且不等于 \"%s\"\n  完整条件：%s",
				a.Field, a.Value, b.Value, desc)
		}
		if a.Operator == "!=" && b.Operator == "=" && a.Value == b.Value {
			return fmt.Errorf("条件矛盾：字段 \"%s\" 无法同时不等于 \"%s\" 且等于 \"%s\"\n  完整条件：%s",
				a.Field, a.Value, b.Value, desc)
		}

		// Numeric range contradiction: lower > X AND upper < Y where X >= Y
		if isLTOp(a.Operator) && isGTOp(b.Operator) {
			_ = checkNumericRangeContradiction(b, a, desc) // swapped: b is lower, a is upper
		} else if isGTOp(a.Operator) && isLTOp(b.Operator) {
			if err := checkNumericRangeContradiction(a, b, desc); err != nil {
				return err
			}
		}
	}

	// Check between with inverted bounds
	for _, c := range conditions {
		if strings.ToLower(c.Operator) == "between" && c.ExtraValue != "" {
			lo, e1 := strconv.ParseFloat(c.Value, 64)
			hi, e2 := strconv.ParseFloat(c.ExtraValue, 64)
			if e1 == nil && e2 == nil && lo > hi {
				return fmt.Errorf("条件矛盾：字段 \"%s\" 的 between 范围无效，下限 %s 大于上限 %s\n  完整条件：%s",
					c.Field, c.Value, c.ExtraValue, desc)
			}
		}
	}

	return nil
}

func checkNumericRangeContradiction(lower, upper common.ProcessCondition, desc string) error {
	lo, e1 := strconv.ParseFloat(lower.Value, 64)
	hi, e2 := strconv.ParseFloat(upper.Value, 64)
	if e1 != nil || e2 != nil {
		return nil // non-numeric, can't check range
	}

	loInclusive := lower.Operator == ">="
	hiInclusive := upper.Operator == "<="

	if lo > hi || (lo == hi && !(loInclusive && hiInclusive)) {
		return fmt.Errorf(
			"条件矛盾：字段 \"%s\" 无法同时满足 %s %s 且 %s %s（下限 %.4g 大于上限 %.4g）\n  完整条件：%s",
			lower.Field, lower.Operator, lower.Value, upper.Operator, upper.Value, lo, hi, desc)
	}
	return nil
}

func isGTOp(op string) bool  { return op == ">" || op == ">=" }
func isLTOp(op string) bool  { return op == "<" || op == "<=" }

func describeConditions(conditions []common.ProcessCondition) string {
	parts := make([]string, 0, len(conditions))
	for i, c := range conditions {
		if strings.ToLower(c.Operator) == "between" && c.ExtraValue != "" {
			parts = append(parts, fmt.Sprintf("%s BETWEEN %s～%s", c.Field, c.Value, c.ExtraValue))
		} else {
			parts = append(parts, fmt.Sprintf("%s %s %s", c.Field, c.Operator, c.Value))
		}
		if i < len(conditions)-1 && c.Extra != "" {
			parts = append(parts, strings.TrimSpace(c.Extra))
		}
	}
	return strings.Join(parts, " ")
}

// ConditionFlowlinkEntry is a flattened representation of a Flowlink for validation.
type ConditionFlowlinkEntry struct {
	ID         int
	Expression string
	Sort       int
}

// ValidateConditionFlowlinks checks all Condition flowlinks for a given step.
// Returns nil if all are valid, or an error describing the first problem found.
func ValidateConditionFlowlinks(flowlinksJSON []ConditionFlowlinkEntry) error {
	hasCatchAll := false
	for _, fl := range flowlinksJSON {
		if fl.Expression == "" {
			return fmt.Errorf("步骤中有未设置条件的条件分支（sort=%d），请完善或删除", fl.Sort)
		}
		if fl.Expression == "1" {
			hasCatchAll = true
			continue
		}

		var conditions []common.ProcessCondition
		if err := json.Unmarshal([]byte(fl.Expression), &conditions); err != nil {
			return fmt.Errorf("条件分支（sort=%d）的表达式格式错误，无法解析为 JSON", fl.Sort)
		}
		if len(conditions) == 0 {
			return fmt.Errorf("条件分支（sort=%d）的表达式为空数组", fl.Sort)
		}
		if err := ValidateConditionCoherence(conditions); err != nil {
			return fmt.Errorf("条件分支（sort=%d）：%w", fl.Sort, err)
		}
	}
	_ = hasCatchAll // non-blocking: missing catch-all is a warning, not an error
	return nil
}

// HasCatchAllBranch checks whether the condition flowlinks include an Expression="1" fallback.
func HasCatchAllBranch(expressions []string) bool {
	for _, e := range expressions {
		if e == "1" {
			return true
		}
	}
	return false
}

// FormatConditionError builds a detailed error message explaining why no condition branch matched.
func FormatConditionError(field string, fieldValue string, flowlinks []models.Flowlink) string {
	var b strings.Builder
	b.WriteString("条件分支评估失败：\n")
	b.WriteString(fmt.Sprintf("  评估字段：%s\n", field))
	if fieldValue != "" {
		b.WriteString(fmt.Sprintf("  当前值：%s\n", fieldValue))
	}
	b.WriteString("  已检查的条件：\n")

	for _, fl := range flowlinks {
		nextLabel := fl.NextProcess.ProcessName
		if nextLabel == "" {
			nextLabel = fmt.Sprintf("步骤%d", fl.Sort)
		}

		switch {
		case fl.Expression == "":
			b.WriteString(fmt.Sprintf("    分支%d (→%s): 未设置表达式 → 已跳过\n", fl.Sort, nextLabel))
		case fl.Expression == "1":
			b.WriteString(fmt.Sprintf("    分支%d (→%s): 兜底分支（无条件匹配）→ 理论上应命中，请检查 sort 排序\n", fl.Sort, nextLabel))
		default:
			var conditions []common.ProcessCondition
			if err := json.Unmarshal([]byte(fl.Expression), &conditions); err != nil {
				b.WriteString(fmt.Sprintf("    分支%d (→%s): 表达式格式错误 → 已跳过\n", fl.Sort, nextLabel))
				continue
			}
			condDesc := describeConditions(conditions)
			b.WriteString(fmt.Sprintf("    分支%d (→%s): %s → 不满足\n", fl.Sort, nextLabel, condDesc))
		}
	}

	b.WriteString("  结果：未找到匹配的条件分支，请检查条件配置或添加兜底分支（Expression=\"1\"）")
	return b.String()
}
