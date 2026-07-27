// Package facades 提供工作流门面（Facade），用于以静态方式快速获取工作流服务实例。
// 门面模式封装了对 Goravel 服务容器的调用，简化了业务代码中获取 Workflow 服务的方式。
package facades

import (
	"goravel/packages/goravel-workflow"
	"goravel/packages/goravel-workflow/contracts"
	"log"
)

// Workflow 返回工作流服务实例（门面方法）。
// 通过 Goravel 服务容器解析 contracts.Workflow 接口的绑定实现，
// 供调用方以静态方式直接使用工作流引擎的全部能力（发起、审批、驳回、撤回、加签、转交等）。
// 若容器解析失败，将打印错误日志并返回 nil。
func Workflow() contracts.Workflow {
	// 从服务容器中获取工作流绑定的实例
	instance, err := workflow.App.Make(workflow.Binding)
	if err != nil {
		// 容器解析失败时输出错误日志，便于排查服务注册问题
		log.Println(err)
		return nil
	}

	// 类型断言：将 interface{} 转换为 contracts.Workflow 接口类型后返回
	return instance.(contracts.Workflow)
}
