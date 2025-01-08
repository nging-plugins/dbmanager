package handler

import (
	"github.com/coscms/webcore/library/navigate"
	"github.com/webx-top/echo"
)

var LeftNavigate = &navigate.Item{
	Display: true,
	Name:    echo.T(`数据库`),
	Action:  `db`,
	Icon:    `table`,
	Children: &navigate.List{
		{
			Display: true,
			Name:    echo.T(`数据库账号`),
			Action:  `account`,
		},
		{
			Display: true,
			Name:    echo.T(`添加账号`),
			Action:  `account_add`,
			Icon:    `plus`,
		},
		{
			Display: false,
			Name:    echo.T(`修改账号`),
			Action:  `account_edit`,
		},
		{
			Display: false,
			Name:    echo.T(`删除账号`),
			Action:  `account_delete`,
		},
		{
			Display: true,
			Name:    echo.T(`连接数据库`),
			Action:  ``,
		},
		{
			Display: true,
			Name:    echo.T(`表结构同步`),
			Action:  `schema_sync`,
		},
		{
			Display: false,
			Name:    echo.T(`新增同步方案`),
			Action:  `schema_sync_add`,
		},
		{
			Display: false,
			Name:    echo.T(`修改同步方案`),
			Action:  `schema_sync_edit`,
		},
		{
			Display: false,
			Name:    echo.T(`删除同步方案`),
			Action:  `schema_sync_delete`,
		},
		{
			Display: false,
			Name:    echo.T(`预览表结构差异`),
			Action:  `schema_sync_preview`,
		},
		{
			Display: false,
			Name:    echo.T(`执行表结构同步`),
			Action:  `schema_sync_run`,
		},
		{
			Display: false,
			Name:    echo.T(`表结构同步日志列表`),
			Action:  `schema_sync_log/:id`,
		},
		{
			Display: false,
			Name:    echo.T(`表结构同步日志详情`),
			Action:  `schema_sync_log_view/:id`,
		},
		{
			Display: false,
			Name:    echo.T(`删除表结构同步日志`),
			Action:  `schema_sync_log_delete`,
		},
	},
}
