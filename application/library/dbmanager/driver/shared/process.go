package shared

import (
	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/background"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/webx-top/echo"
)

func BackgroundExecManage(ctx echo.Context, cfg driver.DbAuth, op string, cancel bool) error {
	var err error
	if cancel {
		keys := ctx.FormValues(`key`)
		background.Cancel(cfg.ImportAndOutputOpName(op), keys...)
		return err
	}
	ctx.Set(`op`, op)
	group := background.ListBy(cfg.ImportAndOutputOpName(op))
	bgs := map[string]*background.Background{}
	if group != nil {
		bgs = group.Map()
	}
	ctx.Set(`list`, bgs)
	var title string
	if op == OpExport {
		title = ctx.T(`导出SQL`)
	} else {
		title = ctx.T(`导入SQL`)
	}
	ctx.Set(`title`, title)
	ctx.Set(`cacheDir`, backend.URLFor(`/download/file?path=dbmanager/cache/`+op))
	return err
}
