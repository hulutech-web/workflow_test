package models

import (
	"github.com/goravel/framework/database/orm"
)

// CcRecord represents a CC (carbon copy) record created after an approval step completes 抄送记录模型，审批步骤完成后创建的抄送记录
type CcRecord struct {
	// Embedded ORM model providing ID, CreatedAt, UpdatedAt, DeletedAt 内嵌ORM基础模型，提供ID、创建时间、更新时间、删除时间字段
	orm.Model
	// EntryID is the ID of the associated workflow entry 关联的流程实例ID
	EntryID uint `gorm:"column:entry_id;not null" json:"entry_id"`
	// FlowID is the ID of the associated workflow definition 关联的流程定义ID
	FlowID uint `gorm:"column:flow_id;not null" json:"flow_id"`
	// ProcessID is the ID of the process step at which the CC was triggered 触发抄送时所在的流程步骤ID
	ProcessID uint `gorm:"column:process_id;not null" json:"process_id"`
	// ProcID is the ID of the approval task that triggered the CC 触发抄送的审批任务ID
	ProcID uint `gorm:"column:proc_id;not null" json:"proc_id"`
	// EmpID is the ID of the CC recipient employee 抄送人ID
	EmpID int `gorm:"column:emp_id;not null;comment:'抄送人ID'" json:"emp_id"`
	// EmpName is the display name of the CC recipient employee 抄送人名称
	EmpName string `gorm:"column:emp_name;not null;default:'';comment:'抄送人名称'" json:"emp_name"`
	// Status indicates the read state: 0=unread (未读), 1=read (已读) 状态：0未读 1已读
	Status int `gorm:"column:status;not null;default:0;comment:'0未读 1已读'" json:"status"`
	// Entry is the associated workflow entry via foreign key entry_id 关联的流程实例（外键：entry_id）
	Entry Entry `gorm:"foreignKey:entry_id"` // 关联的Entry
}
