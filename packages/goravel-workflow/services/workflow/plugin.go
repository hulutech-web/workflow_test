package workflow

import (
	"sync"
)

// Plugin 定义工作流插件的标准接口。
// 每个插件需要实现注册、动作执行、钩子添加和自定义执行逻辑。
type Plugin interface {
	// Register 注册插件，返回插件名称
	Register() string

	// Action 返回插件注册时执行的动作函数，接收插件名称作为参数
	Action() func(string) error

	// AddHook 为插件添加一个钩子标识
	AddHook(hook string)

	// Execute 执行插件自定义逻辑，接收插件名称和可变参数
	Execute(plugin_name string, args ...interface{}) error
}

// Collector 插件收集器，负责管理所有已注册的插件及其钩子。
// 使用互斥锁保证并发安全。
type Collector struct {
	hooks   []string // 已注册的钩子列表
	mutex   sync.Mutex
	plugins []Plugin // 已加载的插件实例列表
}

// collector 全局单例插件收集器实例
var collector *Collector

// NewCollector 创建或返回全局唯一的插件收集器实例（单例模式）。
// 如果已存在则直接返回，否则使用传入的插件列表初始化。
func NewCollector(plugins []Plugin) *Collector {
	if collector == nil {
		collector = &Collector{plugins: plugins}
	}
	return collector
}

// GetCollectorIns 返回全局插件收集器单例实例。
func GetCollectorIns() *Collector {
	return collector
}

// Boot 启动插件收集器，执行插件初始化逻辑。
func (c *Collector) Boot() {
	// 加载插件
}

// RegisterPlugin 注册并执行指定名称的插件。
// 遍历所有插件，依次调用 Register 注册，再调用 Action 执行插件动作。
// 该方法加锁以保证并发安全。
func (c *Collector) RegisterPlugin(plugin_name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// 第一阶段：注册所有插件
	for _, plugin := range c.plugins {
		plugin.Register()
	}
	// 第二阶段：执行所有插件的动作函数
	for _, plugin := range c.plugins {
		action := plugin.Action()
		action(plugin_name)
	}
}

// AddHook 向收集器中添加一个钩子标识。
func (c *Collector) AddHook(hook string) {
	c.hooks = append(c.hooks, hook)
}

// DoPluginsExec 遍历所有已注册的插件，依次执行它们的 Execute 方法。
// 执行插件中的Execute方法
// 如果收集器为 nil，则直接返回 nil，避免空指针异常。
// 该方法加锁以保证并发安全。
// 若任一插件执行失败，立即返回错误，不再继续执行后续插件。
func (c *Collector) DoPluginsExec(plugin_name string, args ...interface{}) error {
	if c == nil {
		return nil
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, plugin := range c.plugins {
		if err := plugin.Execute(plugin_name, args...); err != nil {
			return err
		}
	}
	return nil
}
