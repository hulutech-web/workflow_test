package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260307122440CreatePermissionsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260307122440CreatePermissionsTable) Signature() string {
	return "20260307122440_create_permissions_table"
}

// Up Run the migrations.
func (r *M20260307122440CreatePermissionsTable) Up() error {
	if !facades.Schema().HasTable("permissions") {
		return facades.Schema().Create("permissions", func(table schema.Blueprint) {
			table.ID()
			table.String("name").Comment("权限标识")
			table.String("code").Comment("权限标识")
			table.Integer("type").Comment("权限类型: 1-菜单，2-按钮，3-API")
			table.String("description").Comment("描述信息")
			table.Integer("menu_id").Comment("菜单ID")
			table.TimestampsTz()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260307122440CreatePermissionsTable) Down() error {
	return facades.Schema().DropIfExists("permissions")
}
