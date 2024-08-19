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
	"database/sql"
	"strings"
)

func (m *mySQL) listPrivileges() (bool, []map[string]string, error) {
	var sqlStr string
	var otherFields []string
	if len(m.dbName) == 0 {
		sqlStr = `SELECT User, Host, plugin FROM mysql.user`
		otherFields = append(otherFields, `plugin`)
	} else {
		sqlStr = "SELECT User, Host FROM mysql.db WHERE Db LIKE " + quoteVal(m.dbName)
	}
	sqlStr += " ORDER BY Host, User"
	res, err := m.kvVal(sqlStr, otherFields...)
	sysUser := true
	if err != nil || res == nil || len(res) == 0 {
		sysUser = false
		sqlStr = `SELECT SUBSTRING_INDEX(CURRENT_USER, '@', 1) AS User, SUBSTRING_INDEX(CURRENT_USER, '@', -1) AS Host, 'mysql_native_password' AS plugin`
		res, err = m.kvVal(sqlStr, `plugin`)
	}
	return sysUser, res, err
}

func (m *mySQL) getAuthPluginByUser(user, host string) (string, error) {
	sqlStr := "SELECT plugin FROM mysql.user WHERE User=" + quoteVal(user) + " AND Host=" + quoteVal(host) + " LIMIT 1"
	rows, err := m.newParam().SetCollection(sqlStr).Query()
	if err != nil {
		return ``, err
	}
	defer rows.Close()
	var plugin string
	for rows.Next() {
		err = rows.Scan(&plugin)
	}
	return plugin, err
}

func (m *mySQL) showPrivileges() (*Privileges, error) {
	r := NewPrivileges()
	sqlStr := "SHOW PRIVILEGES"
	rows, err := m.newParam().SetCollection(sqlStr).Query()
	if err != nil {
		return r, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	n := len(cols)
	for rows.Next() {
		v := &Privilege{}
		err = safeScan(rows, n, &v.Privilege, &v.Context, &v.Comment)
		if err != nil {
			break
		}
		r.Privileges = append(r.Privileges, v)
	}
	return r, err
}

func (m *mySQL) showAuthPlugins() ([]*Plugin, error) {
	var r []*Plugin
	sqlStr := "SHOW PLUGINS"
	rows, err := m.newParam().SetCollection(sqlStr).Query()
	if err != nil {
		return r, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return r, err
	}
	n := len(cols)
	var hasNativePassword bool
	for rows.Next() {
		v := &Plugin{}
		err = safeScan(rows, n, &v.Name, &v.Status, &v.Type, &v.Library, &v.License)
		if err != nil {
			break
		}
		if v.Type.String == `AUTHENTICATION` {
			if !hasNativePassword && v.Name.String == `mysql_native_password` {
				hasNativePassword = true
			}
			v.Title = v.Name.String
			if len(v.Status.String) > 0 && v.Status.String != `ACTIVE` {
				v.Title += ` (` + strings.ToLower(v.Status.String) + `)`
			}
			r = append(r, v)
		}
	}
	if !hasNativePassword {
		r = append(r, &Plugin{
			Name:   sql.NullString{String: `mysql_native_password`},
			Status: sql.NullString{String: `DISABLED`},
			Type:   sql.NullString{String: `AUTHENTICATION`},
			Title:  `mysql_native_password (unsupported)`,
		})
	}
	return r, err
}
