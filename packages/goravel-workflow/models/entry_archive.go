package models

import (
	"github.com/goravel/framework/database/orm"
)

// EntryArchive 审批完结后完整快照，防止员工离职后关联查询失败。
// 所有动态数据序列化为 JSON 存储，永久可查。
type EntryArchive struct {
	orm.Model
	EntryID          uint   `gorm:"column:entry_id" json:"entry_id"`
	FlowID           uint   `gorm:"column:flow_id" json:"flow_id"`
	Status           int    `gorm:"column:status" json:"status"` // 9=通过 -1=驳回 -2=撤回
	Title            string `gorm:"column:title" json:"title"`
	FlowSnapshot     string `gorm:"column:flow_snapshot;type:text" json:"flow_snapshot"`
	EntrySnapshot    string `gorm:"column:entry_snapshot;type:text" json:"entry_snapshot"`
	FormDataSnapshot string `gorm:"column:form_data_snapshot;type:text" json:"form_data_snapshot"`
	ProcsSnapshot    string `gorm:"column:procs_snapshot;type:text" json:"procs_snapshot"`
	CommentsSnapshot string `gorm:"column:comments_snapshot;type:text" json:"comments_snapshot"`
	CCSnapshot       string `gorm:"column:cc_snapshot;type:text" json:"cc_snapshot"`
}
