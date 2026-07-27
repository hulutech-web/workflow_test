package requests

import (
	"github.com/goravel/framework/contracts/http"
)

// AddSignRequest 加签请求结构体
// 用于处理审批流程中的加签操作，支持在当前审批节点之前或之后增加审批人
type AddSignRequest struct {
	// EntryID 流程实例ID
	EntryID int `form:"entry_id" json:"entry_id"`
	// ProcessID 当前审批步骤ID
	ProcessID int `form:"process_id" json:"process_id"`
	// SignEmpID 被加签的员工ID
	SignEmpID int `form:"sign_emp_id" json:"sign_emp_id"`
	// SignType 加签类型："before" 表示前加签（在当前审批人之前审批），"after" 表示后加签（在当前审批人之后审批）
	SignType string `form:"sign_type" json:"sign_type"` // "before" or "after"（前加签或后加签）
}

// Authorize 授权验证方法
// 验证当前请求是否具有执行加签操作的权限，此处不做额外限制，直接放行
func (r *AddSignRequest) Authorize(ctx http.Context) error {
	return nil
}

// Rules 验证规则方法
// 定义加签请求各字段的验证规则，包括必填、整数类型及最小值约束
func (r *AddSignRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id":    "required|integer|min:1",  // 流程实例ID：必填，整数，最小值为1
		"process_id":  "required|integer|min:1",  // 审批步骤ID：必填，整数，最小值为1
		"sign_emp_id": "required|integer|min:1",  // 被加签员工ID：必填，整数，最小值为1
		"sign_type":   "required|in:before,after", // 加签类型：必填，仅允许 before 或 after
	}
}

// Messages 自定义错误消息方法
// 为验证规则定义中文错误提示信息，覆盖框架默认的英文错误消息
func (r *AddSignRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id.required":   "流程ID不能为空",
		"sign_type.in":        "加签类型必须为 before 或 after",
	}
}
