package requests

import (
	"goravel/packages/goravel-workflow/models/common"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
)

// TemplateformRequest 表单模板字段请求结构体，用于验证表单模板字段的增改操作
type TemplateformRequest struct {
	ID                uint              `json:"id" form:"-"`
	Field             string            `json:"field" form:"field"`
	FieldName         string            `json:"field_name" form:"field_name"`
	FieldType         string            `json:"field_type" form:"field_type"`
	FieldValue        common.FieldValue `json:"field_value" form:"field_value"`
	FieldDefaultValue string            `json:"field_default_value" form:"field_default_value"`
	FieldRules        common.Rule       `json:"field_rules" form:"field_rules"`
	Sort              int               `json:"sort" form:"sort"`
	TemplateID        uint              `json:"template_id" form:"template_id"`
}

// Authorize 授权验证，所有请求均允许通过
func (r *TemplateformRequest) Authorize(ctx http.Context) error {
	return nil
}

// Rules 定义请求参数的验证规则
func (r *TemplateformRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"template_id": "required", // 模板ID必填
		"field":       "required", // 表单字段英文名必填
		"field_name":  "required", // 表单字段中文名必填
		"field_type":  "required", // 表单字段类型必填
	}
}

// Messages 定义验证失败时的自定义错误消息
func (r *TemplateformRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"template_id.required": "模板ID不能为空",
		"field.required":       "表单字段英文名不能为空",
		"field_name.required":  "表单字段中文名不能为空",
		"field_type.required":  "表单字段类型不能为空",
	}
}

// Attributes 定义请求参数的属性名称，用于错误消息中的字段标识
func (r *TemplateformRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"template_id":         "模板ID",
		"field":               "表单字段英文名",
		"field_name":          "表单字段中文名",
		"field_type":          "表单字段类型",
		"field_value":         "表单字段值",
		"field_default_value": "表单字段默认值",
		"field_rules":         "表单字段规则",
		"sort":                "排序",
	}
}

// PrepareForValidation 验证前的数据预处理，将请求数据转换为目标类型
func (r *TemplateformRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	// 将 sort 字段从字符串转换为整数类型
	if name, exist := data.Get("sort"); exist {
		data.Set("sort", cast.ToInt(name))
	}
	// 将 field_rules 字段从原始数据转换为 Rule 结构体类型
	if val, exist := data.Get("field_rules"); exist {
		//将val转换为Rule类型
		// Chinese translation: 将原始值转换为 Rule 类型
		r.FieldRules = common.Rule{}
		//	使用mapstruct将json字符串转换为Rule类型
		// Chinese translation: 使用 mapstructure 将 JSON 字符串解析为 Rule 结构体
		if err := mapstructure.Decode(val, &r.FieldRules); err != nil {
			return err
		}

	}
	return nil
}
