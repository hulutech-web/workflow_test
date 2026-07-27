package official_plugins

import "github.com/goravel/framework/contracts/foundation"

// RouteApi 注册插件系统的 HTTP API 路由
// 该方法在服务提供者启动时被调用，将所有插件管理相关的路由注册到应用中，
// 为工作流节点绑定和执行插件提供完整的 RESTful 接口支持
func (c *DistributePlugin) RouteApi(app foundation.Application) {
	// 获取应用的路由实例，用于注册 HTTP 路由规则
	router := app.MakeRoute()
	// 创建插件控制器实例，负责处理所有插件相关的 HTTP 请求
	distributeCtrl := NewDeptController()

	// 步骤 1：命令行新建插件（通过 artisan 命令生成插件骨架代码后，调用此接口持久化插件产物）
	// 步骤 2：开发者通过可视化设计器，设计该插件的选项和规则参数（前端设计阶段，不涉及后端接口）
	router.Post("api/plugin/product", distributeCtrl.Product)
	// 步骤 3：为流程安装或卸载插件
	// 安装：将指定插件绑定到目标流程，使流程执行时触发该插件的回调逻辑
	router.Post("api/flow/install_plugin", distributeCtrl.InstallPlugin)
	// 卸载：解除插件与目标流程的绑定关系，流程执行时不再触发该插件
	router.Post("api/flow/uninstall_plugin", distributeCtrl.UninstallPlugin)
	// 步骤 4：获取系统中所有已注册的插件列表，供前端选择和管理
	router.Get("api/plugin/list", distributeCtrl.List)
	// 步骤 5：为某个流程的某个步骤中的某个字段配置插件规则
	// 保存插件配置：将特定字段的插件规则（如校验规则、数据处理规则）持久化到数据库
	router.Post("api/plugin/store_plugin_config", distributeCtrl.StorePluginConfig)
	// 获取单个插件配置：根据条件查询指定插件在特定流程/步骤/字段下的配置详情
	router.Post("api/plugin/get_plugin_config", distributeCtrl.GetPluginConfig)
	// 获取全部插件配置：查询某个流程或步骤下所有已安装插件的完整配置信息
	router.Post("api/plugin/getall_plugin_config", distributeCtrl.GetAllPluginConfig)
}
