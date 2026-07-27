package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

// ProcessVar 流程步骤变量模型，用于条件分支路由时的表达式字段映射
// ProcessVar represents a process step variable used for expression field mapping in conditional routing 流程步骤变量模型
type ProcessVar struct {
	orm.Model
	// ProcessID 关联的流程步骤ID
	ProcessID       int     `gorm:"column:process_id;not null" json:"process_id"`
	// FlowID 关联的流程定义ID 流程id
	FlowID          int     `gorm:"column:flow_id;not null;comment:'流程id'" json:"flow_id"`
	// ExpressionField 条件表达式字段名称，对应 EntryData 中的字段名，用于条件分支匹配 条件表达式字段名称
	ExpressionField string  `gorm:"column:expression_field;not null;comment:'条件表达式字段名称'" json:"expression_field"`
	// Process 关联的 Process 模型，外键为 ProcessID
	Process         Process `gorm:"foreignKey:ProcessID;references:ID"`
}

// TableName 返回 ProcessVar 对应的数据库表名 "processvars"
func (e *ProcessVar) TableName() string {
	return "processvars"
}

// LoadProcess 预加载关联的 Process 数据并填充到当前 ProcessVar
func (p *ProcessVar) LoadProcess(db *gorm.DB) error {
	return db.Preload("Process").Find(p).Error
}
