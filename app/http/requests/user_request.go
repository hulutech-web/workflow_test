package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserRequest struct {
	Realname     string `form:"realname" json:"realname"`
	Phone        string `form:"phone" json:"phone"`
	Username     string `form:"username" json:"username"`
	Password     string `form:"password" json:"password"`
	Sex          string `form:"sex" json:"sex"`
	Status       string `form:"status" json:"status"`
	Remark       string `form:"remark" json:"remark"`
	Avatar       string `form:"avatar" json:"avatar"`
	IDCardNumber string `form:"id_card_number" json:"id_card_number"`
	RoleID       uint   `form:"role_id" json:"role_id"`
}

func (r *UserRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserRequest) Filters(ctx http.Context) map[string]any {
	return map[string]any{}
}

func (r *UserRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"realname": "required",
		"phone":    "required",
		"username": "required",
		"password": "required",
		"sex":      "required",
		"role_id":  "required",
	}
}

func (r *UserRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"reqlname.required": "真实姓名不为空",
		"phone.required":    "手机号不为空",
		"username":          "用户名不为空",
		"password":          "密码不为空",
		"sex":               "性别不为空",
		"role_id":           "请配置角色",
	}
}

func (r *UserRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{}
}

func (r *UserRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
