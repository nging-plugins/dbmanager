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
	"database/sql"
	"sort"
	"strings"

	"github.com/webx-top/echo"
	"github.com/webx-top/echo/code"

	"github.com/coscms/webcore/library/notice"

	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/mysql/utils"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/shared"
)

func (m *mySQL) getCharsetList() []string {
	e, _ := m.getCharsets()
	cs := make([]string, 0)
	for k := range e {
		cs = append(cs, k)
	}
	sort.Strings(cs)
	return cs
}

func (m *mySQL) exporting() error {
	return m.bgExecManage(shared.OpExport)
}

func (m *mySQL) ImportAndOutputOpName(op string) string {
	return `dbmanager.` + m.DbAuth.Driver + `.` + op
}

func (m *mySQL) bgExecManage(op string) error {
	var err error
	if m.IsPost() {
		shared.BackgroundExecManage(m.Context, *m.DbAuth, op, true)
		data := m.Data()
		data.SetInfo(m.T(`操作成功`))
		m.ok(m.T(`操作成功`))
		return m.returnTo(m.GenURL(op) + `&process=1`)
	}
	err = shared.BackgroundExecManage(m.Context, *m.DbAuth, op, false)
	return m.Render(`db/mysql/process_store`, m.checkErr(err))
}

func (m *mySQL) Export() error {
	if len(m.Form(`docType`)) > 0 {
		return m.ExportDoc()
	}
	//fmt.Printf("%#v\n", m.getCharsetList())
	process := m.Queryx(`process`).Bool()
	if process {
		return m.exporting()
	}
	var err error
	if m.IsPost() {
		var tables []string
		if len(m.dbName) == 0 {
			m.fail(m.T(`请选择数据库`))
			return m.returnTo(m.GenURL(`listDb`))
		}
		if m.Formx(`all`).Bool() {
			var ok bool
			tables, ok = m.Get(`tableList`).([]string)
			if !ok {
				tables, err = m.getTables()
				if err != nil {
					m.fail(err.Error())
					return m.returnTo(m.GenURL(`export`))
				}
			}
		} else {
			tables = m.FormValues(`table`)
			if len(tables) == 1 && len(tables[0]) > 0 {
				tables = strings.Split(tables[0], `,`)
			}
			views := m.FormValues(`view`)
			if len(views) == 1 && len(views[0]) > 0 {
				views = strings.Split(views[0], `,`)
			}
			if len(views) > 0 {
				tables = append(tables, views...)
			}
		}
		coll, err := m.getCollation(m.dbName, nil)
		if err != nil {
			return err
		}
		charset := strings.SplitN(coll, `_`, 2)[0]

		exportExecutor := func(c context.Context, noticer notice.Noticer, cfg *driver.DbAuth, tables []string, structWriter, dataWriter interface{}) error {
			if utils.SupportedExport() { // 采用 mysqldump 命令导出
				gtidMode, _ := m.showVariables(`gtid_mode`)
				var hasGTID bool
				if len(gtidMode) > 0 && len(gtidMode[0]) > 0 {
					if k, y := gtidMode[0][`k`]; y && len(k) > 0 {
						hasGTID = true
					}
				}
				return utils.Export(c, noticer, cfg, tables, structWriter, dataWriter, m.getVersion(), hasGTID, true)
			}
			var err error
			if structWriter != nil {
				err = m.exportDBStruct(c, noticer, cfg, tables, structWriter, m.getVersion(), true)
			}
			if err == nil && dataWriter != nil {
				err = m.exportDBData(c, noticer, cfg, tables, dataWriter, m.getVersion())
			}
			return err
		}

		var data echo.Data
		data, err = shared.ExportSQLFiles(m.Context, *m.DbAuth, m.dbName, charset, tables, exportExecutor)
		if err != nil {
			return err
		}

		if data == nil {
			return err
		}

		return m.JSON(data)
	}
	return m.Redirect(m.GenURL(`listTable`, m.dbName))
}

func (m *mySQL) ExportDoc() error {
	if m.IsPost() {
		var tables []string
		if len(m.dbName) == 0 {
			m.fail(m.T(`请选择数据库`))
			return m.returnTo(m.GenURL(`listDb`))
		}
		tables = m.FormValues(`table`)
		if len(tables) == 1 && len(tables[0]) > 0 {
			tables = strings.Split(tables[0], `,`)
		}
		docType := m.Form(`docType`)
		newExportorDoc, ok := docExportors[docType]
		if !ok {
			return m.NewError(code.InvalidParameter, `不支持导出文档类型: %s`, docType)
		}
		exportor := newExportorDoc(m.dbName)
		err := exportor.Open(m.Context)
		if err != nil {
			return err
		}
		for _, table := range tables {
			origFields, sortFields, err := m.tableFields(table)
			if err != nil {
				return err
			}
			stt, _, err := m.getTableStatus(m.dbName, table, false)
			if err != nil {
				return err
			}
			var tableStatus *TableStatus
			if ts, ok := stt[table]; ok {
				tableStatus = ts
			} else {
				tableStatus = &TableStatus{Name: sql.NullString{Valid: true, String: table}}
			}
			postFields := make([]*Field, len(sortFields))
			for k, v := range sortFields {
				postFields[k] = origFields[v]
			}
			err = exportor.Write(m.Context, tableStatus, postFields)
			if err != nil {
				return err
			}
		}
		return exportor.Close(m.Context)
	}
	return m.Redirect(m.GenURL(`listTable`, m.dbName))
}

type DocExportor interface {
	Open(echo.Context) error
	Write(echo.Context, *TableStatus, []*Field) error
	Close(echo.Context) error
}

var docExportors = map[string]func(dbName string) DocExportor{
	`html`:     newHTMLDocExportor,
	`markdown`: newMarkdownDocExportor,
	`csv`:      newCSVDocExportor,
}
