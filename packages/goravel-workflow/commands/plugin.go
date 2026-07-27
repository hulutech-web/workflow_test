package commands

import (
	"fmt"
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"goravel/packages/goravel-workflow/services/workflow/official_plugins"
)

// Plugin 工作流插件管理命令，用于通过命令行交互式地创建新的流程框架插件。
// Plugin is the workflow plugin management command, used to interactively create new workflow framework plugins via the command line.
type Plugin struct{}

// NewPlugin 创建并返回一个新的 Plugin 命令实例。
// NewPlugin creates and returns a new Plugin command instance.
func NewPlugin() *Plugin {
	return &Plugin{}
}

// Signature 控制台命令的名称和签名。返回 "make:plugin" 作为 Artisan 命令名。
// Signature The name and signature of the console command. Returns "make:plugin" as the Artisan command name.
func (receiver *Plugin) Signature() string {
	return "make:plugin"
}

// Description 控制台命令的描述信息，说明该命令用于创建流程框架插件。
// Description The console command description, explaining that this command is used to create a workflow framework plugin.
func (receiver *Plugin) Description() string {
	return "您正在创建一个流程框架插件"
}

// Extend 控制台命令的扩展配置。当前返回空扩展，预留后续自定义选项。
// Extend The console command extend. Currently returns an empty extend, reserved for future custom options.
func (receiver *Plugin) Extend() command.Extend {
	return command.Extend{}
}

// Handle 执行控制台命令的核心逻辑。
// 流程：询问插件信息 → 确认创建 → 初始化数据库表 → 写入插件记录。
// Handle Execute the console command core logic.
// Flow: prompt for plugin info → confirm creation → initialize database tables → insert plugin record.
func (receiver *Plugin) Handle(ctx console.Context) error {
	// 交互式询问：依次获取插件名称、版本、功能描述和作者信息
	// Interactive prompts: collect plugin name, version, description, and author
	name, _ := ctx.Ask("插件名称?")
	version, _ := ctx.Ask("插件版本?")
	description, _ := ctx.Ask("功能描述?")
	author, _ := ctx.Ask("插件作者?")

	// 设置确认选项：默认选择"是"
	// Set confirmation options: default to "yes"
	question := "确认创建吗?"
	options := []console.Choice{
		{Key: "yes", Value: "是", Selected: true},
		{Key: "no", Value: "否"},
	}

	// 弹出确认选择框
	// Display confirmation choice dialog
	c, err := ctx.Choice(question, options, console.ChoiceOption{
		Default: "yes",
	})
	if err != nil {
		return err
	}

	// 用户确认创建
	// User confirmed creation
	if c == "是" {
		ctx.Info("创建中...")

		// 启动插件表的自动迁移，确保 plugin 数据表存在
		// Boot plugin table auto-migration to ensure the plugin table exists
		orm := official_plugins.BootMS()
		if err != nil {
			fmt.Println("AutoMigrate error:", err)
			// 处理错误: 自动迁移失败时仍继续尝试创建记录
			// Handle error: continue trying to create record even if auto-migration fails
		} else {
			fmt.Println("AutoMigrate successful")
		}

		// 在数据库中创建插件记录，Status 默认为 1（启用）
		// Create plugin record in database, Status defaults to 1 (enabled)
		row := orm.Create(&official_plugins.Plugin{
			Name:        name,
			Version:     version,
			Status:      1,
			Description: description,
			Author:      author,
		})

		// 检查插入结果：RowsAffected 为 0 或有错误则表示创建失败
		// Check insert result: RowsAffected of 0 or an error indicates creation failure
		if row.RowsAffected == 0 || row.Error != nil {
			fmt.Println("Create error:", err)
		} else {
			fmt.Println("Create successful")
		}
	} else {
		// 用户取消创建，输出提示信息并退出
		// User cancelled creation, output cancel message and exit
		ctx.Info("取消创建")
		return nil
	}

	// 注意：由于上面的每个分支都已 return，此行只有在逻辑异常时才会执行
	// Note: since each branch above has returned, this line only executes on logic error
	ctx.Info("创建失败")
	return nil
}
