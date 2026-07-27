package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
	"gorm.io/gorm"
)

// Proc status constants 审批任务状态常量
const (
	ProcStatusPending     = 0  // 待处理
	ProcStatusApproved    = 1  // 已通过
	ProcStatusRejected    = -1 // 已驳回
	ProcStatusRevoked     = -2 // 已撤回
	ProcStatusTransferred = 3  // 已转交
	ProcStatusSkipped     = 4  // 已跳过（或签中未被选中的审批人）
	ProcStatusConsensus   = 9  // 会签通过
)

// Proc represents an individual approval task within a workflow entry 审批任务模型，表示流程实例中的单个审批环节
type Proc struct {
	orm.Model
	EntryID               uint             `gorm:"column:entry_id;not null" json:"entry_id" form:"entry_id"`                                          // 所属流程实例ID
	FlowID                int              `gorm:"column:flow_id;not null;comment:'流程id'" json:"flow_id" form:"flow_id"`                               // 所属流程定义ID
	ProcessID             int              `gorm:"column:process_id;not null;comment:'当前步骤'" json:"process_id" form:"process_id"`                      // 当前步骤ID
	ProcessName           string           `gorm:"column:process_name;not null;default:'';comment:'当前步骤名称'" json:"process_name" form:"process_name"`   // 当前步骤名称
	EmpID                 int              `gorm:"column:emp_id;not null;comment:'审核人'" json:"emp_id" form:"emp_id"`                                   // 指派的审批人ID
	EmpName               string           `gorm:"column:emp_name;default:null;comment:'审核人名称'" json:"emp_name" form:"emp_name"`                       // 审批人姓名
	DeptName              string           `gorm:"column:dept_name;default:null;comment:'审核人部门名称'" json:"dept_name" form:"dept_name"`                  // 审批人所属部门名称
	AuditorID             int              `gorm:"column:auditor_id;not null;default:0;comment:'具体操作人'" json:"auditor_id" form:"auditor_id"`           // 实际执行操作的人员ID（可能与审批人不同，如转交、加签场景）
	AuditorName           string           `gorm:"column:auditor_name;not null;default:'';comment:'操作人名称'" json:"auditor_name" form:"auditor_name"`    // 实际操作人姓名
	AuditorDept           string           `gorm:"column:auditor_dept;not null;default:'';comment:'操作人部门'" json:"auditor_dept" form:"auditor_dept"`    // 实际操作人所属部门
	Status                int              `gorm:"column:status;not null;comment:'当前处理状态 0待处理 1已通过 9会签通过 -1驳回 -2撤回 3转交 4跳过'" json:"status" form:"status"` // 当前处理状态：0待处理 1已通过 9会签通过 -1驳回 -2撤回 3转交 4跳过
	Content               string           `gorm:"column:content;default:null;comment:'批复内容'" json:"content" form:"content"`                          // 审批意见/批复内容
	IsRead                int              `gorm:"column:is_read;not null;default:0;comment:'是否查看'" json:"is_read" form:"is_read"`                     // 是否已被查看 0未读 1已读
	IsReal                bool             `gorm:"column:is_real;not null;default:1;comment:'审核人和操作人是否同一人'" json:"is_real" form:"is_real"`           // 审核人与操作人是否为同一人
	Circle                int              `gorm:"column:circle;not null;default:1" json:"circle" form:"circle"`                                       // 轮次编号（重发时递增）
	Beizhu                string           `gorm:"column:beizhu;type:text;comment:'备注'" json:"beizhu" form:"beizhu"`                                   // 备注信息
	Concurrence           *carbon.DateTime `gorm:"column:concurrence;not null;default:0;comment:'并行查找解决字段， 部门 角色 指定 分组用'" json:"concurrence" form:"concurrence"` // 并行审批的创建时间，用于按部门/角色/指定/分组查找同一批次的审批任务
	UnpassTargetProcessID int              `gorm:"column:unpass_target_process_id;not null;default:0;comment:'驳回到指定节点的目标步骤ID'" json:"unpass_target_process_id" form:"unpass_target_process_id"` // 驳回到指定节点时记录的目标步骤ID
	Emp                   Emp              `gorm:"foreignKey:EmpID"`                                                  // 关联的审批人（Emp模型） 关联员工
	Entry                 Entry            `gorm:"foreignKey:EntryID"`                                                // 关联的流程实例（Entry模型） 关联流程实例
	Process               Process          `gorm:"foreignKey:ProcessID"`                                              // 关联的流程步骤（Process模型） 关联流程步骤
	Flow                  Flow             `gorm:"foreignKey:FlowID"`                                                 // 关联的流程定义（Flow模型） 关联流程定义
	SubProcs              []Proc           `gorm:"foreignkey:EntryID;constraint:OnUpdate:CASCADE,OnDelete:NO ACTION"` // HasMany Proc 同一流程实例下的所有审批任务
	AddedSigns            []ProcAddSign    `gorm:"foreignkey:ProcID;constraint:OnUpdate:CASCADE,OnDelete:NO ACTION"`  // HasMany ProcAddSign 关联的加签记录
}

// LoadEmp preloads the associated Emp for a Proc. 预加载关联的员工信息
func (p *Proc) LoadEmp(db *gorm.DB) error {
	return db.Preload("Emp").Find(p).Error
}

// LoadEntry preloads the associated Entry for a Proc. 预加载关联的流程实例信息
func (p *Proc) LoadEntry(db *gorm.DB) error {
	return db.Preload("Entry").Find(p).Error
}

// LoadProcess preloads the associated Process for a Proc. 预加载关联的流程步骤信息
func (p *Proc) LoadProcess(db *gorm.DB) error {
	return db.Preload("Process").Find(p).Error
}

// LoadFlow preloads the associated Flow for a Proc. 预加载关联的流程定义信息
func (p *Proc) LoadFlow(db *gorm.DB) error {
	return db.Preload("Flow").Find(p).Error
}

// LoadSubProcs preloads the associated SubProcs (child processes) for a Proc. 预加载关联的子审批任务列表
func (p *Proc) LoadSubProcs(db *gorm.DB) error {
	return db.Preload("SubProcs", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "entry_id") // Optionally select specific fields. 可选地仅查询特定字段
	}).Find(p).Error
}
