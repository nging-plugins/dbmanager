/*
   Nging is a toolbox for webmasters
   Copyright (C) 2019-present  Wenhui Shen <swh@admpub.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package shared

import (
	"context"
	"time"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/background"
	"github.com/coscms/webcore/library/config"
	"github.com/coscms/webcore/library/notice"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/webx-top/echo"
	"github.com/webx-top/echo/defaults"
)

type ImportExecutor func(ctx context.Context, noticer *notice.NoticeAndProgress, cfg *driver.DbAuth, cacheDir string, files []string) error

func ImportSQLFiles(ctx echo.Context, cfg driver.DbAuth, dbName string, charset string, sqlFiles []string, async bool, importExecutor ImportExecutor) error {
	user := backend.User(ctx)
	var username string
	if user != nil {
		username = user.Username
	}
	bgExec := background.New(context.TODO(), echo.H{
		`database`: dbName,
		`sqlFiles`: sqlFiles,
		`async`:    async,
	})
	cacheKey := bgExec.Started.Format(`20060102150405`)
	imports, err := background.Register(ctx, cfg.ImportAndOutputOpName(OpImport), cacheKey, bgExec)
	if err != nil {
		return err
	}
	noticer := notice.NewP(ctx, `databaseImport`, username, bgExec.Context())
	noticer.Success(ctx.T(`文件上传成功`))
	cfg.Db = dbName
	cfg.Charset = charset
	importor := func(c context.Context, noticer *notice.NoticeAndProgress, cfg *driver.DbAuth, cacheDir string, files []string) (err error) {
		if len(files) == 0 {
			return
		}
		return importExecutor(c, noticer, cfg, cacheDir, files)
	}
	if async {
		ctx := defaults.NewMockContext()
		ctx.SetTransaction(ctx.Transaction())
		ctx.SetTranslator(config.FromFile().GetTranslator(ctx))
		go func(ctx echo.Context, cfg driver.DbAuth) {
			done := make(chan error)
			go func(cfg driver.DbAuth) {
				err := importor(bgExec.Context(), noticer, &cfg, TempDir(OpImport), sqlFiles)
				if err != nil {
					noticer.Failure(ctx.T(`导入失败`) + `: ` + err.Error())
					noticer.Complete().Failure(ctx.T(`导入结束 :(`))
				} else {
					noticer.Complete().Success(ctx.T(`导入结束 :)`))
				}
				imports.Cancel(cacheKey)
				done <- err
				close(done)
			}(cfg)
			t := time.NewTicker(24 * time.Hour)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					imports.Cancel(cacheKey)
					return
				case <-done:
					return
				}
			}
		}(ctx, cfg)
		noticer.Success(ctx.T(`正在后台导入，请稍候...`))
	} else {
		done := make(chan struct{})

		go func() {
			ctx := ctx.StdContext()
			defer imports.Cancel(cacheKey)
			for {
				select {
				case <-ctx.Done():
					return
				case <-done:
					return
				}
			}
		}()
		err = importor(bgExec.Context(), noticer, &cfg, TempDir(OpImport), sqlFiles)
		if err != nil {
			noticer.Failure(ctx.T(`导入失败`) + `: ` + err.Error())
		}
		done <- struct{}{}
		close(done)
	}
	return err
}
