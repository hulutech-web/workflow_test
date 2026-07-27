// Package official_plugins 提供工作流官方插件系统的数据模型定义，包含插件注册、配置和规则管理。
package official_plugins

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/goravel/framework/database/orm"
	"goravel/packages/goravel-workflow/models"
)

// Plugin 插件模型，存储插件的基本信息和关联的配置数据。
// 插件通过 FlowPlugin 中间表与流程(Flow)建立多对多关系，每个流程的每个步骤可以配置不同的插件规则。
type Plugin struct {
	orm.Model
	Name          string         `gorm:"column:name;unique;comment:'插件名称'" json:"name" form:"name"`    // 插件唯一名称
	Version       string         `gorm:"column:version;comment:'版本号'" json:"version" form:"version"`     // 插件版本号
	Status        int            `gorm:"column:status;comment:'状态'" json:"status" form:"status"`         // 插件状态（0:禁用, 1:启用）
	Description   string         `gorm:"column:description;comment:'描述'" json:"description" form:"description"` // 插件描述信息
	Author        string         `gorm:"column:author;comment:'作者'" json:"author" form:"author"`          // 插件作者
	IsDesigned    int            `gorm:"column:is_designed;comment:'是否完成设计'" json:"is_designed" form:"is_designed"` // 是否已完成设计（0:未完成, 1:已完成）
	PluginConfigs []PluginConfig `gorm:"foreignKey:PluginID;references:ID"`                             // 关联的插件配置列表
	Flows         []*models.Flow `gorm:"many2many:flow_plugins"`                                        // 关联的流程列表（多对多）
}

// FlowPlugin 流程与插件的中间关联表，用于连接 Flow 和 Plugin 的多对多关系。
// flow_plugin中间表
type FlowPlugin struct {
	orm.Model
	PluginID uint `gorm:"column:plugin_id;comment:'插件ID'" json:"plugin_id" form:"plugin_id"`
	FlowID   uint `gorm:"column:flow_id;comment:'流程ID'" json:"flow_id" form:"flow_id"`
}

// PluginConfig 插件配置模型，用于为特定流程、特定步骤定义插件的触发规则。
// 一个插件可以有多个配置，每个配置关联一个流程的一个步骤，并绑定表单中的某个字段。
// 为某一个流程中的某一个步骤添加规则
type PluginConfig struct {
	orm.Model
	PluginID     uint `gorm:"column:plugin_id;comment:'插件ID'" json:"plugin_id" form:"plugin_id"`                             // 关联的插件ID
	FlowID       uint `gorm:"column:flow_id" json:"flow_id" form:"flow_id"`                                                    // 关联的流程ID
	ProcessID    uint `gorm:"column:process_id" json:"process_id" form:"process_id"`                                            // 关联的流程步骤ID
	FieldID      int  `gorm:"column:field_id;comment:'对应template_form中的字段field对应的id'" json:"field_id" form:"field_id"`        // 关联的表单字段ID（对应template_form中的字段）
	Rules        Rule `gorm:"column:rules;type:text" json:"config" form:"rules"`                                               // 规则配置（JSON格式存储在text字段中）
	Flow         *models.Flow
	Process      *models.Process
	TemplateForm *models.TemplateForm `gorm:"foreignKey:FieldID;references:ID"`
}

// RuleItem 规则项，表示单个规则的具体配置。包含规则标识、显示标签和对应的值。
type RuleItem struct {
	RuleID    int    `json:"rule_id" form:"rule_id"`       // 规则ID（部门ID）
	RuleLabel string `json:"rule_label" form:"rule_label"` // 规则标签（部门名称）
	RuleValue int    `json:"rule_value" form:"rule_value"` // 规则值（部门值）
}

// Rule 规则类型，是 RuleItem 的切片，表示一组规则配置。
// 实现了 database/sql 的 Scanner 和 Valuer 接口，支持将 Rule 序列化/反序列化为 JSON 存入数据库。
type Rule []RuleItem

// Scan 实现 sql.Scanner 接口，将数据库中的 JSON 字节数据反序列化为 Rule 类型。
// 用于从数据库读取规则配置时自动将 JSON 数据转换为 Go 结构体。
func (t *Rule) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

// Value 实现 driver.Valuer 接口，将 Rule 类型序列化为 JSON 格式存入数据库。
// 当 Rule 为 nil 时，json.Marshal 会返回 "null"，直接存入数据库。
func (t Rule) Value() (driver.Value, error) {
	//如果t为nil,返回nil
	return json.Marshal(t)
}
