package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

// UserRequest 用户表单请求验证结构体
// 用于验证用户相关的请求参数，包含姓名和手机号字段
type UserRequest struct {
	// Name 用户姓名
	Name   string `form:"name" json:"name"`
	// Mobile 用户手机号
	Mobile string `form:"mobile" json:"mobile"`
}

// Authorize 授权验证，判断当前请求是否有权限执行
// 返回 nil 表示所有请求均放行，如需权限控制可在此处添加逻辑
func (r *UserRequest) Authorize(ctx http.Context) error {
	return nil
}

// Rules 定义字段验证规则
// 返回字段名到规则字符串的映射，name 和 mobile 均为必填字段
func (r *UserRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   "required", // name 字段为必填
		"mobile": "required", // mobile 字段为必填
	}
}

// Messages 定义验证失败时的自定义错误消息
// 返回字段规则到错误消息的映射，用于替代默认的英文错误提示
func (r *UserRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":   "用户名不能为空", // name 必填校验失败时的中文提示
		"mobile.required": "手机号不能为空", // mobile 必填校验失败时的中文提示
	}
}

// Attributes 定义字段的中文属性名
// 用于在错误消息中替换字段名占位符，使提示信息更友好
func (r *UserRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   "用户名", // name 字段的中文名称
		"mobile": "手机号", // mobile 字段的中文名称
	}
}

// PrepareForValidation 验证前的数据预处理钩子
// 可在正式验证之前对请求数据进行清洗或加工，当前无需额外处理
func (r *UserRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
