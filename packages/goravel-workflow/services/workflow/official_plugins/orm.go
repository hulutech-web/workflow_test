package official_plugins

import (
	"sync"

	"github.com/goravel/framework/contracts/database/driver"
	gormdriver "github.com/goravel/framework/database/driver"
	"github.com/goravel/framework/facades"
	gormio "gorm.io/gorm"
)

var (
	// gormIns 是全局 GORM 数据库实例，用于插件内部直接操作数据库
	gormIns *gormio.DB
	// gormInsOnce 保证 GORM 实例只初始化一次，线程安全
	gormInsOnce sync.Once
)

// BootMS 启动并返回一个独立的 GORM 数据库连接实例
// 该方法通过 Goravel 框架配置获取 MySQL 连接驱动，并使用 sync.Once 确保单例初始化
// 返回的 *gormio.DB 可直接用于插件的数据库操作，绕过 ORM 门面
func BootMS() *gormio.DB {
	gormInsOnce.Do(func() {
		// 从框架配置中获取 MySQL 数据库连接的回调函数
		driverCallback, exist := facades.Config().Get("database.connections.mysql.via").(func() (driver.Driver, error))
		if !exist || driverCallback == nil {
			return
		}
		// 调用回调获取数据库驱动实例
		drv, err := driverCallback()
		if err != nil {
			return
		}
		// 获取底层连接池
		pool := drv.Pool()
		// 通过连接池构建 GORM 实例
		db, _, err := gormdriver.BuildGorm(facades.Config(), nil, pool, "mysql", nil)
		if err != nil {
			return
		}
		// 赋值给全局单例
		gormIns = db
	})
	return gormIns
}
