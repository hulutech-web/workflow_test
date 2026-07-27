package official_plugins

import (
	"goravel/packages/goravel-workflow/models"

	httpfacades "github.com/hulutech-web/http_result"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// DistributeController 分发插件控制器，负责插件的安装、卸载、配置管理及分发规则处理 / DistributePlugin controller: manages plugin install, uninstall, config, and distribution rules
type DistributeController struct {
	//Dependent services / 依赖服务
}

// NewDeptController 创建分发控制器实例 / Creates a new DistributeController instance
func NewDeptController() *DistributeController {
	return &DistributeController{
		//Inject services / 注入服务
	}
}

// DistributeRequest 分发请求参数结构体，包含流程ID、步骤ID及分发规则 / Request payload for plugin distribution: flow ID, process ID, and distribution rules
type DistributeRequest struct {
	FlowID    uint `json:"flow_id" form:"flow_id"`
	ProcessID uint `json:"process_id" form:"process_id"`
	Rules     Rule `json:"rules" form:"rules"`
}

// InstallPlugin 为流程安装插件，建立流程与插件的多对多关联关系 / Install a plugin for a flow: creates the many-to-many association between flow and plugin
func (r *DistributeController) InstallPlugin(ctx http.Context) http.Response {
	type SelRequest struct {
		FlowID   int `json:"flow_id" form:"flow_id"`
		PluginID int `json:"plugin_id" form:"plugin_id"`
	}
	var selRequest SelRequest
	// 绑定请求参数 / Bind request parameters
	ctx.Request().Bind(&selRequest)

	// 查询流程是否存在 / Check if the flow exists
	var flow models.Flow
	query := facades.Orm().Query()
	query.Model(&flow).Where("id=?", selRequest.FlowID).Find(&flow)

	// 查询插件是否存在 / Check if the plugin exists
	var plugin Plugin
	facades.Orm().Query().Model(&Plugin{}).Where("id=?", selRequest.PluginID).Find(&plugin)

	// 流程或插件不存在则返回错误 / Return error if flow or plugin does not exist
	if flow.ID == 0 || plugin.ID == 0 {
		return httpfacades.NewResult(ctx).Error(500, "流程或插件不存在", "")
	}

	// 创建流程与插件的关联记录 / Create flow-plugin association record
	query.Model(&FlowPlugin{}).Create(&FlowPlugin{
		FlowID:   uint(selRequest.FlowID),
		PluginID: uint(selRequest.PluginID),
	})
	// 通过 GORM 关联追加插件 / Append plugin via GORM association
	query.Model(&flow).Association("Plugins").Append(&plugin)

	return httpfacades.NewResult(ctx).Success("安装成功", "")
}

// List 获取所有插件列表，及其关联的插件配置 / List all plugins with their associated plugin configurations
func (r *DistributeController) List(ctx http.Context) http.Response {
	var plugins []Plugin
	// 预加载 PluginConfigs 关联数据 / Eager load PluginConfigs association
	err := facades.Orm().Query().Model(&plugins).With("PluginConfigs").Find(&plugins)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "获取失败", err)
	}
	return httpfacades.NewResult(ctx).Success("", plugins)
}

// StorePluginConfig 添加或更新插件配置规则，使用 UpdateOrCreate 实现幂等保存 / Add or update plugin config rules; uses UpdateOrCreate for idempotent save
func (r *DistributeController) StorePluginConfig(ctx http.Context) http.Response {
	type PluginConfigRequest struct {
		FieldID   int  `json:"field_id" form:"field_id"`
		FlowID    uint `json:"flow_id" form:"flow_id"`
		PluginID  uint `json:"plugin_id" form:"plugin_id"`
		ProcessID uint `json:"process_id" form:"process_id"`
		Rules     Rule `json:"rules" form:"rules"`
	}
	var pluginConfigRequest PluginConfigRequest
	// 绑定请求参数 / Bind request parameters
	ctx.Request().Bind(&pluginConfigRequest)

	// 创建或者更新：根据唯一键匹配（PluginID + FlowID + ProcessID + FieldID），存在则更新 Rules，不存在则创建 / Upsert: match by composite key (PluginID + FlowID + ProcessID + FieldID); update Rules if exists, create if not
	facades.Orm().Query().UpdateOrCreate(&PluginConfig{}, PluginConfig{
		PluginID:  pluginConfigRequest.PluginID,
		FlowID:    pluginConfigRequest.FlowID,
		ProcessID: pluginConfigRequest.ProcessID,
		FieldID:   pluginConfigRequest.FieldID,
	}, PluginConfig{Rules: pluginConfigRequest.Rules})

	return httpfacades.NewResult(ctx).Success("保存成功", "")
}

