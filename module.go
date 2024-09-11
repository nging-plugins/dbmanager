package dbmanager

import (
	"github.com/coscms/webcore/library/cron"
	"github.com/coscms/webcore/library/module"

	"github.com/nging-plugins/dbmanager/application/handler"
	"github.com/nging-plugins/dbmanager/application/library/setup"
)

const ID = `db`

var Module = module.Module{
	TemplatePath: map[string]string{
		ID: `dbmanager/template/backend`,
	},
	AssetsPath: []string{
		`dbmanager/public/assets`,
	},
	SQLCollection: setup.RegisterSQL,
	Navigate:      RegisterNavigate,
	Route:         handler.RegisterRoute,
	CronJobs: []*cron.Jobx{
		{
			Name:         `mysql_schema_sync`,
			RunnerGetter: handler.SchemaSyncJob,
			Example:      `>mysql_schema_sync:1`,
			Description:  `同步MySQL数据表结构`,
		},
	},
	DBSchemaVer: 0.0000,
}
