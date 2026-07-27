// Package config 提供工作流引擎的配置定义。
// 该文件在服务提供者注册时自动加载，将工作流配置注入到 Goravel 框架的配置容器中。
// 发布后，宿主应用可通过编辑此文件来覆盖默认的模型映射关系。
package config

import (
	"github.com/goravel/framework/facades"
)

// init 在包初始化时自动执行，向 Goravel 配置容器中注册工作流相关配置。
// 此处定义了工作流引擎所依赖的宿主应用模型映射，允许将内置 Emp/Dept 逻辑
// 绑定到宿主应用实际使用的模型名称上。
func init() {
	// 获取 Goravel 框架的全局配置实例
	config := facades.Config()

	// 注册 workflow 配置组，以 workflow 为键名写入配置容器
	config.Add("workflow", map[string]any{
		// Dept 指定部门表对应的模型名称
		// 默认值为 "Dept"，宿主应用应改为自己实际的部门模型名
		"Dept": "Dept", // 部门关联应用中的模型

		// Emp 指定员工/用户表对应的模型名称
		// 默认值为 "User"，宿主应用应改为自己实际的员工模型名
		"Emp": "User", // 员工关联应用中的模型
	})
}