// GetPluginConfig 根据流程、步骤、字段及插件ID获取单个插件配置 / Get a single plugin config by flow, process, field, and plugin ID
func (r *DistributeController) GetPluginConfig(ctx http.Context) http.Response {
	type PluginConfigRequest struct {
		FieldID   int  `json:"field_id" form:"field_id"`
		FlowID    uint `json:"flow_id" form:"flow_id"`
		PluginID  uint `json:"plugin_id" form:"plugin_id"`
		ProcessID uint `json:"process_id" form:"process_id"`
		Rules     Rule `json:"rules" form:"rules"`
	}
	var pluginConfigRequest PluginConfigRequest
	// 绑定请求参数 / Bind request parameters
	ctx.Request().Bind(&pluginConfigRequest)

	var pluginConfig PluginConfig
	// 多条件精确查询插件配置 / Query plugin config with multiple exact-match conditions
	facades.Orm().Query().Model(&PluginConfig{}).
		Where("field_id=?", pluginConfigRequest.FieldID).
		Where("flow_id=?", pluginConfigRequest.FlowID).
		Where("plugin_id=?", pluginConfigRequest.PluginID).
		Where("process_id=?", pluginConfigRequest.ProcessID).Find(&pluginConfig)

	return httpfacades.NewResult(ctx).Success("", pluginConfig)
}

// GetAllPluginConfig 获取指定流程和插件的所有配置项，预加载流程、步骤及表单模板关联数据 / Get all config entries for a given flow and plugin, with Flow, Process, and TemplateForm preloaded
func (r *DistributeController) GetAllPluginConfig(ctx http.Context) http.Response {
	type PluginConfigRequest struct {
		FieldID   int  `json:"field_id" form:"field_id"`
		FlowID    uint `json:"flow_id" form:"flow_id"`
		PluginID  uint `json:"plugin_id" form:"plugin_id"`
		ProcessID uint `json:"process_id" form:"process_id"`
		Rules     Rule `json:"rules" form:"rules"`
	}
	var pluginConfigRequest PluginConfigRequest
	// 绑定请求参数 / Bind request parameters
	ctx.Request().Bind(&pluginConfigRequest)

	var pluginConfigs []PluginConfig
	// 预加载 Flow、Process、TemplateForm 关联，按 flow_id 和 plugin_id 筛选 / Eager load associations and filter by flow_id and plugin_id
	facades.Orm().Query().Model(&PluginConfig{}).With("Flow").With("Process").With("TemplateForm").
		Where("flow_id=?", pluginConfigRequest.FlowID).
		Where("plugin_id=?", pluginConfigRequest.PluginID).
		Find(&pluginConfigs)

	return httpfacades.NewResult(ctx).Success("", pluginConfigs)
}

// Product 开发者提交插件信息，通过设计生成插件的选项 / Developer submits plugin info; generates plugin options from the design specification
func (r *DistributeController) Product(ctx http.Context) http.Response {
	var distributeRequest DistributeRequest
	// 绑定分发请求参数 / Bind distribution request parameters
	ctx.Request().Bind(&distributeRequest)

	var pluginConfig PluginConfig
	// 创建新的插件配置记录（注：此处未使用 distributeRequest 的实际数据） / Create a new plugin config record (note: distributeRequest data is not used in this creation)
	err := facades.Orm().Query().Model(&PluginConfig{}).Create(&pluginConfig)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "制作成功", err)
	}
	return httpfacades.NewResult(ctx).Success("制作成功", pluginConfig)
}

// UninstallPlugin 卸载插件：删除流程与插件的关联关系及中间表记录 / Uninstall a plugin: removes the flow-plugin association and junction table records
func (r *DistributeController) UninstallPlugin(ctx http.Context) http.Response {
	type SelRequest struct {
		FlowID   int `json:"flow_id" form:"flow_id"`
		PluginID int `json:"plugin_id" form:"plugin_id"`
	}
	var selRequest SelRequest
	// 绑定请求参数 / Bind request parameters
	ctx.Request().Bind(&selRequest)

	// 查询流程是否存在 / Check if the flow exists
	var flow models.Flow
	query := facades.Orm().Query()
	query.Model(&flow).Where("id=?", selRequest.FlowID).Find(&flow)

	// 查询插件是否存在 / Check if the plugin exists
	var plugin Plugin
	facades.Orm().Query().Model(&Plugin{}).Where("id=?", selRequest.PluginID).Find(&plugin)

	// 流程或插件不存在则返回错误 / Return error if flow or plugin does not exist
	if flow.ID == 0 || plugin.ID == 0 {
		return httpfacades.NewResult(ctx).Error(500, "流程或插件不存在", "")
	}

	// 删除中间表记录（注：第二个 Where 条件使用了 selRequest.FlowID，疑似应为 selRequest.PluginID，这是一个潜在的 bug） / Delete junction table record (note: second Where uses FlowID instead of likely PluginID — potential bug)
	query.Model(&FlowPlugin{}).Where("flow_id=?", selRequest.FlowID).Where("plugin_id=?", selRequest.FlowID).
		Delete(&FlowPlugin{})

	// 通过 GORM 关联删除插件 / Remove plugin via GORM association
	query.Model(&flow).Association("Plugins").Delete(&plugin)

	return httpfacades.NewResult(ctx).Success("卸载成功", "")
}
