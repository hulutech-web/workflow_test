package workflow

import (
	"goravel/app/models"
	commands "goravel/packages/goravel-workflow/commands"
	"goravel/packages/goravel-workflow/routes"
	"goravel/packages/goravel-workflow/services/workflow"
	"reflect"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
)

// Binding 是服务容器中绑定的键名，用于在其他地方通过 app.Make("workflow") 获取工作流实例
const Binding = "workflow"

// App 持有全局的 foundation.Application 实例，供工作流包内部在需要时直接使用
var App foundation.Application

// ServiceProvider 是 Goravel 的服务提供者结构体，负责注册工作流引擎所需的所有资源：
// 配置、路由、数据库迁移、种子数据、命令行等
type ServiceProvider struct {
}

// Register 在应用启动时注册工作流相关的所有资源。
// 该方法由 Goravel 框架自动调用，不应手动调用。
func (receiver *ServiceProvider) Register(app foundation.Application) {
	// 将应用实例保存到包级全局变量，供内部各处使用
	App = app

	// 配置文件 — 注入工作流默认配置到应用配置容器中
	config := app.MakeConfig()
	config.Add("workflow", map[string]any{
		"Dept": "Department", //部门关联应用中的模型
		"Emp":  "User",       //员工关联应用中的模型
	})
	// 将 Workflow 实例绑定到服务容器，后续可通过 Binding 常量获取
	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return NewWorkflow(nil), nil
	})

	// 路由 — 注册所有工作流 API 路由（入口、审批、流程管理等）
	routes.Api(app)

	// 数据库迁移 — 发布 Go 迁移文件（带标签），供本地项目引入迁移表结构
	app.Publishes("goravel/packages/goravel-workflow", map[string]string{
		"migrations/2024_06_24_000000_create_workflow_base_tables.go": app.DatabasePath("migrations/2024_06_24_000000_create_workflow_base_tables.go"),
	}, "migrations")

	// 种子数据 — 发布测试/初始数据填充文件
	app.Publishes("goravel/packages/goravel-workflow", map[string]string{
		"seeders/workflow_seeder.go":          app.DatabasePath("seeders/workflow_seeder.go"),
		"seeders/workflow_dept_seeder.go":     app.DatabasePath("seeders/workflow_dept_seeder.go"),
		"seeders/workflow_emp_seeder.go":      app.DatabasePath("seeders/workflow_emp_seeder.go"),
		"seeders/workflow_flowtype_seeder.go": app.DatabasePath("seeders/workflow_flowtype_seeder.go"),
	}, "seeders")

	// 配置文件 — 发布工作流配置文件模板到应用的 config 目录
	app.Publishes("goravel/packages/goravel-workflow", map[string]string{
		"config/workflow.go": app.ConfigPath("workflow.go"),
	}, "config")

	// 注册 Artisan 命令行：插件管理命令、超时检查命令
	app.Commands([]console.Command{
		commands.NewPlugin(),
		commands.NewTimeoutCheckCommand(),
	})
}

// Boot 在 Register 之后执行，负责启动阶段的初始化操作：
// 注册资源发布命令并将子应用的钩子方法注册到工作流引擎中
func (receiver *ServiceProvider) Boot(app foundation.Application) {
	// 注册资源发布命令 — 发布工作流迁移、配置、种子文件到宿主应用
	app.Commands([]console.Command{
		commands.NewPublishWorkflow(),
	})

	// 创建基础工作流实例
	wf := workflow.NewBaseWorkflow()
	// 将子应用 User 模型的钩子方法注册到工作流引擎中，供审批流转时回调
	user := &models.User{Workflow: wf}
	// 注册通知钩子：当条目被驳回或流程完成时，通知流程发起人
	wf.RegisterHook("NotifySendOneHook", reflect.ValueOf(user.NotifySendOne))
	// 注册通知钩子：当前审批人通过后，通知下一级审批人
	wf.RegisterHook("NotifyNextAuditorHook", reflect.ValueOf(user.NotifyNextAuditor))
}
