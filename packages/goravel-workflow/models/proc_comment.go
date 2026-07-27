package models

import (
	"github.com/goravel/framework/database/orm"
)

// ProcComment 审批评论模型，记录审批流程中的评论与回复
type ProcComment struct {
	orm.Model
	// EntryID 所属流程实例ID
	EntryID uint `gorm:"column:entry_id;not null" json:"entry_id"`
	// ProcID 所属审批任务ID
	ProcID uint `gorm:"column:proc_id;not null" json:"proc_id"`
	// EmpID 评论发言人ID
	EmpID int `gorm:"column:emp_id;not null;comment:'发言人ID'" json:"emp_id"`
	// EmpName 评论发言人名称
	EmpName string `gorm:"column:emp_name;not null;default:'';comment:'发言人名称'" json:"emp_name"`
	// Content 评论内容
	Content string `gorm:"column:content;type:text;not null;comment:'评论内容'" json:"content"`
	// Status 评论状态: 1正常 2删除
	Status int `gorm:"column:status;not null;default:1;comment:'1正常 2删除'" json:"status"`
	// ParentID 父评论ID，0表示顶级评论
	ParentID uint `gorm:"column:parent_id;not null;default:0;comment:'父评论ID，0表示顶级评论'" json:"parent_id"`
	// ReplyToEmpID 被回复的员工ID
	ReplyToEmpID int `gorm:"column:reply_to_emp_id;not null;default:0;comment:'被回复的员工ID'" json:"reply_to_emp_id"`
	// ReplyToEmpName 被回复的员工名称
	ReplyToEmpName string `gorm:"column:reply_to_emp_name;not null;default:'';comment:'被回复的员工名称'" json:"reply_to_emp_name"`
	// Children 子评论列表（回复链），通过ParentID外键关联
	Children []ProcComment `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}
