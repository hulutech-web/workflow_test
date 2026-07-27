package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

// PublishWorkflow 发布工作流资源的命令结构体，实现 console.Command 接口。
type PublishWorkflow struct{}

// NewPublishWorkflow 创建并返回一个新的 PublishWorkflow 命令实例。
func NewPublishWorkflow() *PublishWorkflow {
	return &PublishWorkflow{}
}

// Signature 返回该命令的签名（命令名称），用于在 Artisan 命令行中调用。
func (receiver *PublishWorkflow) Signature() string {
	return "workflow:publish"
}

// Description The console command description.
// 控制台命令的描述信息，返回命令的简要说明。
func (receiver *PublishWorkflow) Description() string {
	return "发布workflow资源"
}

// Extend The console command extend.
// 控制台命令的扩展配置，可定义命令的参数和选项。
func (receiver *PublishWorkflow) Extend() command.Extend {
	return command.Extend{}
}

// Handle Execute the console command.
// 执行控制台命令的核心逻辑，当前为占位实现，待后续补充资源发布功能。
func (receiver *PublishWorkflow) Handle(ctx console.Context) error {
	return nil
}
