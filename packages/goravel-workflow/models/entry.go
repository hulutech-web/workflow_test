package models

import (
	"github.com/goravel/framework/database/orm"
)

// Entry status constants 流程实例状态常量
const (
	EntryStatusPending   = 0  // In progress 进行中
	EntryStatusCompleted = 9  // Completed 已完成
	EntryStatusRejected  = -1 // Rejected 已驳回
	EntryStatusRevoked   = -2 // Revoked 已撤回
)

// Entry represents a workflow instance / approval entry 流程实例模型，代表一次具体的审批申请
type Entry struct {
	orm.Model
	// Title is the display title of this entry 流程实例标题
	Title string `gorm:"column:title;not null;default:''" json:"title" form:"title"`
	// FlowID is the ID of the workflow definition this entry belongs to 所属流程定义ID
	FlowID uint `gorm:"column:flow_id;not null;default:0" json:"flow_id" form:"flow_id"`
	// EmpID is the ID of the initiator / applicant 发起人（员工）ID
	EmpID uint `gorm:"column:emp_id;not null;default:0" json:"emp_id" form:"emp_id"`
	// ProcessID is the ID of the current step this entry is at 当前所在步骤ID
	ProcessID uint `gorm:"column:process_id;not null;default:0" json:"process_id" form:"process_id"`
	// Circle is the round number, incremented on resend 轮次编号，重新发起时递增
	Circle int `gorm:"column:circle;not null;default:1" json:"circle" form:"circle"`
	// Status is the current status of this entry (0=pending, 9=completed, -1=rejected, -2=revoked) 当前状态（0=进行中, 9=已完成, -1=已驳回, -2=已撤回）
	Status int `gorm:"column:status;not_null" json:"status" form:"status"`
	// Pid is the parent entry ID, used for child workflow linkage 父流程实例ID，用于子流程关联
	Pid int `gorm:"column:pid;not null;default:0" json:"pid" form:"pid"`
	// EnterProcessID is the entry process ID when this entry was created 创建时的入口步骤ID
	EnterProcessID int `gorm:"column:enter_process_id;not null;default:0" json:"enter_process_id" form:"enter_process_id"`
	// EnterProcID is the entry proc task ID when this entry was created 创建时的入口审批任务ID
	EnterProcID int `gorm:"column:enter_proc_id;not null;default:0" json:"enter_proc_id" form:"enter_proc_id"`
	// Child is the current process ID of the child workflow 子流程当前所在步骤ID
	Child int `gorm:"column:child;not null;default:0" json:"child" form:"child"`
	// Flow is the associated Flow (workflow definition) 关联的流程定义
	Flow Flow `gorm:"foreignKey:flow_id"`
	// Emp is the associated Employee (initiator) 关联的发起人（员工）
	Emp Emp `gorm:"foreignKey:emp_id"`
	// Procs is the list of approval tasks under this entry 该流程实例下的所有审批任务
	Procs []*Proc
	// Process is the associated Process (current step) 关联的当前步骤
	Process Process `gorm:"foreignKey:process_id"`
	// EntryDatas is the list of form data records for this entry 该流程实例的表单数据记录
	EntryDatas []EntryData
	// ParentEntry is the associated parent Entry (for child workflows) 关联的父流程实例（用于子流程）
	ParentEntry *Entry `gorm:"foreignKey:pid"`
	// Children is the list of child Entries (cascading delete) 子流程实例列表，级联删除
	Children []Entry `gorm:"foreignKey:pid"`
	// EnterProcess is the associated entry process 关联的入口步骤
	EnterProcess Process `gorm:"foreignKey:enter_process_id"`
}
