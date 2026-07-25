package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260307122519CreateMenusTable struct{}

// Signature The unique signature for the migration.
func (r *M20260307122519CreateMenusTable) Signature() string {
	return "20260307122519_create_menus_table"
}

// Up Run the migrations.
func (r *M20260307122519CreateMenusTable) Up() error {
	if !facades.Schema().HasTable("menus") {
		return facades.Schema().Create("menus", func(table schema.Blueprint) {
			table.ID()
			table.BigInteger("pid").Default(0)   // 父级 ID
			table.String("name")                 // 路由名称
			table.String("path")                 // 路由路径
			table.String("component").Nullable() // 组件路径
			table.String("title")                // 标题
			table.String("icon").Nullable()      // 图标
			table.String("locale")               // 国际化 key
			table.Boolean("requires_auth").Default(true)
			table.Text("roles").Nullable() // JSON 格式：["admin","user"]
			table.Boolean("hide_in_menu").Default(false)
			table.Integer("order").Default(0)
			table.Integer("sort").Default(0)
			table.String("target").Comment("目标").Nullable()
			table.String("badge").Comment("角标").Nullable()
			table.TimestampsTz()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260307122519CreateMenusTable) Down() error {
	return facades.Schema().DropIfExists("menus")
}
