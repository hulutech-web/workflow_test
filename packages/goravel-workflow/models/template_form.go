package models

import (
	"github.com/goravel/framework/database/orm"
	"goravel/packages/goravel-workflow/models/common"
	"gorm.io/gorm"
)

// TemplateForm represents a form template field definition
// 表单模板字段定义模型，用于定义工作流模板中的表单字段结构
type TemplateForm struct {
	orm.Model
	// Field is the English field name used as the form field key
	// 表单字段英文名，用作表单字段的标识键
	Field string `gorm:"column:field;not null;default:'';comment:'表单字段英文名'" json:"field" form:"field"`
	// FieldName is the Chinese display name of the form field
	// 表单字段中文名，用于前端页面展示
	FieldName string `gorm:"column:field_name;not null;default:'';comment:'表单字段中文名'" json:"field_name" form:"field_name"`
	// FieldType is the form field type (e.g., text, select, radio, checkbox, textarea, date)
	// 表单字段类型，如：text（文本）、select（下拉）、radio（单选）、checkbox（多选）、textarea（文本域）、date（日期）等
	FieldType string `gorm:"column:field_type;not null;default:'';comment:'表单字段类型'" json:"field_type" form:"field_type"`
	// FieldValue holds the selectable options for select, radio, and checkbox field types
	// 表单字段可选值，用于 select、radio、checkbox 类型字段的选项配置
	FieldValue common.FieldValue `gorm:"column:field_value;type:text;comment:'表单字段值，select radio checkbox用'" json:"field_value" form:"field_value"`
	// FieldDefaultValue is the default value for the form field when the form is first loaded
	// 表单字段默认值，在表单首次加载时显示
	FieldDefaultValue string `gorm:"column:field_default_value;type:text;comment:'表单字段默认值'" json:"field_default_value" form:"field_default_value"`
	// FieldRules defines the validation rules for this form field
	// 表单字段校验规则，用于动态验证用户提交的表单数据
	FieldRules common.Rule `gorm:"column:field_rules;" json:"field_rules" form:"field_rules"`
	// Sort determines the display order of this field in the form (default: 100)
	// 排序字段，决定表单字段在页面上的显示顺序，默认值为 100
	Sort int `gorm:"column:sort;not null;default:100;comment:'排序'" json:"sort" form:"sort"`
	// TemplateID is the foreign key linking this field to a specific form template
	// 模板ID，外键，关联到所属的表单模板
	TemplateID uint `gorm:"column:template_id;not null;default:0;comment:'模板ID'" json:"template_id" form:"template_id"`
	// Template is the associated form template (GORM relationship)
	// 所属表单模板，GORM 关联关系
	Template Template
}

// TableName returns the database table name for the TemplateForm model
// 返回 TemplateForm 模型对应的数据库表名
func (p *TemplateForm) TableName() string {
	return "templateforms"
}

// LoadTemplateForm loads the TemplateForm record along with its associated Template
// 加载表单模板字段记录，同时预加载关联的模板数据
func (p *TemplateForm) LoadTemplateForm(db *gorm.DB) error {
	return db.Preload("Template").Find(p).Error
}
