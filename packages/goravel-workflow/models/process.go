package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

// Process represents a step definition within a workflow flow 流程步骤定义模型，表示流程中的一个审批节点
type Process struct {
	orm.Model
	FlowID           int       `gorm:"column:flow_id;not null;default:0;comment:'流程id'" json:"flow_id"`                                                               // FlowID is the ID of the parent flow this process belongs to 所属流程ID
	ProcessName      string    `gorm:"column:process_name;not null;default:'';comment:'步骤名称'" json:"process_name"`                                                    // ProcessName is the display name of this process step 步骤名称
	LimitTime        int       `gorm:"column:limit_time;not null;default:0;comment:'限定时间,单位秒'" json:"limit_time"`                                                     // LimitTime is the time limit for this step in seconds, 0 means no limit 审批限定时间（单位：秒），0表示不限时
	Type             string    `gorm:"column:type;not null;default:'operation';comment:'流程图显示操作框类型'" json:"type"`                                                      // Type determines the shape/style of the process box in the flowchart diagram 流程图显示的操作框类型
	Icon             string    `gorm:"column:icon;default:'';comment:'流程图显示图标'" json:"icon,omitempty"`                                                                 // Icon is the icon displayed for this process in the flowchart 流程图显示的图标
	ProcessTo        string    `gorm:"column:process_to;not null;default:''" json:"process_to"`                                                                        // ProcessTo specifies connection target info for this process 流程指向目标信息
	Style            string    `gorm:"column:style;type:text;" json:"style,omitempty"`                                                                                  // Style defines custom CSS style for the flowchart node 流程图节点自定义样式
	StyleColor       string    `gorm:"column:style_color;not null;default:'#78a300'" json:"style_color"`                                                                // StyleColor is the background color of the process node in the flowchart 流程图节点的背景颜色
	StyleHeight      int       `gorm:"column:style_height;not null;default:30" json:"style_height"`                                                                     // StyleHeight is the height of the process node in pixels 流程图节点高度（像素）
	StyleWidth       int       `gorm:"column:style_width;not null;default:30" json:"style_width"`                                                                       // StyleWidth is the width of the process node in pixels 流程图节点宽度（像素）
	PositionLeft     string    `gorm:"column:position_left;not null;default:'100px'" json:"position_left"`                                                              // PositionLeft is the horizontal position of the node in the flowchart canvas 流程图节点水平位置
	PositionTop      string    `gorm:"column:position_top;not null;default:'200px'" json:"position_top"`                                                                // PositionTop is the vertical position of the node in the flowchart canvas 流程图节点垂直位置
	Position         int       `gorm:"column:position;not null;default:1;comment:'步骤位置：0第一步(开始) 1正常步骤 2转入子流程 9结束 当为2时 child_flow_id child_after child_back_process 可设置'" json:"position"` // Position indicates the role of this step: 0=first step (start), 1=normal step, 2=enter child workflow, 9=end 步骤位置：0第一步（开始），1正常步骤，2转入子流程，9结束。当值为2时可设置child_flow_id、child_after、child_back_process
	ChildFlowID      int       `gorm:"column:child_flow_id;not null;default:0;comment:'子流程id'" json:"child_flow_id"`                                                   // ChildFlowID is the ID of the child flow when this is a child-workflow entry step (Position=2) 子流程ID，当Position=2时指定要转入的子流程
	ChildAfter       int       `gorm:"column:child_after;not null;default:2;comment:'子流程结束后 1.同时结束父流程 2.返回父流程'" json:"child_after"`                                    // ChildAfter defines behavior after child workflow completes: 1=end parent flow too, 2=return to parent flow 子流程结束后的行为：1同时结束父流程，2返回父流程继续
	ChildBackProcess int       `gorm:"column:child_back_process;not null;default:0;comment:'子流程结束后返回父流程进程'" json:"child_back_process"`                                   // ChildBackProcess is the parent process ID to return to when ChildAfter=2 子流程结束后返回父流程的目标步骤ID（ChildAfter=2时生效）
	Description      string    `gorm:"column:description;not null;default:'';comment:'步骤描述'" json:"description"`                                                        // Description provides a human-readable description of this process step 步骤描述信息
	CcEmpIDs         string    `gorm:"column:cc_emp_ids;type:text;comment:'抄送人ID列表,逗号分隔'" json:"cc_emp_ids"`                                                            // CcEmpIDs stores comma-separated employee IDs who should receive a carbon copy when this step is approved 抄送人员ID列表，逗号分隔，审批通过后自动抄送
	ProcessVars      []Process `gorm:"many2many:process_vars;" json:"process_vars"`                                                                                     // ProcessVars defines the many-to-many relationship with condition variables for this process 关联的条件变量列表，用于条件分支判断
	Flow             Flow      // Flow is the parent workflow that this process belongs to 所属的流程定义
}

// LoadFlow preloads the associated Flow for a Process.
// 预加载关联的流程定义（Flow）数据
func (p *Process) LoadFlow(db *gorm.DB) error {
	return db.Preload("Flow").Find(p).Error
}

// LoadProcessVars preloads the associated Process variable definitions.
// 预加载关联的流程变量（ProcessVars）数据，用于条件分支判断
func (p *Process) LoadProcessVars(db *gorm.DB) error {
	return db.Preload("Processvars").Find(p).Error
}
