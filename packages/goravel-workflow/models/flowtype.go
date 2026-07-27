package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

// Flowtype represents a workflow type that groups related flows together
// 流程类型模型，用于将相关流程归类分组
type Flowtype struct {
	orm.Model
	// TypeName is the name of this flow type
	// 流程类型名称，用于在前端展示和区分不同类别的流程
	TypeName string `gorm:"column:type_name;not null;default:''" json:"type_name"`
	// Flows contains the associated Flows under this Flowtype, not persisted to DB
	// 该流程类型下关联的流程列表，通过 Preload 加载，不会持久化到数据库
	Flows []Flow `gorm:"-"`
}

// LoadFlowsForType preloads the associated Flows for a FlowType.
// 预加载该流程类型下的所有关联流程
func (ft *Flowtype) LoadFlowsForType(db *gorm.DB) error {
	return db.Preload("Flows").Find(ft).Error
}
