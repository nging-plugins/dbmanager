/*
Nging is a toolbox for webmasters
Copyright (C) 2018-present  Wenhui Shen <swh@admpub.com>

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

package mysql

import (
	"context"
	"fmt"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/admpub/errors"
	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"
	"github.com/webx-top/echo/defaults"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/background"
	"github.com/coscms/webcore/library/config"
	"github.com/coscms/webcore/library/notice"
	"github.com/coscms/webcore/library/respond"

	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/mysql/utils"
)

func responseDropzone(err error, ctx echo.Context) error {
	if err != nil {
		if user := backend.User(ctx); user != nil {
			notice.OpenMessage(user.Username, `upload`)
			notice.Send(user.Username, notice.NewMessageWithValue(`upload`, ctx.T(`文件上传出错`), err.Error(), notice.StateFailure))
		}
	}
	return respond.Dropzone(ctx, err, nil)
}

func (m *mySQL) importing() error {
	return m.bgExecManage(utils.OpImport)
}

func (m *mySQL) Import() error {
	process := m.Queryx(`process`).Bool()
	if process {
		return m.importing()
	}
	var err error
	if m.IsPost() {
		if len(m.dbName) == 0 {
			m.fail(m.T(`请选择数据库`))
			return m.returnTo(m.GenURL(`listDb`))
		}
		user := backend.User(m.Context)
		var username string
		if user != nil {
			username = user.Username
		}
		async := m.Formx(`async`, `true`).Bool()
		dbfile := m.Formx(`dbfile`).String()
		var sqlFiles []string
		saveDir := TempDir(utils.OpImport)
		if len(dbfile) > 0 {
			var fi fs.FileInfo
			fi, err = os.Stat(dbfile)
			if err != nil {
				return fmt.Errorf(`%w: %s`, err, dbfile)
			}
			if fi.IsDir() {
				err = filepath.Walk(dbfile, func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if info.IsDir() {
						return filepath.SkipDir
					}
					extension := strings.ToLower(filepath.Ext(info.Name()))
					switch extension {
					case `.sql`:
					case `.zip`:
					case `.gz`:
						if !strings.HasSuffix(info.Name(), `.tar.gz`) {
							return nil
						}
					default:
						return nil
					}

					sqlFiles = append(sqlFiles, path)
					return nil
				})
				if err != nil {
					return fmt.Errorf(`%w: %s`, err, dbfile)
				}
				if len(sqlFiles) == 0 {
					return m.NewError(code.DataNotFound, `没有找到扩展名为“.sql”/“.zip”/“.tar.gz”的文件`).SetZone(`dbfile`)
				}
			} else {
				extension := strings.ToLower(filepath.Ext(dbfile))
				switch extension {
				case `.sql`:
				case `.zip`:
				case `.gz`:
					if strings.HasSuffix(dbfile, `.tar.gz`) {
						goto END
					}
					fallthrough
				default:
					return m.NewError(code.DataFormatIncorrect, `只支持扩展名为“.sql”/“.zip”/“.tar.gz”的文件`).SetZone(`dbfile`)
				}

			END:
				sqlFiles = append(sqlFiles, dbfile)
			}
		} else {
			err = m.SaveUploadedFiles(`file`, func(fdr *multipart.FileHeader) (string, error) {
				extension := strings.ToLower(filepath.Ext(fdr.Filename))
				switch extension {
				case `.sql`:
				case `.zip`:
				case `.gz`:
					if strings.HasSuffix(fdr.Filename, `.tar.gz`) {
						goto END
					}
					fallthrough
				default:
					return ``, errors.New(`只能上传扩展名为“.sql”/“.zip”/“.tar.gz”的文件`)
				}

			END:
				sqlFile := filepath.Join(saveDir, fdr.Filename)
				sqlFiles = append(sqlFiles, sqlFile)
				return sqlFile, nil
			})
		}
		if err != nil {
			return responseDropzone(err, m.Context)
		}
		bgExec := background.New(context.TODO(), echo.H{
			`database`: m.dbName,
			`sqlFiles`: sqlFiles,
			`async`:    async,
		})
		cacheKey := bgExec.Started.Format(`20060102150405`)
		imports, err := background.Register(m.Context, m.ImportAndOutputOpName(utils.OpImport), cacheKey, bgExec)
		if err != nil {
			return responseDropzone(err, m.Context)
		}
		noticer := notice.NewP(m.Context, `databaseImport`, username, bgExec.Context())
		noticer.Success(m.T(`文件上传成功`))
		coll, err := m.getCollation(m.dbName, nil)
		if err != nil {
			return responseDropzone(err, m.Context)
		}
		cfg := *m.DbAuth
		cfg.Db = m.dbName
		cfg.Charset = strings.SplitN(coll, `_`, 2)[0]
		importor := func(c context.Context, noticer *notice.NoticeAndProgress, cfg *driver.DbAuth, cacheDir string, files []string) (err error) {
			if len(files) == 0 {
				return
			}
			if utils.SupportedImport() { // 采用 mysql 命令导入
				err = utils.Import(c, noticer, cfg, cacheDir, files)
			} else {
				err = m.importDB(c, noticer, cfg, cacheDir, files)
			}
			return
		}
		if async {
			ctx := defaults.NewMockContext()
			ctx.SetTransaction(m.Transaction())
			ctx.SetTranslator(config.FromFile().GetTranslator(m.Context))
			go func(ctx echo.Context, cfg driver.DbAuth) {
				done := make(chan error)
				go func(cfg driver.DbAuth) {
					err := importor(bgExec.Context(), noticer, &cfg, TempDir(utils.OpImport), sqlFiles)
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
			noticer.Success(m.T(`正在后台导入，请稍候...`))
		} else {
			done := make(chan struct{})
			ctx := m.StdContext()
			go func() {
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
			err = importor(bgExec.Context(), noticer, &cfg, TempDir(utils.OpImport), sqlFiles)
			if err != nil {
				noticer.Failure(m.T(`导入失败`) + `: ` + err.Error())
			}
			done <- struct{}{}
			close(done)
		}
		return responseDropzone(err, m.Context)
	}

	return m.Render(`db/mysql/import`, m.checkErr(err))
}
