package requests

import (
	"github.com/goravel/framework/contracts/http"
)

// RevokeEntryRequest 撤回流程请求验证结构体
// 用于验证用户发起撤回操作时的请求参数
type RevokeEntryRequest struct {
	// EntryID 需要撤回的流程实例 ID
	EntryID int `form:"entry_id" json:"entry_id"`
}

// Authorize 授权验证，允许所有经过 JWT 认证的用户发起撤回请求
// 具体的业务权限校验（如仅允许发起人撤回）在业务层 Revoke() 方法中处理
func (r *RevokeEntryRequest) Authorize(ctx http.Context) error {
	return nil
}

// Rules 定义请求参数的验证规则
// entry_id 为必填项，且必须为正整数
func (r *RevokeEntryRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id": "required|integer|min:1",
	}
}

// Messages 定义验证失败时的自定义错误消息
func (r *RevokeEntryRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id.required": "流程ID不能为空", // 当 entry_id 未提供时的错误提示
	}
}
