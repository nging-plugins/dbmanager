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

package postgres

import (
	"fmt"
)

// getDatabases returns a list of all databases
func (p *Postgres) getDatabases() ([]string, error) {
	sqlStr := `SELECT datname FROM pg_database WHERE datistemplate = false AND datallowconn = true ORDER BY datname`
	rows, err := p.db.Query(sqlStr)
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

// createDatabase creates a new PostgreSQL database
func (p *Postgres) createDatabase(dbName, owner, encoding, template string) error {
	sqlStr := `CREATE DATABASE ` + QuoteCol(dbName)
	if len(owner) > 0 {
		sqlStr += ` OWNER ` + QuoteCol(owner)
	}
	if len(encoding) > 0 {
		sqlStr += ` ENCODING ` + QuoteVal(encoding)
	}
	if len(template) > 0 {
		sqlStr += ` TEMPLATE ` + QuoteCol(template)
	}
	_, err := p.db.Exec(sqlStr)
	if err != nil {
		return fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	return nil
}

// dropDatabase drops a PostgreSQL database
func (p *Postgres) dropDatabase(dbName string) error {
	// Terminate existing connections first
	terminateSQL := fmt.Sprintf(
		`SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '%s' AND pid <> pg_backend_pid()`,
		dbName,
	)
	p.db.Exec(terminateSQL)

	sqlStr := `DROP DATABASE ` + QuoteCol(dbName)
	_, err := p.db.Exec(sqlStr)
	if err != nil {
		return fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	return nil
}

// alterDatabase renames a PostgreSQL database
func (p *Postgres) alterDatabase(oldName, newName string) error {
	sqlStr := fmt.Sprintf(`ALTER DATABASE %s RENAME TO %s`, QuoteCol(oldName), QuoteCol(newName))
	_, err := p.db.Exec(sqlStr)
	if err != nil {
		return fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	return nil
}
