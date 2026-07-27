// Package official_plugins 官方插件包，提供工作流引擎的内置插件实现。
// 包含数据二次分配插件（DistributePlugin），用于在工作流审批过程中
// 根据预设规则将表单数据按比例分配给不同的审批人。
package official_plugins

import (
	"errors"
	"fmt"
	"github.com/goravel/framework/facades"
	"sync"
)

var (
	// Once 确保 AutoMigrate 中的数据库初始化操作仅执行一次，防止并发重复创建表和记录。
	// Once ensures the database initialization in AutoMigrate runs only once, preventing duplicate table/record creation during concurrent access.
	Once sync.Once
)

// DistributePlugin 数据二次分配插件结构体。
// 用于在流程设计时为指定节点绑定数据分配规则，审批通过时自动执行数据二次分配逻辑。
// 分配任务插件
type DistributePlugin struct {
	// HookName 插件注册时绑定的钩子名称，用于与工作流引擎的钩子系统关联。
	HookName string
}

// NewDistributePlugin 创建并返回一个新的 DistributePlugin 实例。
// 这是插件的构造函数，负责初始化插件对象。
func NewDistributePlugin() *DistributePlugin {
	return &DistributePlugin{}
}

// Register 注册插件到工作流引擎。
// 该方法由插件管理器调用，负责完成插件的路由注册和钩子绑定，
// 返回插件的唯一标识名称 "distribute_plugin"。
func (c *DistributePlugin) Register() string {
	fmt.Println("register distribute_plugin called")
	// 分配数据插件 - 获取应用实例并注册插件的 API 路由
	app := facades.App()
	c.RouteApi(app)
	return "distribute_plugin"
}

// Action 返回插件的启动回调函数。
// 该回调在插件向工作流引擎注册后被调用，接收插件的钩子名称作为参数，
// 完成钩子绑定和数据库自动迁移两个初始化步骤。
// 返回的函数签名为 func(string) error。
func (c *DistributePlugin) Action() func(string) error {
	fmt.Println("distribute plugin action called")
	return func(task string) error {
		// 将传入的钩子名称绑定到当前插件实例
		c.AddHook(task)
		// 执行数据库自动迁移，确保插件所需的表结构和初始数据存在
		return c.AutoMigrate()
	}
}

// AddHook 设置插件的钩子名称。
// 将外部传入的钩子标识字符串保存到插件的 HookName 字段，
// 用于后续 Execute 方法中判断当前触发的是否为本插件。
func (c *DistributePlugin) AddHook(hook string) {
	c.HookName = hook
}

// AutoMigrate 执行数据库自动迁移，初始化插件所需的表和默认数据。
// 仅在程序生命周期内执行一次（通过 sync.Once 保证），
// 包括创建 Plugin、PluginConfig、FlowPlugin 三张表，
// 并插入默认的"数据二次分配"插件记录。
// 返回迁移过程中遇到的错误，成功则返回 nil。
func (c *DistributePlugin) AutoMigrate() error {
	err_ := errors.New("")
	// 使用 sync.Once 确保并发安全，所有表创建和初始数据插入仅执行一次
	Once.Do(func() {
		// 获取数据库连接实例（插件模块专用）
		orm := BootMS()

		// 检查并创建 Plugin 表，存储插件基本信息
		if !orm.Migrator().HasTable(&Plugin{}) {
			err_ = orm.AutoMigrate(&Plugin{})
		}
		// 检查并创建 PluginConfig 表，存储插件配置参数
		if !orm.Migrator().HasTable(&PluginConfig{}) {
			err_ = orm.AutoMigrate(&PluginConfig{})
		}
		// 检查并创建 FlowPlugin 表，存储流程与插件的绑定关系
		if !orm.Migrator().HasTable(&FlowPlugin{}) {
			err_ = orm.AutoMigrate(&FlowPlugin{})
		}

		// 插入或获取默认的"数据二次分配"插件记录
		// 如果记录已存在则不重复创建（FirstOrCreate）
		row := orm.FirstOrCreate(&Plugin{
			Name:    "数据二次分配",
			Version: "v1.0",
			Status:  1,
			Description: "1、设计流程时，将某一个数字类型的字段绑定插件，" +
				"2、节点设计规则，在某一个节点“如：主管审批”，规则为：员工1：500元，员工2：1000元，绑定该字段进行二次分配，下一个节点的审批人获取到该规则，" +
				"3、数据查看，下一节点审批人，根据规则获取到自身可以查看的规则内容，员工1：看到奖励500元，员工2：看到奖励1000元",
			Author: "hulu-web",
		})
		if row.RowsAffected == 0 || row.Error != nil {
			// 记录已存在或插入出错时，记录错误信息
			err_ = row.Error
		} else {
			fmt.Println("AutoMigrate successful")
		}
	})
	return err_
}

// Execute 插件执行方法，当流程流转到某一个节点的审批操作完成时，
// 工作流引擎自动调用该方法，根据预设规则执行数据二次分配逻辑，将数据交给下一级审批人。
// 插件执行方法，当流程执行到某一个流程的某一个节点，会自动调用该执行方法，将数据交给下一级
//
// 参数:
//   - plugin_name: 插件名称，用于判断是否为当前插件应响应的调用
//   - args: 可变参数，期望包含 flow_id (uint) 和 process_id (uint)
//
// 返回值: 执行过程中遇到的错误，无错误则返回 nil
func (c *DistributePlugin) Execute(plugin_name string, args ...interface{}) error {
	// 当当前节点执行时，先查询该 flowID 和 processID 中是否存在数据，如果存在，则将 flowID 对应的 entry_data 中的
	// 扩展字段找出，并应用执行方案

	// 仅当传入的 plugin_name 与当前插件绑定的钩子名称匹配时才执行具体逻辑
	if plugin_name != c.HookName {
		return nil
	} else {
		// 从 args 中获取 flow_id 和 process_id
		// 从args中获取flow_id,process_id
		flow_id := args[0].(uint)
		process_id := args[1].(uint)
		fmt.Println("distribute_plugin execute"+",flow_id:", flow_id, ",process_id:", process_id)
	}
	return nil
}
