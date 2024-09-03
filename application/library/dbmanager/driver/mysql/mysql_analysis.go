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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	dl "github.com/admpub/go-download"
	"github.com/admpub/log"
	"github.com/admpub/nging/v5/application/handler"
	"github.com/admpub/nging/v5/application/library/config"
	"github.com/admpub/nging/v5/application/library/notice"
	md2html "github.com/russross/blackfriday"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
)

func (m *mySQL) Analysis() error {
	sql := m.Form(`sql`)
	data := m.Data()
	if len(sql) == 0 {
		data.SetInfo(m.T(`请输入SQL语句`), 0)
		return m.JSON(data)
	}
	command := `soar`
	var extension string
	if com.IsWindows {
		extension = `.exe`
	}
	_, err := exec.LookPath(command + extension)
	if err != nil {
		files := []string{
			filepath.Join(echo.Wd(), `support`, command+`_`+runtime.GOOS+`_`+runtime.GOARCH) + extension,
			filepath.Join(echo.Wd(), `support`, command) + extension,
		}
		for _, support := range files {
			if com.FileExists(support) {
				err = nil
				command = support
				break
			}
		}

		if err != nil {
			gpath := os.Getenv("GOPATH")
			if len(gpath) > 0 {
				command = filepath.Join(gpath, `src/github.com/XiaoMi/soar`, command) + extension
				if com.FileExists(command) {
					err = nil
				}
			}
		}
	} else {
		command += extension
	}
	if err != nil {
		var downloaded bool
		if errors.Is(err, exec.ErrNotFound) {
			if config.FromFile().Extend.GetStore(`dbmanager`).Bool(`downloadSOAR`) {
				downloaded, _ = downloadSOAR(m.Context)
			}
			if !downloaded {
				err = m.E(m.T(`没有找到 soar 命令，请取保已经安装。`))
			}
		}
		if !downloaded {
			data.SetError(err)
			return m.JSON(data)
		}
	}
	charset := m.DbAuth.Charset
	if len(charset) == 0 {
		charset = `utf8`
	}
	params := []string{
		command,
		//`-online-dsn`, url.QueryEscape(m.DbAuth.Username) + `:` + url.QueryEscape(m.DbAuth.Password) + `@tcp(` + m.DbAuth.Host + `)/` + m.dbName + `?timeout=3s&charset=` + charset,
		//`-test-dsn`, url.QueryEscape(m.DbAuth.Username) + `:` + url.QueryEscape(m.DbAuth.Password) + `@tcp(` + m.DbAuth.Host + `)/` + m.dbName + `?timeout=3s&charset=` + charset,
	}
	output := []byte{}
	cmd := com.CreateCmd(params, func(b []byte) error {
		output = append(output, b...)
		return nil
	})
	reader := strings.NewReader(sql)
	cmd.Stdin = reader
	err = cmd.Run()
	if err != nil {
		data.SetError(err)
	} else {
		output = md2html.MarkdownCommon(output)
		data.SetData(com.Bytes2str(output))
	}
	return m.JSON(data)
}

func downloadSOAR(ctx echo.Context) (bool, error) {
	command := `soar`
	if com.IsWindows {
		command += `.exe`
	}
	var downloaded bool
	if runtime.GOARCH == `amd64` || runtime.GOARCH == `x86_64` {
		var extension string
		switch {
		case runtime.GOOS == `darwin`:
			extension = `darwin-amd64`
		case runtime.GOOS == `linux`:
			extension = `linux-amd64`
		case runtime.GOOS == `linux`:
			extension = `windows-amd64 `
		default:
		}
		if len(extension) > 0 {
			fileURL := `https://github.com/XiaoMi/soar/releases/download/0.11.0/soar.` + extension
			savePath := filepath.Join(echo.Wd(), `support`, command)
			var username string
			user := handler.User(ctx)
			if user != nil {
				username = user.Username
			}
			np := notice.NewP(ctx, `downloadSOAR`, username, context.Background()).AutoComplete(true)
			np.Send(`start downloading soar...`, notice.StateSuccess)
			dlCfg := &dl.Options{
				Proxy: notice.DownloadProxyFn(np),
			}
			_, derr := dl.Download(fileURL, savePath, dlCfg)
			if derr != nil {
				derr = fmt.Errorf(`failed to download soar from %s: %v`, fileURL, derr)
				log.Error(derr)
				np.Send(derr.Error(), notice.StateFailure)
			} else {
				os.Chmod(savePath, os.ModeExclusive)
				downloaded = true
				np.Send(`downloading soar successfully`, notice.StateSuccess)
			}
			np.Complete()
		}
	}
	return downloaded, nil
}
