package models

import (
	"github.com/goravel/framework/database/orm"
)

// EntryData represents form field data submitted with a workflow entry 工作流表单字段数据模型，记录发起流程时填写的表单字段及其值
type EntryData struct {
	orm.Model
	// EntryID is the ID of the workflow entry this data belongs to 所属流程实例ID
	EntryID int `gorm:"column:entry_id;not null;default:0" form:"entry_id" json:"entry_id"`
	// FlowID is the ID of the flow definition 所属流程定义ID
	FlowID int `gorm:"column:flow_id;not null;default:0" form:"flow_id" json:"flow_id"`
	// FieldName is the name of the form field 表单字段名
	FieldName string `gorm:"column:field_name;not null;default:''" form:"field_name" json:"field_name"`
	// FieldValue is the value submitted for this field 表单字段值
	FieldValue string `gorm:"column:field_value" json:"field_value" form:"field_value" json:"field_value"`
	// FieldRemark is an optional remark/description for this field 字段备注说明
	FieldRemark string `gorm:"column:field_remark;not null;default:''" form:"field_remark" json:"field_remark"`
}

// TableName returns the database table name for EntryData 返回数据库表名
func (e *EntryData) TableName() string {
	return "entrydatas"
}
