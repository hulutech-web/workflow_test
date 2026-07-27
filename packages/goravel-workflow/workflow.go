// Package workflow 工作流引擎核心包，提供服务提供者注册、路由、命令等功能
package workflow

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/http"
	commands "goravel/packages/goravel-workflow/commands"
)

// Workflow 工作流结构体，封装 HTTP 上下文，作为工作流引擎的入口点
type Workflow struct {
	// Context HTTP 上下文，用于处理请求和响应
	Context http.Context
}

// NewWorkflow 创建新的 Workflow 实例
// 接收 HTTP 上下文作为参数，返回初始化的 Workflow 指针
func NewWorkflow(ctx http.Context) *Workflow {
	return &Workflow{
		Context: ctx,
	}
}

// RegisterCommands registers all workflow commands
// 注册所有工作流相关的 Artisan 命令
func (w *Workflow) RegisterCommands() []console.Command {
	return []console.Command{
		commands.NewTimeoutCheckCommand(), // 超时检查命令：定期扫描超时未处理的审批任务
	}
}
