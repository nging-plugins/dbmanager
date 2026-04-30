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

package clickhouse

import (
	"fmt"
)

// getDatabases returns a list of all databases
func (c *ClickHouse) getDatabases() ([]string, error) {
	sqlStr := `SHOW DATABASES`
	rows, err := c.db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	defer rows.Close()

	var ret []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		ret = append(ret, v)
	}
	return ret, rows.Err()
}

// createDatabase creates a new ClickHouse database
func (c *ClickHouse) createDatabase(dbName string) error {
	sqlStr := `CREATE DATABASE ` + QuoteCol(dbName)
	_, err := c.db.Exec(sqlStr)
	if err != nil {
		return fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	return nil
}

// dropDatabase drops a ClickHouse database
func (c *ClickHouse) dropDatabase(dbName string) error {
	sqlStr := `DROP DATABASE ` + QuoteCol(dbName)
	_, err := c.db.Exec(sqlStr)
	if err != nil {
		return fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	return nil
}
