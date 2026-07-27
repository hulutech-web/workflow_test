package requests

import (
	"github.com/goravel/framework/contracts/http"
)

// TransferProcRequest 审批转交请求结构体
// 用于将当前审批任务转交给其他员工处理
type TransferProcRequest struct {
	EntryID      int `form:"entry_id" json:"entry_id"`            // 流程实例ID
	ProcID       int `form:"proc_id" json:"proc_id"`              // 审批任务ID
	TargetEmpID  int `form:"target_emp_id" json:"target_emp_id"`  // 目标员工ID（转交接收人）
}

// Authorize 授权验证
// 验证当前用户是否有权限执行转交操作，默认返回 nil 表示通过认证
func (r *TransferProcRequest) Authorize(ctx http.Context) error {
	return nil
}

// Rules 字段验证规则
// 定义转交请求中各字段的验证规则
func (r *TransferProcRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id":      "required|integer|min:1",  // 流程实例ID：必填、整数、最小值1
		"proc_id":       "required|integer|min:1",   // 审批任务ID：必填、整数、最小值1
		"target_emp_id": "required|integer|min:1",   // 目标员工ID：必填、整数、最小值1
	}
}

// Messages 自定义错误消息
// 定义验证失败时返回的中文错误提示信息
func (r *TransferProcRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id.required":      "流程ID不能为空",
		"target_emp_id.required": "目标员工ID不能为空",
	}
}
