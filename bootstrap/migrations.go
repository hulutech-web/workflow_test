package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260307122615CreateUsersTable{},
		&migrations.M20260307122440CreatePermissionsTable{},
		&migrations.M20260307122519CreateMenusTable{},
		&migrations.M20240624000000CreateWorkflowBaseTables{},
	}
}
