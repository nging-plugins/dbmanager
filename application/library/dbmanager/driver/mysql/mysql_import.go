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
	"strings"

	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/mysql/utils"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/shared"
)

func (m *mySQL) importing() error {
	return m.bgExecManage(shared.OpImport)
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
		var sqlFiles []string
		sqlFiles, err = shared.GetSQLFiles(m.Context)
		if err != nil {
			return shared.ResponseDropzone(err, m.Context)
		}

		coll, err := m.getCollation(m.dbName, nil)
		if err != nil {
			return shared.ResponseDropzone(err, m.Context)
		}
		charset := strings.SplitN(coll, `_`, 2)[0]
		var importor shared.ImportExecutor
		if utils.SupportedImport() { // 采用 mysql 命令导入
			importor = utils.Import
		} else {
			importor = m.importDB
		}
		async := m.Formx(`async`, `true`).Bool()
		err = shared.ImportSQLFiles(m.Context, *m.DbAuth, m.dbName, charset, sqlFiles, async, importor)
		return shared.ResponseDropzone(err, m.Context)
	}

	return m.Render(`db/mysql/import`, m.checkErr(err))
}
