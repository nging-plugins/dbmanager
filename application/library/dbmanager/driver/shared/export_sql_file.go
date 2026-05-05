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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/admpub/log"
	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/background"
	"github.com/coscms/webcore/library/notice"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
)

type ExportExecutor func(ctx context.Context, noticer notice.Noticer,
	cfg *driver.DbAuth, tables []string, structWriter, dataWriter interface{}) error

func ExportSQLFiles(ctx echo.Context, cfg driver.DbAuth, dbName string, charset string, tables []string, exportExecutor ExportExecutor) (echo.Data, error) {
	cfg.Db = dbName
	cfg.Charset = charset
	output := ctx.Form(`output`)
	types := ctx.FormValues(`type`)
	cacheKey := com.Md5(com.Dump([]interface{}{tables, output, types}, false))
	var (
		structWriter, dataWriter interface{}
		sqlFiles                 []string
		dbSaveDir                string
		sqlSaveDir               string
		async                    bool
		bgExec                   = background.New(context.TODO(), echo.H{
			`database`: dbName,
			`tables`:   tables,
			`output`:   output,
			`types`:    types,
		})
		fileInfos = &FileInfos{}
	)
	exports, err := background.Register(ctx, cfg.ImportAndOutputOpName(OpExport), cacheKey, bgExec)
	if err != nil {
		return nil, err
	}
	nowTime := time.Now().Format("20060102150405.000")
	saveDir := TempDir(OpExport)
	switch output {
	case `down`:
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
		ctx.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", dbName+"-"+nowTime+".sql"))
		fallthrough
	case `open`:
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
		if com.InSlice(`struct`, types) {
			structWriter = ctx.Response()
		}
		if com.InSlice(`data`, types) {
			dataWriter = ctx.Response()
		}
	default:
		async = true
		dbSaveDir = filepath.Join(saveDir, dbName)
		sqlSaveDir = filepath.Join(dbSaveDir, nowTime)
		com.MkdirAll(sqlSaveDir, os.ModePerm)
		if com.InSlice(`struct`, types) {
			structFile := filepath.Join(sqlSaveDir, `struct-`+nowTime+`.sql`)
			sqlFiles = append(sqlFiles, structFile)
			structWriter = structFile
			fi := &FileInfo{
				Start: time.Now(),
				Path:  structFile,
			}
			*fileInfos = append(*fileInfos, fi)
		}
		if com.InSlice(`data`, types) {
			dataFile := filepath.Join(sqlSaveDir, `data-`+nowTime+`.sql`)
			sqlFiles = append(sqlFiles, dataFile)
			dataWriter = dataFile
			fi := &FileInfo{
				Start: time.Now(),
				Path:  dataFile,
			}
			*fileInfos = append(*fileInfos, fi)
		}
	}
	user := backend.User(ctx)
	var username string
	if user != nil {
		username = user.Username
	}
	noticer := notice.New(ctx, `databaseExport`, username, bgExec.Context())

	worker := func(c context.Context, cfg driver.DbAuth) error {
		defer func() {
			exports.Cancel(cacheKey)
			if r := recover(); r != nil {
				err = fmt.Errorf(`RECOVER: %v`, r)
			}
		}()
		err := exportExecutor(ctx, noticer, &cfg, tables, structWriter, dataWriter)
		if err != nil {
			log.Error(err)
			return err
		}
		if len(sqlFiles) > 0 {
			now := time.Now()
			for _, fi := range *fileInfos {
				fi.End = now
				fi.Size, err = com.FileSize(fi.Path)
				if err != nil {
					fi.Error = err.Error()
				}
				fi.Elapsed = fi.End.Sub(fi.Start)
			}
			zipFile := filepath.Join(dbSaveDir, dbName+"-"+nowTime+".zip")
			fi := &FileInfo{
				Start:      now,
				Path:       zipFile,
				Compressed: true,
			}
			fi.Size, err = com.Zip(sqlSaveDir, zipFile)
			if err != nil {
				log.Error(err)
				return err
			}
			os.RemoveAll(sqlSaveDir)
			fi.End = time.Now()
			fi.Elapsed = fi.End.Sub(fi.Start)
			fileInfos.Add(fi)
			os.WriteFile(zipFile+`.txt`, com.Str2bytes(com.Dump(fileInfos, false)), os.ModePerm)
		}
		return nil
	}
	if !async {
		done := make(chan struct{})
		go func() {
			ctx := ctx.StdContext()
			defer exports.Cancel(cacheKey)
			for {
				select {
				case <-ctx.Done():
					return
				case <-done:
					return
				}
			}
		}()
		err = worker(bgExec.Context(), cfg)
		if err != nil {
			noticer(ctx.T(`导出失败`)+`: `+err.Error(), 0)
		}
		done <- struct{}{}
		return nil, err
	}
	go worker(bgExec.Context(), cfg)
	data := ctx.Data()
	data.SetInfo(ctx.T(`任务已经在后台成功启动`))
	data.SetURL(backend.URLFor(`/download/file?path=dbmanager/cache/` + OpExport + `/` + dbName))
	return data, err
}
