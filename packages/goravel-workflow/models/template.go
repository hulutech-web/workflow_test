package models

import (
	"github.com/goravel/framework/database/orm"
)

// Template represents a form template definition associated with a workflow flow
// 表单模板定义模型，与工作流流程定义关联的表单模板
type Template struct {
	orm.Model
	// TemplateName is the display name of this form template 模板名称，对应数据库中 workflow_form_template 表的模板名
	TemplateName string `gorm:"column:template_name;not null;default:''" json:"template_name"`
	// TemplateForms are the form fields that belong to this template 模板包含的表单字段列表，一对多关联 TemplateForm
	TemplateForms []TemplateForm
}
