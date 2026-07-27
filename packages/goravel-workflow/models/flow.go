package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

// Flow 流程定义模型，代表一个完整的审批流程（如请假申请、报销审批等）
// Flow represents a workflow definition 流程定义模型
type Flow struct {
	orm.Model
	FlowNo      string       `gorm:"column:flow_no;not null" json:"flow_no" form:"flow_no"`          // 流程编号 Flow number
	FlowName    string       `gorm:"column:flow_name;not null;default:''" json:"flow_name" form:"flow_name"` // 流程显示名称 Flow display name
	TemplateID  int          `gorm:"column:template_id;not null;default:0" json:"template_id" form:"template_id"` // 关联的表单模板ID  Associated form template ID
	Flowchart   string       `gorm:"column:flowchart" json:"flowchart" form:"flowchart"`                  // 流程图数据（JSON格式） Flowchart data (JSON format)
	Jsplumb     string       `gorm:"column:jsplumb;comment:'jsplumb流程图数据'" json:"jsplumb" form:"jsplumb"` // jsPlumb流程图JSON数据 jsPlumb flowchart visualization data
	TypeID      int          `gorm:"column:type_id;not null;default:0" json:"type_id" form:"type_id"`     // 流程分类ID Flow category/type ID
	IsPublish   bool         `gorm:"column:is_publish;not null;default:0" json:"is_publish" form:"is_publish"` // 是否已发布 Whether the flow is published
	IsShow      bool         `gorm:"column:is_show;not null;default:1" json:"is_show" form:"is_show"`     // 是否在前端显示 Whether to show in the frontend
	Processes   []Process    `gorm:"foreignKey:FlowID"`     // HasMany Process 一对多关联：流程包含的审批步骤
	ProcessVars []ProcessVar `gorm:"foreignKey:FlowID"`     // HasMany ProcessVar 一对多关联：流程的条件变量
	Template    Template     `gorm:"foreignKey:TemplateID"` // BelongsTo Template 多对一关联：所属的表单模板
	Flowtype    Flowtype     `gorm:"foreignKey:TypeID"`     // BelongsTo FlowType 多对一关联：所属的流程分类
}

// LoadProcesses 预加载当前流程关联的所有审批步骤（Processes）
// LoadProcesses preloads the associated Processes for a Flow.
func (f *Flow) LoadProcesses(db *gorm.DB) error {
	return db.Preload("Processes").Find(f).Error
}

// LoadProcessVars 预加载当前流程关联的所有条件变量（ProcessVars）
// LoadProcessVars preloads the associated ProcessVars for a Flow.
func (f *Flow) LoadProcessVars(db *gorm.DB) error {
	return db.Preload("ProcessVars").Find(f).Error
}

// LoadTemplate 预加载当前流程关联的表单模板（Template）
// LoadTemplate preloads the associated Template for a Flow.
func (f *Flow) LoadTemplate(db *gorm.DB) error {
	return db.Preload("Template").Find(f).Error
}

// LoadFlowType 预加载当前流程关联的流程分类（FlowType）
// LoadFlowType preloads the associated FlowType for a Flow.
func (f *Flow) LoadFlowType(db *gorm.DB) error {
	return db.Preload("FlowType").Find(f).Error
}
