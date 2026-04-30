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
	"fmt"
	"io/fs"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/admpub/errors"
	"github.com/admpub/log"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"
)

// SQLTempDir sql文件缓存目录获取函数(用于导入导出SQL)
var SQLTempDir = os.TempDir

func TempDir(op string) string {
	dir := filepath.Join(SQLTempDir(), `dbmanager/cache`, op)
	err := com.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Error(err)
	}
	return dir
}

func GetSQLFiles(ctx echo.Context) ([]string, error) {
	dbfile := ctx.Formx(`dbfile`).String()
	var sqlFiles []string
	saveDir := TempDir(OpImport)
	if len(dbfile) > 0 {
		fi, err := os.Stat(dbfile)
		if err != nil {
			return nil, fmt.Errorf(`%w: %s`, err, dbfile)
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
				return nil, fmt.Errorf(`%w: %s`, err, dbfile)
			}
			if len(sqlFiles) == 0 {
				return nil, ctx.NewError(code.DataNotFound, `没有找到扩展名为“.sql”/“.zip”/“.tar.gz”的文件`).SetZone(`dbfile`)
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
				return nil, ctx.NewError(code.DataFormatIncorrect, `只支持扩展名为“.sql”/“.zip”/“.tar.gz”的文件`).SetZone(`dbfile`)
			}

		END:
			sqlFiles = append(sqlFiles, dbfile)
		}
	} else {
		err := ctx.SaveUploadedFiles(`file`, func(fdr *multipart.FileHeader) (string, error) {
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
		if err != nil {
			return nil, err
		}
	}
	return sqlFiles, nil
}
