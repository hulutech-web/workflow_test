package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

// Flowlink represents a connection/link between workflow processes, defining routing rules and approver configuration
// Flowlink 表示工作流步骤之间的流转连线，定义路由规则和审批人配置
type Flowlink struct {
	orm.Model
	// Type is the link type: "Condition" for conditional routing, or role-based ("Emp", "Dept", "Sys")
	// 类型："Condition"表示条件流转，角色类型表示当前步骤操作人（"Emp"指定人员、"Dept"指定部门、"Sys"系统自动）
	Type string `gorm:"column:type;not null;comment:'Condition:表示步骤流转\\nRole:当前步骤操作人'"`
	// Auditor specifies the approver settings: system auto, specific employee, specific department, or specific role
	// Not used when Type=Condition. Special values: -1000=发起人, -1001=部门主管, -1002=部门经理, -1003=表单字段, -1004=动态表达式
	// 审批人设置：系统自动、指定人员、指定部门、指定角色。type=Condition时不启用。特殊值：-1000=发起人自己、-1001=部门主管、-1002=部门经理、-1003=从表单字段读取、-1004=动态表达式
	Auditor string `gorm:"column:auditor;not null;default:'0';comment:'审批人 系统自动 指定人员 指定部门 指定角色\\ntype=Condition时不启用'"`
	// Expression is the condition evaluation expression in JSON format
	// "1" means unconditional (always true), passing directly to the next step
	// 条件判断表达式，JSON格式。"1"表示无条件为true，通过则直接进入下一步骤
	Expression string `gorm:"column:expression;not null;default:'';comment:'条件判断表达式\\n为1表示true，通过的话直接进入下一步骤'"`
	// ConditionExpr is the raw condition expression string
	// 条件表达式字符串
	ConditionExpr string `gorm:"column:condition_expr;not null;default:'';comment:'条件表达式'"`
	// Sort determines the evaluation order of condition branches (ascending)
	// 条件分支的判断顺序，按升序排列
	Sort int `gorm:"column:sort;not null;comment:'条件判断顺序'"`
	// ConcurrencyType is the signing mode: 0=sequential (依次), 1=consensus/countersign (会签), 2=any-sign (或签)
	// 并签模式：0=依次审批，1=会签（全部通过才流转），2=或签（任一通过即流转，其余跳过）
	ConcurrencyType int `gorm:"column:concurrency_type;not null;default:0;comment:'并签模式: 0=依次, 1=会签, 2=或签'"`
	// ApproverRule is the approver assignment rule
	// -1003: reads approver from a form field (ApproverRule stores the field name)
	// -1004: dynamic expression (ApproverRule stores the expression mapping key)
	// 审批人分配规则：-1003=从表单字段读取审批人（ApproverRule存储字段名），-1004=动态表达式（ApproverRule存储映射键）
	ApproverRule string `gorm:"column:approver_rule;not null;default:'';comment:'审批人分配规则: -1003=表单字段, -1004=动态表达式'"`

	// FlowID is the foreign key to the parent Flow 流程ID，关联所属流程
	FlowID uint `gorm:"column:flow_id"`
	// ProcessID is the foreign key to the current/source Process 当前步骤ID，关联来源步骤
	ProcessID uint `gorm:"column:process_id"`
	// NextProcessID is the foreign key to the target Process (-1 means last step / end of workflow)
	// 下一步骤ID，关联目标步骤。-1表示最后一步（流程结束）
	NextProcessID int  `gorm:"column:next_process_id;default:2"`
	// Process is the source process relationship (HasOne) 来源步骤关联（HasOne）
	Process Process `gorm:"foreignKey:ProcessID;references:id"`
	// NextProcess is the target/destination process relationship (HasOne) 目标步骤关联（HasOne）
	NextProcess Process `gorm:"foreignKey:NextProcessID;references:id"`
	// Flow is the parent flow relationship (BelongsTo) 所属流程关联（BelongsTo）
	Flow Flow `gorm:"foreignKey:FlowID;references:id"`
}

// LoadProcess eagerly loads the source Process relationship for this Flowlink
// 预加载当前步骤（Process）关联数据
func (fl *Flowlink) LoadProcess(db *gorm.DB) error {
	return db.Preload("Process").Find(fl).Error
}

// LoadProcesses eagerly loads the Processes relationship for this Flowlink
// 预加载步骤集合（Processes）关联数据
func (fl *Flowlink) LoadProcesses(db *gorm.DB) error {
	return db.Preload("Processes").Find(fl).Error
}

// LoadNextProcess eagerly loads the target NextProcess relationship for this Flowlink
// 预加载下一步骤（NextProcess）关联数据
func (fl *Flowlink) LoadNextProcess(db *gorm.DB) error {
	return db.Preload("NextProcess").Find(fl).Error
}
