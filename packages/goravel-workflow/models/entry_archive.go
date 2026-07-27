package models

import (
	"github.com/goravel/framework/database/orm"
)

// EntryArchive 审批完结后完整快照，防止员工离职后关联查询失败。
// 所有动态数据序列化为 JSON 存储，永久可查。
// EntryArchive is a complete snapshot taken after workflow completion, preventing query failures when employees leave.
// EntryArchive 是审批完结后的完整快照模型，在流程实例结束后保存所有关联数据的完整副本，
// 确保即使员工离职、部门变更或原始数据被修改后，历史审批记录仍然可以完整查询。
// 所有动态数据（流程定义、表单数据、审批记录、评论、抄送等）均序列化为 JSON 文本存储，永久可查。
type EntryArchive struct {
	// orm.Model 嵌入基础模型字段（ID, CreatedAt, UpdatedAt, DeletedAt）
	orm.Model
	// EntryID is the ID of the original workflow entry 原始流程实例ID
	EntryID uint `gorm:"column:entry_id" json:"entry_id"`
	// FlowID is the ID of the workflow definition 所属流程定义ID
	FlowID uint `gorm:"column:flow_id" json:"flow_id"`
	// Status is the final status of the entry 流程最终状态
	// 9=通过(已完成) -1=驳回 -2=撤回
	Status int `gorm:"column:status" json:"status"` // 9=通过 -1=驳回 -2=撤回
	// Title is the title of the entry 流程实例标题
	Title string `gorm:"column:title" json:"title"`
	// FlowSnapshot stores the JSON-serialized flow definition at completion time 流程定义快照（JSON），记录完结时的流程结构
	FlowSnapshot string `gorm:"column:flow_snapshot;type:text" json:"flow_snapshot"`
	// EntrySnapshot stores the JSON-serialized entry data at completion time 流程实例快照（JSON），记录完结时的实例数据
	EntrySnapshot string `gorm:"column:entry_snapshot;type:text" json:"entry_snapshot"`
	// FormDataSnapshot stores the JSON-serialized form field values at completion time 表单数据快照（JSON），记录完结时所有表单字段的值
	FormDataSnapshot string `gorm:"column:form_data_snapshot;type:text" json:"form_data_snapshot"`
	// ProcsSnapshot stores the JSON-serialized approval task records at completion time 审批任务快照（JSON），记录完结时所有审批节点的处理记录
	ProcsSnapshot string `gorm:"column:procs_snapshot;type:text" json:"procs_snapshot"`
	// CommentsSnapshot stores the JSON-serialized approval comments at completion time 审批评论快照（JSON），记录完结时所有评论内容
	CommentsSnapshot string `gorm:"column:comments_snapshot;type:text" json:"comments_snapshot"`
	// CCSnapshot stores the JSON-serialized CC (carbon copy) records at completion time 抄送记录快照（JSON），记录完结时所有抄送信息
	CCSnapshot string `gorm:"column:cc_snapshot;type:text" json:"cc_snapshot"`
}
