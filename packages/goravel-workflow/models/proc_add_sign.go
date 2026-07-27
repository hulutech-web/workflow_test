package models

import (
	"github.com/goravel/framework/database/orm"
)

// ProcAddSign represents an add-signer record in a workflow approval process
// 加签记录模型，用于记录审批流程中的加签操作（前加签/后加签）
type ProcAddSign struct {
	// orm.Model embeds the base ORM model fields (ID, CreatedAt, UpdatedAt, DeletedAt)
	// 嵌入的基础ORM模型字段（ID、创建时间、更新时间、删除时间）
	orm.Model
	// EntryID is the ID of the workflow entry this add-sign belongs to
	// 所属流程实例ID
	EntryID uint `gorm:"column:entry_id;not null" json:"entry_id"`
	// ProcID is the ID of the approval task (proc) where the add-sign was initiated
	// 发起加签的审批任务ID
	ProcID uint `gorm:"column:proc_id;not null" json:"proc_id"`
	// SignType indicates the type of add-sign: "before" for pre-sign (前加签) or "after" for post-sign (后加签)
	// 加签类型：before 表示前加签（在当前审批人之前加签），after 表示后加签（在当前审批人之后加签）
	SignType string `gorm:"column:sign_type;not null;default:'before';comment:'前加签/后加签'" json:"sign_type"`
	// SignEmpID is the employee ID of the added signer
	// 被加签人的员工ID
	SignEmpID int `gorm:"column:sign_emp_id;not null" json:"sign_emp_id"`
	// SignEmpName is the name of the added signer
	// 被加签人的员工姓名
	SignEmpName string `gorm:"column:sign_emp_name;not null;default:''" json:"sign_emp_name"`
	// Status indicates the processing state: 0=pending (待处理), 1=completed (已完成)
	// 加签状态：0 表示待处理，1 表示已完成
	Status int `gorm:"column:status;not null;default:0;comment:'0待处理 1已完成'" json:"status"`
}
