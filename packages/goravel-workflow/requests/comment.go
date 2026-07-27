package requests

import (
	"github.com/goravel/framework/contracts/http"
)

// CommentRequest 评论请求验证结构体
// 用于验证添加评论接口的请求参数
type CommentRequest struct {
	EntryID int    `form:"entry_id" json:"entry_id"` // 流程实例ID
	ProcID  int    `form:"proc_id" json:"proc_id"`   // 审批任务ID（可选，关联到具体审批节点）
	Content string `form:"content" json:"content"`   // 评论内容
}

// Authorize 授权验证
// 返回 nil 表示所有用户均可发表评论，无需额外的权限校验
func (r *CommentRequest) Authorize(ctx http.Context) error {
	return nil
}

// Rules 字段验证规则
// 定义请求参数必须满足的校验规则：
//   - entry_id: 必填，整数，最小值 1
//   - content:  必填，字符串，最大长度 500
func (r *CommentRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id": "required|integer|min:1",
		"content":  "required|string|max:500",
	}
}

// Messages 自定义验证错误消息
// 当验证规则未通过时，返回对应字段的中文提示信息
func (r *CommentRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.max": "评论内容不能超过500字",
	}
}
