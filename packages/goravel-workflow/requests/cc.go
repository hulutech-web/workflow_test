package requests

import (
	"github.com/goravel/framework/contracts/http"
)

// CcRequest 抄送（CC）请求结构体
// 用于验证抄送列表查询接口的请求参数
type CcRequest struct {
	// EntryID 流程实例ID，用于查询该流程相关的抄送记录
	EntryID int `form:"entry_id" json:"entry_id"`
}

// Authorize 授权验证方法
// 所有用户均可发起抄送查询请求，无需额外权限校验
func (r *CcRequest) Authorize(ctx http.Context) error {
	return nil
}

// Rules 请求参数验证规则
// 定义抄送查询接口的字段验证规则
func (r *CcRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		// entry_id 为必填字段，类型为整数，最小值为1
		"entry_id": "required|integer|min:1",
	}
}

// Messages 自定义验证错误消息
// 为验证规则提供中文错误提示信息
func (r *CcRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		// entry_id.required 对应 entry_id 字段为必填时的错误提示
		"entry_id.required": "流程ID不能为空",
	}
}
