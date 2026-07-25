package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260308235913CreateUserRolesTable struct{}

// Signature The unique signature for the migration.
func (r *M20260308235913CreateUserRolesTable) Signature() string {
	return "20260308235913_create_user_roles_table"
}

// Up Run the migrations.
func (r *M20260308235913CreateUserRolesTable) Up() error {
	if !facades.Schema().HasTable("user_roles") {
		return facades.Schema().Create("user_roles", func(table schema.Blueprint) {
			table.BigInteger("user_id")
			table.BigInteger("role_id")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260308235913CreateUserRolesTable) Down() error {
	return facades.Schema().DropIfExists("user_roles")
}
