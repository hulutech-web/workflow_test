package workflow

import (
	"encoding/json"
	"fmt"
	"sort"
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
	var parsedBranches []conditionBranch

	for _, fl := range flowlinksJSON {
		if fl.Expression == "" {
			return fmt.Errorf("步骤中有未设置条件的条件分支（sort=%d），请完善或删除", fl.Sort)
		}
		if fl.Expression == "1" {
			hasCatchAll = true
			parsedBranches = append(parsedBranches, conditionBranch{
				id:         fl.ID,
				sort:       fl.Sort,
				isCatchAll: true,
			})
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

		parsedBranches = append(parsedBranches, conditionBranch{
			id:         fl.ID,
			sort:       fl.Sort,
			conditions: conditions,
		})
	}

	// 跨分支完整性检测：重叠 + 缺口 + 等式冲突
	if err := validateBranchCompleteness(parsedBranches); err != nil {
		return err
	}

	_ = hasCatchAll // non-blocking: missing catch-all is a warning, not an error
	return nil
}

// conditionBranch 一个已解析的条件分支
type conditionBranch struct {
	id         int
	sort       int
	isCatchAll bool
	conditions []common.ProcessCondition
}

// validateBranchCompleteness ensures that for any possible input, exactly one branch matches.
// It detects:
//   1. Overlaps — a value could match more than one branch (ambiguous routing)
//   2. Gaps — a value matches no branch (unreachable fallback)
//
// For pure numeric-range-style conditions (all branches on the same single field with
// comparison operators), it checks range coverage end-to-end.
// For mixed-field or equality-based conditions, it requires an explicit catch-all
// branch (Expression="1") and checks for contradictory equalities (same field
// "=" but different values across branches with no other distinguishing field).
func validateBranchCompleteness(branches []conditionBranch) error {
	if len(branches) == 0 {
		return nil
	}

	// Separate catch-all from regular branches
	regular := make([]conditionBranch, 0, len(branches))
	hasCatchAll := false
	for _, b := range branches {
		if b.isCatchAll {
			hasCatchAll = true
		} else {
			regular = append(regular, b)
		}
	}

	if len(regular) == 0 {
		return nil
	}

	// ---- 1. Pure numeric-range single-field overlap + gap detection ----
	if err := validateRangeCoverage(regular); err != nil {
		return err
	}

	// ---- 2. Equality conflict detection (mixed-field scenarios) ----
	if err := validateEqualityConflicts(regular); err != nil {
		return err
	}

	// ---- 3. If range-detection couldn't handle all branches, require catch-all ----
	allRange := allBranchesRangeOnly(regular)
	if !allRange && !hasCatchAll {
		return fmt.Errorf("条件不完整：存在非纯数值范围的分支（如等于、包含等），且未设置兜底分支（Expression=\"1\"）。请添加兜底分支确保所有情况都有匹配")
	}

	return nil
}

// allBranchesRangeOnly returns true if every regular branch can be reduced to a
// numeric range (single comparison or AND-chain of comparisons on the same field).
func allBranchesRangeOnly(branches []conditionBranch) bool {
	for _, b := range branches {
		if !conditionBranchIsRangeBased(b) {
			return false
		}
	}
	return true
}

// conditionBranchIsRangeBased returns true if all conditions in the branch
// are numeric comparisons on the same field, making it representable as a single range.
func conditionBranchIsRangeBased(b conditionBranch) bool {
	if len(b.conditions) == 0 {
		return false
	}
	// All conditions must be numeric comparisons on the same field
	var firstField string
	for _, c := range b.conditions {
		_, _, ok := isRangeOp(c)
		if !ok {
			return false
		}
		if firstField == "" {
			firstField = c.Field
		} else if c.Field != firstField {
			return false
		}
	}
	return true
}

// isRangeOp checks if the condition uses a numeric comparison operator.
// Returns (lo, hi, isRange) where isRange is false for equality/like/in/notin operators.
func isRangeOp(c common.ProcessCondition) (lo, hi *float64, ok bool) {
	switch c.Operator {
	case ">", ">=":
		v, err := strconv.ParseFloat(c.Value, 64)
		if err != nil {
			return nil, nil, false
		}
		return &v, nil, true
	case "<", "<=":
		v, err := strconv.ParseFloat(c.Value, 64)
		if err != nil {
			return nil, nil, false
		}
		return nil, &v, true
	case "between":
		loVal, err1 := strconv.ParseFloat(c.Value, 64)
		hiVal, err2 := strconv.ParseFloat(c.ExtraValue, 64)
		if err1 != nil || err2 != nil {
			return nil, nil, false
		}
		return &loVal, &hiVal, true
	default:
		return nil, nil, false
	}
}

// numRange represents a numeric interval for one branch.
type numRange struct {
	field  string
	lo     float64
	hi     float64
	loIncl bool // true if lower bound is inclusive (>=)
	hiIncl bool // true if upper bound is inclusive (<=)
	sort   int
	hasNonRangeCond bool // true if branch has additional non-range constraints (e.g. AND gender='male')
}

// extractPrimaryRange 从多条件分支中提取主数值字段的范围。
// 例如 money >= 1000 AND money < 5000 → lo=1000, hi=5000, loIncl=true, hiIncl=false
// 只提取同一字段的 AND 连接的范围条件
func extractPrimaryRange(b conditionBranch) (numRange, bool) {
	if len(b.conditions) == 0 {
		return numRange{}, false
	}

	// 找出所有范围条件中使用最多的字段
	fieldCount := make(map[string]int)
	for _, c := range b.conditions {
		_, _, ok := isRangeOp(c)
		if ok {
			fieldCount[c.Field]++
		}
	}
	if len(fieldCount) == 0 {
		return numRange{}, false
	}

	// 取第一个有范围条件的字段作为主字段
	var primaryField string
	primaryField = b.conditions[0].Field
	if _, _, ok := isRangeOp(b.conditions[0]); ok {
		primaryField = b.conditions[0].Field
	} else {
		for _, c := range b.conditions {
			if _, _, ok := isRangeOp(c); ok {
				primaryField = c.Field
				break
			}
		}
	}

	r := numRange{
		field:  primaryField,
		sort:   b.sort,
		lo:     -1e308,
		hi:     1e308,
		loIncl: true,
		hiIncl: true,
	}

	// 合并所有同字段 AND 连接的范围条件
	for _, c := range b.conditions {
		if c.Field != primaryField {
			// 跨字段条件（如 AND gender = 'male'），不参与范围计算，只标记有额外约束
			r.hasNonRangeCond = true
			continue
		}
		lo, hi, ok := isRangeOp(c)
		if !ok {
			continue
		}
		if lo != nil {
			newLo := *lo
			newLoIncl := c.Operator == ">=" || c.Operator == "between"
			if newLo > r.lo || (newLo == r.lo && newLoIncl && !r.loIncl) {
				r.lo = newLo
				r.loIncl = newLoIncl
			}
		}
		if hi != nil {
			newHi := *hi
			newHiIncl := c.Operator == "<=" || c.Operator == "between"
			if newHi < r.hi || (newHi == r.hi && newHiIncl && !r.hiIncl) {
				r.hi = newHi
				r.hiIncl = newHiIncl
			}
		}
	}

	if r.lo >= r.hi {
		return numRange{}, false // invalid or empty range
	}
	return r, true
}

// validateRangeCoverage detects overlaps and gaps where all branches are numeric
// single-field comparisons.
func validateRangeCoverage(branches []conditionBranch) error {
	// Group branches by field
	byField := make(map[string][]numRange)
	for _, b := range branches {
		r, ok := extractPrimaryRange(b)
		if !ok {
			continue
		}
		byField[r.field] = append(byField[r.field], r)
	}

	for field, ranges := range byField {
		if len(ranges) < 2 {
			continue
		}

		// 1. Detect pairwise overlaps
		for i := 0; i < len(ranges); i++ {
			for j := i + 1; j < len(ranges); j++ {
				a, b := ranges[i], ranges[j]

				overlapLo := max(a.lo, b.lo)
				overlapHi := min(a.hi, b.hi)

				loIncl := (a.lo == overlapLo && a.loIncl) || (b.lo == overlapLo && b.loIncl)
				hiIncl := (a.hi == overlapHi && a.hiIncl) || (b.hi == overlapHi && b.hiIncl)

				if overlapLo < overlapHi {
					return fmt.Errorf(
						"条件重叠：字段 \"%s\" 的分支[sort=%d] (%s %.4g, %.4g %s) 与分支[sort=%d] (%s %.4g, %.4g %s) 存在重叠区间 [%.4g, %.4g]",
						field,
						a.sort, loSymbol(a.loIncl), a.lo, a.hi, hiSymbol(a.hiIncl),
						b.sort, loSymbol(b.loIncl), b.lo, b.hi, hiSymbol(b.hiIncl),
						overlapLo, overlapHi,
					)
				}
				// 表示一个区间包含边界而另一个不包含（如 (1000, 5000] vs (5000, 10000]）
				// 只有两端都包含才是真正的重叠，否则只是恰好相邻
				if overlapLo == overlapHi && loIncl && hiIncl && a.loIncl != b.loIncl && a.hiIncl != b.hiIncl {
					// 相邻但不是重叠：如 [1,2] 和 (2,3]
					// loIncl: one at 2, hiIncl: one at 2, but loIncl vs hiIncl are from different branches
					// 跳过 — 边界互斥
				} else if overlapLo == overlapHi && loIncl && hiIncl {
					return fmt.Errorf(
						"条件重叠：字段 \"%s\" 的分支[sort=%d] 与分支[sort=%d] 在边界值 %.4g 处有歧义（两个分支都包含该值）",
						field, a.sort, b.sort, overlapLo,
					)
				}
			}
		}

		// 2. Detect gaps by sorting and checking coverage
		// Sort by lower bound
		sort.Slice(ranges, func(i, j int) bool {
			if ranges[i].lo == ranges[j].lo {
				return !ranges[i].loIncl && ranges[j].loIncl // exclusive first
			}
			return ranges[i].lo < ranges[j].lo
		})

		// Check coverage from -inf to +inf
		coveredUpTo := -1e308
		coveredIncl := true // whether the coveredUpTo point is included

		for _, r := range ranges {
			if r.lo > coveredUpTo {
				// Gap: there's space between coveredUpTo and r.lo
				return fmt.Errorf(
					"条件缺口：字段 \"%s\" 在区间 (%s, %s) 处没有分支能匹配。请确保所有可能的输入值都能命中唯一的分支",
					field, formatVal(coveredUpTo), formatVal(r.lo),
				)
			}
			if r.lo == coveredUpTo && !coveredIncl && !r.loIncl {
				// Gap at boundary, e.g., (0, 10] and next is (10, ...] — value 10 has gap
				return fmt.Errorf(
					"条件缺口：字段 \"%s\" 在值 %.4g 处没有分支匹配。请在相邻分支间接好边界",
					field, coveredUpTo,
				)
			}

			// Extend coverage
			if r.hi > coveredUpTo || (r.hi == coveredUpTo && r.hiIncl && !coveredIncl) {
				coveredUpTo = r.hi
				coveredIncl = r.hiIncl
			}
		}

		if coveredUpTo < 1e307 {
			return fmt.Errorf(
				"条件缺口：字段 \"%s\" 在值 %s 以上的部分没有分支能匹配。请扩展已有分支或添加兜底分支",
				field, formatVal(coveredUpTo),
			)
		}
	}

	return nil
}

func loSymbol(incl bool) string {
	if incl {
		return "["
	}
	return "("
}

func hiSymbol(incl bool) string {
	if incl {
		return "]"
	}
	return ")"
}

// validateEqualityConflicts detects conflicts like: branch A: gender='male', branch B: gender='male'
// with no other field to differentiate them (same-value conflict).
func validateEqualityConflicts(branches []conditionBranch) error {
	type eqKey struct {
		field string
		value string
	}
	byEq := make(map[eqKey][]conditionBranch)
	for _, b := range branches {
		for _, c := range b.conditions {
			if c.Operator == "=" {
				key := eqKey{c.Field, c.Value}
				byEq[key] = append(byEq[key], b)
			}
		}
	}

	for key, bs := range byEq {
		if len(bs) > 1 {
			// Check if any pair has other fields to differentiate
			for i := 0; i < len(bs); i++ {
				for j := i + 1; j < len(bs); j++ {
					if !hasDifferentiatingField(bs[i], bs[j]) {
						return fmt.Errorf(
							"条件重叠：分支[sort=%d] 和分支[sort=%d] 都包含 \"%s = %s\" 且无其他区分字段。对于该输入，两条分支会同时命中，请增加额外条件区分",
							bs[i].sort, bs[j].sort, key.field, key.value,
						)
					}
				}
			}
		}
	}

	return nil
}

// hasDifferentiatingField checks if two branches test a field that the other doesn't,
// meaning they could target different input subspaces.
func hasDifferentiatingField(a, b conditionBranch) bool {
	aFields := make(map[string]bool)
	bFields := make(map[string]bool)
	for _, c := range a.conditions {
		aFields[c.Field] = true
	}
	for _, c := range b.conditions {
		bFields[c.Field] = true
	}
	// If there's a field present in one but not the other, they can be differentiated
	for f := range aFields {
		if !bFields[f] {
			return true
		}
	}
	for f := range bFields {
		if !aFields[f] {
			return true
		}
	}
	return false
}

func formatVal(v float64) string {
	if v <= -1e307 {
		return "-∞"
	}
	if v >= 1e307 {
		return "+∞"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
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
			if err := json.Unmarshal([]byte(UnescapeExpressionJSON(fl.Expression)), &conditions); err != nil {
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

// UnescapeExpressionJSON fixes double-escaped JSON stored in MySQL.
// Go json.Marshal escapes < and > as < and >; MySQL then doubles
// the backslash to \u003c / \u003e. This function normalizes the raw JSON
// bytes back to literal < and > so json.Unmarshal produces correct operators.
func UnescapeExpressionJSON(raw string) string {
	s := raw
	// Fix double-escaped: \u003c → \u003c (which json.Unmarshal correctly decodes to <)
	// The issue is the DB has \u003e which Go sees as literal "<" string
	s = strings.ReplaceAll(s, `>`, `>`)
	s = strings.ReplaceAll(s, `>=`, `>=`)
	s = strings.ReplaceAll(s, `<`, `<`)
	s = strings.ReplaceAll(s, `<=`, `<=`)
	// Also fix straight literal > that may come from DB
	s = strings.ReplaceAll(s, `\u003e`, `>`)
	s = strings.ReplaceAll(s, `\u003e=`, `>=`)
	s = strings.ReplaceAll(s, `\u003c`, `<`)
	s = strings.ReplaceAll(s, `\u003c=`, `<=`)
	return s
}
