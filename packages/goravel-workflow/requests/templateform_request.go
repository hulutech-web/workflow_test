package requests

import (
	"goravel/packages/goravel-workflow/models/common"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
)

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

func (r *TemplateformRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *TemplateformRequest) Rules(ctx http.Context) map[string]any {
	return map[string]any{
		"template_id": "required",
		"field":       "required",
		"field_name":  "required",
		"field_type":  "required",
	}
}

func (r *TemplateformRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"template_id.required": "模板ID不能为空",
		"field.required":       "表单字段英文名不能为空",
		"field_name.required":  "表单字段中文名不能为空",
		"field_type.required":  "表单字段类型不能为空",
	}
}

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

func (r *TemplateformRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	if name, exist := data.Get("sort"); exist {
		data.Set("sort", cast.ToInt(name))
	}
	if val, exist := data.Get("field_rules"); exist {
		//将val转换为Rule类型
		r.FieldRules = common.Rule{}
		//	使用mapstruct将json字符串转换为Rule类型
		if err := mapstructure.Decode(val, &r.FieldRules); err != nil {
			return err
		}

	}
	return nil
}
