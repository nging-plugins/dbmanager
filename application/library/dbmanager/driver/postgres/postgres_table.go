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
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coscms/webcore/library/backend"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/shared"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
	"github.com/webx-top/pagination"
)

// getTables returns a list of tables in the current database
func (p *Postgres) getTables() ([]string, error) {
	sqlStr := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name`
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

// TableStatus holds table status information for PostgreSQL
type TableStatus struct {
	Name       string
	Schema     string
	Owner      string
	Rows       int64
	Size       string
	SizePretty string
	Comment    string
}

// getTableStatus returns detailed table status information
func (p *Postgres) getTableStatus(tableName string) (map[string]*TableStatus, error) {
	sqlStr := `
		SELECT 
			c.relname AS name,
			n.nspname AS schema,
			pg_get_userbyid(c.relowner) AS owner,
			c.reltuples::bigint AS rows,
			pg_size_pretty(pg_total_relation_size(c.oid)) AS size_pretty,
			pg_total_relation_size(c.oid) AS size,
			obj_description(c.oid) AS comment
		FROM pg_class c
		LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relkind = 'r'
		AND n.nspname = 'public'`

	if len(tableName) > 0 {
		sqlStr += ` AND c.relname = '` + strings.ReplaceAll(tableName, `'`, `''`) + `'`
	}
	sqlStr += ` ORDER BY c.relname`

	rows, err := p.db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	defer rows.Close()

	ret := map[string]*TableStatus{}
	for rows.Next() {
		var (
			name       string
			schema     string
			owner      string
			rowCnt     sql.NullInt64
			sizePretty sql.NullString
			size       sql.NullString
			comment    sql.NullString
		)
		if err := rows.Scan(&name, &schema, &owner, &rowCnt, &sizePretty, &size, &comment); err != nil {
			return nil, err
		}
		v := &TableStatus{
			Name:       name,
			Schema:     schema,
			Owner:      owner,
			Rows:       rowCnt.Int64,
			Size:       size.String,
			SizePretty: sizePretty.String,
			Comment:    comment.String,
		}
		ret[v.Name] = v
	}
	return ret, rows.Err()
}

// FieldInfo holds column information
type FieldInfo struct {
	ColumnName    string
	DataType      string
	IsNullable    string
	ColumnDefault string
	CharacterMax  int64
	Comment       string
}

// getTableFields returns column information for a table
func (p *Postgres) getTableFields(tableName string) ([]*FieldInfo, error) {
	sqlStr := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default,
			character_maximum_length,
			pg_catalog.col_description(
				(SELECT c.oid FROM pg_catalog.pg_class c WHERE c.relname = $1),
				ordinal_position
			) AS comment
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
		ORDER BY ordinal_position`

	rows, err := p.db.Query(sqlStr, tableName)
	if err != nil {
		return nil, fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	defer rows.Close()

	var ret []*FieldInfo
	for rows.Next() {
		var (
			columnName    string
			dataType      string
			isNullable    string
			columnDefault sql.NullString
			charMax       sql.NullInt64
			comment       sql.NullString
		)
		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &charMax, &comment); err != nil {
			return nil, err
		}
		v := &FieldInfo{
			ColumnName:    columnName,
			DataType:      dataType,
			IsNullable:    isNullable,
			ColumnDefault: columnDefault.String,
			CharacterMax:  charMax.Int64,
			Comment:       comment.String,
		}
		ret = append(ret, v)
	}
	return ret, rows.Err()
}

// getTableDDL returns the CREATE TABLE DDL statement
func (p *Postgres) getTableDDL(tableName string) (string, error) {
	// Use pg_dump-like approach: we query the table definition
	sqlStr := `
		SELECT 
			'CREATE TABLE ' || quote_ident(table_name) || ' (' ||
			string_agg(column_def, ', ' ORDER BY ordinal_position) || ');' AS ddl
		FROM (
			SELECT 
				c.table_name,
				c.ordinal_position,
				quote_ident(c.column_name) || ' ' || 
				CASE 
					WHEN c.character_maximum_length IS NOT NULL 
					THEN c.data_type || '(' || c.character_maximum_length || ')'
					ELSE c.data_type
				END ||
				CASE WHEN c.is_nullable = 'NO' THEN ' NOT NULL' ELSE '' END ||
				COALESCE(' DEFAULT ' || c.column_default, '') AS column_def
			FROM information_schema.columns c
			WHERE c.table_schema = 'public' AND c.table_name = $1
		) sub
		GROUP BY table_name`

	var ddl sql.NullString
	row := p.db.QueryRow(sqlStr, tableName)
	if err := row.Scan(&ddl); err != nil {
		return "", fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	if ddl.Valid {
		return ddl.String, nil
	}
	return "", nil
}

// ListDb handles the list databases operation
func (p *Postgres) ListDb() error {
	dbList, err := p.getDatabases()
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(backend.URLFor(`/db`))
	}
	p.Set(`dbList`, dbList)
	return p.Render(`db/postgres/list_db`, p.checkErr(err))
}

// CreateDb handles the create database operation
func (p *Postgres) CreateDb() error {
	if p.IsPost() {
		data := p.Data()
		dbName := p.Form(`db`)
		owner := p.Form(`owner`)
		encoding := p.Form(`encoding`)
		template := p.Form(`template`)

		if len(dbName) == 0 {
			data.SetInfo(p.T(`数据库名不能为空`), 0)
			return p.JSON(data)
		}

		err := p.createDatabase(dbName, owner, encoding, template)
		if err != nil {
			data.SetError(err)
		} else {
			data.SetInfo(p.T(`数据库创建成功`), 1)
		}
		return p.JSON(data)
	}
	return p.Render(`db/postgres/create_db`, p.checkErr(nil))
}

// ModifyDb handles the modify/delete database operation
func (p *Postgres) ModifyDb() error {
	if p.IsPost() {
		data := p.Data()
		operate := p.Form(`operate`)

		switch operate {
		case `drop`:
			dbName := p.Form(`db`)
			if len(dbName) == 0 {
				data.SetInfo(p.T(`数据库名不能为空`), 0)
				return p.JSON(data)
			}
			err := p.dropDatabase(dbName)
			if err != nil {
				data.SetError(err)
			} else {
				data.SetInfo(p.T(`数据库删除成功`), 1)
			}
		case `rename`:
			oldName := p.Form(`oldName`)
			newName := p.Form(`newName`)
			if len(oldName) == 0 || len(newName) == 0 {
				data.SetInfo(p.T(`数据库名不能为空`), 0)
				return p.JSON(data)
			}
			err := p.alterDatabase(oldName, newName)
			if err != nil {
				data.SetError(err)
			} else {
				data.SetInfo(p.T(`数据库重命名成功`), 1)
			}
		default:
			data.SetInfo(p.T(`不支持的操作`), 0)
		}
		return p.JSON(data)
	}
	return p.Render(`db/postgres/modify_db`, p.checkErr(nil))
}

// ListTable handles the list tables operation
func (p *Postgres) ListTable() error {
	tables, err := p.getTables()
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(p.GenURL(`listDb`))
	}
	p.Set(`tableList`, tables)

	// Also get table statuses
	statuses, _ := p.getTableStatus(``)
	p.Set(`tableStatuses`, statuses)

	return p.Render(`db/postgres/list_table`, p.checkErr(err))
}

// ViewTable handles the view table structure operation
func (p *Postgres) ViewTable() error {
	tableName := p.Form(`table`)
	if len(tableName) == 0 {
		p.fail(p.T(`请选择表`))
		return p.returnTo(p.GenURL(`listTable`))
	}

	fields, err := p.getTableFields(tableName)
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(p.GenURL(`listTable`))
	}
	p.Set(`fields`, fields)

	ddl, err := p.getTableDDL(tableName)
	if err == nil {
		p.Set(`ddl`, ddl)
	}

	// Get indexes for inline display
	idxRows, idxErr := p.db.Query(`
		SELECT
			i.relname AS index_name,
			am.amname AS index_type,
			pg_get_indexdef(ix.indexrelid) AS index_def,
			ix.indisunique AS is_unique,
			ix.indisprimary AS is_primary
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_am am ON am.oid = i.relam
		WHERE t.relname = $1
		AND t.relkind = 'r'
		ORDER BY i.relname`, tableName)
	if idxErr == nil {
		defer idxRows.Close()
		type idxInfo struct {
			Name    string
			Columns string
			Unique  bool
		}
		var indexes []idxInfo
		reCols := regexp.MustCompile(`\(([^)]+)\)`)
		for idxRows.Next() {
			var name, idxType, def string
			var isUnique, isPrimary bool
			if err := idxRows.Scan(&name, &idxType, &def, &isUnique, &isPrimary); err != nil {
				continue
			}
			columns := ""
			if matches := reCols.FindStringSubmatch(def); len(matches) > 1 {
				columns = matches[1]
			}
			indexes = append(indexes, idxInfo{Name: name, Columns: columns, Unique: isUnique})
		}
		p.Set(`indexes`, indexes)
	}

	return p.Render(`db/postgres/view_table`, p.checkErr(err))
}

// RunCommand handles running SQL commands
func (p *Postgres) RunCommand() error {
	if p.IsPost() {
		data := p.Data()
		sqlStr := strings.TrimSpace(p.Form(`sql`))
		if len(sqlStr) == 0 {
			data.SetInfo(p.T(`请输入SQL语句`), 0)
			return p.JSON(data)
		}

		upperSQL := strings.ToUpper(sqlStr)

		// Determine if it's a query or exec command
		isQuery := strings.HasPrefix(upperSQL, `SELECT`) ||
			strings.HasPrefix(upperSQL, `SHOW`) ||
			strings.HasPrefix(upperSQL, `EXPLAIN`) ||
			strings.HasPrefix(upperSQL, `WITH`)

		if isQuery {
			rows, err := p.db.Query(sqlStr)
			if err != nil {
				data.SetError(err)
				return p.JSON(data)
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				data.SetError(err)
				return p.JSON(data)
			}

			var results []map[string]interface{}
			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range columns {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					data.SetError(err)
					return p.JSON(data)
				}

				row := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					// Convert byte arrays to strings
					if b, ok := val.([]byte); ok {
						row[col] = string(b)
					} else {
						row[col] = val
					}
				}
				results = append(results, row)
			}

			p.Set(`columns`, columns)
			p.Set(`results`, results)
			p.Set(`rowsAffected`, int64(len(results)))
		} else {
			result, err := p.db.Exec(sqlStr)
			if err != nil {
				data.SetError(err)
				return p.JSON(data)
			}
			rowsAffected, _ := result.RowsAffected()
			p.Set(`rowsAffected`, rowsAffected)
		}

		p.Set(`sql`, sqlStr)
		data.SetInfo(p.T(`执行成功`), 1)
		return p.JSON(data)
	}

	return p.Render(`db/postgres/sql`, p.checkErr(nil))
}

// ListData handles displaying table data (paginated)
func (p *Postgres) ListData() error {
	tableName := p.Form(`table`)
	if len(tableName) == 0 {
		p.fail(p.T(`请选择表`))
		return p.returnTo(p.GenURL(`listTable`))
	}

	limit := p.Formx(`limit`).Int()
	page := p.Formx(`page`).Int()
	totalRows := p.Formx(`rows`).Int()

	if limit < 1 {
		limit = 50
		p.Request().Form().Set(`limit`, strconv.Itoa(limit))
	}
	if page < 1 {
		page = 1
	}

	// Get fields for column display
	fields, err := p.getTableFields(tableName)
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(p.GenURL(`listTable`))
	}
	p.Set(`fields`, fields)

	// Get total row count if not provided
	if totalRows < 1 {
		countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, QuoteCol(tableName))
		var count sql.NullInt64
		p.db.QueryRow(countSQL).Scan(&count)
		totalRows = int(count.Int64)
	}

	offset := (page - 1) * limit

	// Query data
	sqlStr := fmt.Sprintf(`SELECT * FROM %s ORDER BY 1 LIMIT %d OFFSET %d`, QuoteCol(tableName), limit, offset)
	rows, err := p.db.Query(sqlStr)
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(p.GenURL(`listTable`))
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(p.GenURL(`listTable`))
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	p.Set(`columns`, columns)
	p.Set(`values`, results)
	p.Set(`total`, totalRows)
	p.Set(`table`, tableName)

	// Pagination - same pattern as MySQL
	q := p.Request().URL().Query()
	q.Del(`page`)
	q.Del(`rows`)
	q.Del(`_pjax`)
	p.Set(`pagination`, pagination.New(p.Context).SetURL(backend.URLFor(`/db`)+`?`+q.Encode()+`&page={page}&rows={rows}`).SetPage(page).SetRows(totalRows))

	return p.Render(`db/postgres/list_data`, p.checkErr(err))
}

// CreateTable handles the create table operation
func (p *Postgres) CreateTable() error {
	if p.IsPost() {
		tableName := p.Form(`table`)
		if len(tableName) == 0 {
			p.fail(p.T(`表名不能为空`))
			return p.returnTo(p.GenURL(`createTable`, p.dbName))
		}

		columnNames := p.FormValues(`column_name[]`)
		columnTypes := p.FormValues(`column_type[]`)
		columnNulls := p.FormValues(`column_null[]`)
		columnDefaults := p.FormValues(`column_default[]`)

		if len(columnNames) == 0 {
			p.fail(p.T(`请至少定义一个字段`))
			return p.returnTo(p.GenURL(`createTable`, p.dbName))
		}

		var colDefs []string
		for i, colName := range columnNames {
			if len(colName) == 0 || i >= len(columnTypes) {
				continue
			}
			def := QuoteCol(colName) + ` ` + columnTypes[i]
			if i < len(columnNulls) && columnNulls[i] == `NO` {
				def += ` NOT NULL`
			}
			if i < len(columnDefaults) && len(columnDefaults[i]) > 0 {
				def += ` DEFAULT ` + columnDefaults[i]
			}
			colDefs = append(colDefs, def)
		}

		sqlStr := fmt.Sprintf(`CREATE TABLE %s (%s)`, QuoteCol(tableName), strings.Join(colDefs, `, `))
		_, err := p.db.Exec(sqlStr)
		if err != nil {
			p.fail(err.Error())
		} else {
			p.ok(p.T(`表创建成功`))
		}
		return p.returnTo(p.GenURL(`listTable`, p.dbName))
	}

	// Provide data type suggestions
	p.Set(`dataTypes`, []string{
		`integer`, `bigint`, `smallint`, `serial`, `bigserial`,
		`varchar(255)`, `text`, `char(1)`,
		`boolean`,
		`timestamp`, `date`, `time`,
		`numeric(10,2)`, `real`, `double precision`,
		`json`, `jsonb`,
		`uuid`,
		`bytea`,
	})
	return p.Render(`db/postgres/create_table`, p.checkErr(nil))
}

// ModifyTable handles the modify table operation
func (p *Postgres) ModifyTable() error {
	tableName := p.Form(`table`)
	if len(tableName) == 0 {
		p.fail(p.T(`请选择表`))
		return p.returnTo(p.GenURL(`listTable`))
	}

	if p.IsPost() || len(p.Form("operate")) > 0 {
		operate := p.Form(`operate`)

		switch operate {
		case `drop`:
			sqlStr := fmt.Sprintf(`DROP TABLE %s`, QuoteCol(tableName))
			if p.Formx(`cascade`).Bool() {
				sqlStr += ` CASCADE`
			}
			_, err := p.db.Exec(sqlStr)
			if err != nil {
				p.fail(err.Error())
			} else {
				p.ok(p.T(`表删除成功`))
			}
			return p.returnTo(p.GenURL(`listTable`, p.dbName))
		case `truncate`:
			sqlStr := fmt.Sprintf(`TRUNCATE TABLE %s`, QuoteCol(tableName))
			if p.Formx(`cascade`).Bool() {
				sqlStr += ` CASCADE`
			}
			_, err := p.db.Exec(sqlStr)
			if err != nil {
				p.fail(err.Error())
			} else {
				p.ok(p.T(`表已清空`))
			}
			return p.returnTo(p.GenURL(`listTable`, p.dbName))
		case `rename`:
			newName := p.Form(`newName`)
			if len(newName) == 0 {
				p.fail(p.T(`新表名不能为空`))
				return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
			}
			sqlStr := fmt.Sprintf(`ALTER TABLE %s RENAME TO %s`, QuoteCol(tableName), QuoteCol(newName))
			_, err := p.db.Exec(sqlStr)
			if err != nil {
				p.fail(err.Error())
			} else {
				p.ok(p.T(`表重命名成功`))
			}
			return p.returnTo(p.GenURL(`listTable`, p.dbName))
		case `addColumn`:
			colName := p.Form(`columnName`)
			colType := p.Form(`columnType`)
			colDefault := p.Form(`columnDefault`)
			colNotNull := p.Formx(`columnNotNull`).Bool()
			if len(colName) == 0 || len(colType) == 0 {
				p.fail(p.T(`列名和类型不能为空`))
				return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
			}
			sqlStr := fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, QuoteCol(tableName), QuoteCol(colName), colType)
			if colNotNull {
				sqlStr += ` NOT NULL`
			}
			if len(colDefault) > 0 {
				sqlStr += ` DEFAULT ` + colDefault
			}
			_, err := p.db.Exec(sqlStr)
			if err != nil {
				p.fail(err.Error())
			} else {
				p.ok(p.T(`列添加成功`))
			}
			return p.returnTo(p.GenURL(`viewTable`, p.dbName, tableName))
		case `modifyColumn`:
			colName := p.Form(`colName`)
			newColName := p.Form(`newColName`)
			colType := p.Form(`colType`)
			colDefault := p.Form(`colDefault`)
			colNotNull := p.Form(`colNotNull`)
			if len(colName) == 0 {
				p.fail(p.T(`列名不能为空`))
				return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
			}
			// 1. Rename column if newColName is set and different
			if len(newColName) > 0 && newColName != colName {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s RENAME COLUMN %s TO %s`, QuoteCol(tableName), QuoteCol(colName), QuoteCol(newColName))
				_, err := p.db.Exec(sqlStr)
				if err != nil {
					p.fail(err.Error())
					return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
				}
				colName = newColName
			}
			// 2. Change column type
			if len(colType) > 0 {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s TYPE %s`, QuoteCol(tableName), QuoteCol(colName), colType)
				_, err := p.db.Exec(sqlStr)
				if err != nil {
					p.fail(err.Error())
					return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
				}
			}
			// 3. NOT NULL / DROP NOT NULL
			if colNotNull == "1" {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s SET NOT NULL`, QuoteCol(tableName), QuoteCol(colName))
				_, err := p.db.Exec(sqlStr)
				if err != nil {
					p.fail(err.Error())
					return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
				}
			} else if colNotNull == "0" {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL`, QuoteCol(tableName), QuoteCol(colName))
				_, err := p.db.Exec(sqlStr)
				if err != nil {
					p.fail(err.Error())
					return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
				}
			}
			// 4. DEFAULT
			if len(colDefault) > 0 {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s`, QuoteCol(tableName), QuoteCol(colName), colDefault)
				_, err := p.db.Exec(sqlStr)
				if err != nil {
					p.fail(err.Error())
					return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
				}
			}
			p.ok(p.T(`列修改成功`))
			return p.returnTo(p.GenURL(`viewTable`, p.dbName, tableName))
		case `dropColumn`:
			colName := p.Form(`colName`)
			if len(colName) == 0 {
				p.fail(p.T(`列名不能为空`))
				return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
			}
			sqlStr := fmt.Sprintf(`ALTER TABLE %s DROP COLUMN %s`, QuoteCol(tableName), QuoteCol(colName))
			_, err := p.db.Exec(sqlStr)
			if err != nil {
				p.fail(err.Error())
			} else {
				p.ok(p.T(`列删除成功`))
				return p.returnTo(p.GenURL(`viewTable`, p.dbName, tableName))
			}
			return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
		default:
			p.fail(p.T(`不支持的操作`))
			return p.returnTo(p.GenURL(`modifyTable`, p.dbName, tableName))
		}
	}

	fields, err := p.getTableFields(tableName)
	if err == nil {
		p.Set(`fields`, fields)
	}
	return p.Render(`db/postgres/modify_table`, p.checkErr(err))
}

// Indexes handles showing indexes for a table
func (p *Postgres) Indexes() error {
	tableName := p.Form(`table`)
	if len(tableName) == 0 {
		p.fail(p.T(`请选择表`))
		return p.returnTo(p.GenURL(`listTable`))
	}

	// Handle drop/create actions
	act := p.Form(`act`)
	if p.IsPost() || len(act) > 0 {
		switch act {
		case `drop`:
			idxName := p.Form(`name`)
			if len(idxName) == 0 {
				p.fail(p.T(`索引名不能为空`))
			} else {
				sqlStr := fmt.Sprintf(`DROP INDEX IF EXISTS %s`, QuoteCol(idxName))
				_, err := p.db.Exec(sqlStr)
				if err != nil {
					p.fail(err.Error())
				} else {
					p.ok(p.T(`索引 %s 已删除`, idxName))
				}
			}
			return p.returnTo(p.GenURL(`indexes`, p.dbName, tableName))

		case `create`:
			idxName := p.Form(`name`)
			columns := p.Form(`columns`)
			unique := p.Formx(`unique`).Bool()
			if len(idxName) == 0 || len(columns) == 0 {
				p.fail(p.T(`索引名和列名不能为空`))
			} else {
				uniqueStr := ``
				if unique {
					uniqueStr = ` UNIQUE`
				}
				sqlStr := fmt.Sprintf(`CREATE%s INDEX %s ON %s (%s)`, uniqueStr, QuoteCol(idxName), QuoteCol(tableName), columns)
				_, err := p.db.Exec(sqlStr)
				if err != nil {
					p.fail(err.Error())
				} else {
					p.ok(p.T(`索引 %s 创建成功`, idxName))
				}
			}
			return p.returnTo(p.GenURL(`indexes`, p.dbName, tableName))
		}
	}

	sqlStr := `
		SELECT 
			i.relname AS index_name,
			am.amname AS index_type,
			pg_get_indexdef(ix.indexrelid) AS index_def,
			ix.indisunique AS is_unique,
			ix.indisprimary AS is_primary
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_am am ON am.oid = i.relam
		WHERE t.relname = $1
		AND t.relkind = 'r'
		ORDER BY i.relname`

	rows, err := p.db.Query(sqlStr, tableName)
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(p.GenURL(`listTable`))
	}
	defer rows.Close()

	type IndexInfo struct {
		Name      string
		Type      string
		Columns   string
		Unique    bool
		IsPrimary bool
	}

	// extract column names from index definition
	reColumns := regexp.MustCompile(`\(([^)]+)\)`)

	var indexes []IndexInfo
	for rows.Next() {
		var (
			name      string
			idxType   string
			def       string
			isUnique  bool
			isPrimary bool
		)
		if err := rows.Scan(&name, &idxType, &def, &isUnique, &isPrimary); err != nil {
			continue
		}
		columns := ""
		if matches := reColumns.FindStringSubmatch(def); len(matches) > 1 {
			columns = matches[1]
		}
		idx := IndexInfo{
			Name:      name,
			Type:      idxType,
			Columns:   columns,
			Unique:    isUnique,
			IsPrimary: isPrimary,
		}
		indexes = append(indexes, idx)
	}

	p.Set(`indexes`, indexes)
	p.Set(`table`, tableName)
	return p.Render(`db/postgres/indexes`, p.checkErr(err))
}

func (p *Postgres) Export() error {
	var err error
	if p.IsPost() {
		if len(p.dbName) == 0 {
			p.fail(p.T(`请选择数据库`))
			return p.returnTo(p.GenURL(`listDb`))
		}
		var tables []string
		if p.Formx(`all`).Bool() {
			tables, _ = p.getTables()
		} else {
			tables = p.FormValues(`table`)
		}
		if len(tables) == 0 {
			p.fail(p.T(`请选择要导出的表`))
			return p.returnTo(p.GenURL(`listTable`, p.dbName))
		}
		output := p.Form(`output`, `down`)
		types := p.FormValues(`type`)

		host, port := shared.SplitHostPort(p.DbAuth.Host)
		if len(port) == 0 {
			port = `5432`
		}
		cfg := &shared.DBConfig{
			Driver:   `postgres`,
			Host:     host,
			Port:     port,
			Username: p.DbAuth.Username,
			Password: p.DbAuth.Password,
			Database: p.dbName,
		}

		var structWriter, dataWriter io.Writer
		hasStruct := false
		hasData := false
		for _, t := range types {
			if t == `struct` {
				hasStruct = true
			}
			if t == `data` {
				hasData = true
			}
		}
		if !hasStruct && !hasData {
			hasStruct = true
			hasData = true
		}

		switch output {
		case `down`, `open`:
			if output == `down` {
				p.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
				p.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", p.dbName+".sql"))
			} else {
				p.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
			}
			if shared.SupportedExport(`postgres`) {
				w := p.Response()
				if hasStruct {
					structWriter = w
				}
				if hasData {
					dataWriter = w
				}
				err = shared.NativeExportPG(context.Background(), cfg, tables, structWriter, dataWriter)
			} else {
				if hasStruct {
					structWriter = p.Response()
				}
				if hasData {
					dataWriter = p.Response()
				}
				err = shared.ExportMultipleTablesToWriter(p.Response(), p.db, tables, `"`, hasStruct, hasData, p.getVersion())
			}
			if err != nil {
				p.fail(err.Error())
				return p.returnTo(p.GenURL(`listTable`, p.dbName))
			}
			return nil

		default: // send
			saveDir := filepath.Join(os.TempDir(), `dbmanager/cache/export`, p.dbName)
			com.MkdirAll(saveDir, os.ModePerm)
			nowTime := time.Now().Format("20060102150405")
			file := filepath.Join(saveDir, nowTime+`.sql`)
			f, err := os.Create(file)
			if err != nil {
				p.fail(err.Error())
				return p.returnTo(p.GenURL(`export`))
			}
			defer f.Close()
			if shared.SupportedExport(`postgres`) {
				if hasStruct {
					structWriter = f
				}
				if hasData {
					dataWriter = f
				}
				err = shared.NativeExportPG(context.Background(), cfg, tables, structWriter, dataWriter)
			} else {
				err = shared.ExportMultipleTablesToWriter(f, p.db, tables, `"`, hasStruct, hasData, p.getVersion())
			}
			if err != nil {
				p.fail(err.Error())
			} else {
				p.ok(p.T(`导出成功，文件: %s`, file))
			}
			return p.returnTo(p.GenURL(`export`))
		}
	}
	p.Set(`tableList`, p.Get(`tableList`))
	p.Set(`dbName`, p.dbName)
	return p.Render(`db/postgres/export`, p.checkErr(err))
}

func (p *Postgres) Import() error {
	if p.IsPost() {
		if len(p.dbName) == 0 {
			p.fail(p.T(`请选择数据库`))
			return p.returnTo(p.GenURL(`listDb`))
		}
		sqlContent := p.Form(`sql_content`)
		var stats shared.SQLStats
		if len(sqlContent) > 0 {
			stats = shared.ExecuteSQL(p.db, sqlContent)
		} else {
			file, hdr, err := p.Request().FormFile(`file`)
			if err != nil {
				p.fail(p.T(`请提供SQL内容或上传文件`))
				return p.returnTo(p.GenURL(`import`, p.dbName))
			}
			defer file.Close()
			stats, err = shared.ExecuteUploadedSQLFile(p.db, file, hdr)
			if err != nil {
				p.fail(err.Error())
				return p.returnTo(p.GenURL(`import`, p.dbName))
			}
		}
		if stats.Failed > 0 {
			p.fail(p.T(`成功 %d 条，失败 %d 条: %s`, stats.Success, stats.Failed, strings.Join(stats.Errors, "; ")))
		} else {
			p.ok(p.T(`成功执行 %d 条SQL语句`, stats.Success))
		}
		return p.returnTo(p.GenURL(`import`, p.dbName))
	}
	return p.Render(`db/postgres/import`, p.checkErr(nil))
}

// Info handles the server info operation
func (p *Postgres) Info() error {
	infos := map[string]string{}

	// Version
	var version string
	p.db.QueryRow(`SELECT version()`).Scan(&version)
	infos[`version`] = version

	// Database size
	var dbSize string
	p.db.QueryRow(`SELECT pg_size_pretty(pg_database_size(current_database()))`).Scan(&dbSize)
	infos[`database_size`] = dbSize

	// Current connections
	var connections int
	p.db.QueryRow(`SELECT count(*) FROM pg_stat_activity`).Scan(&connections)
	infos[`connections`] = fmt.Sprintf(`%d`, connections)

	// Uptime
	var uptime string
	p.db.QueryRow(`SELECT pg_postmaster_start_time()::text`).Scan(&uptime)
	infos[`uptime`] = uptime

	p.Set(`infos`, infos)
	return p.Render(`db/postgres/info`, p.checkErr(nil))
}

// ProcessList handles showing the process list
func (p *Postgres) ProcessList() error {
	sqlStr := `
		SELECT 
			pid,
			datname,
			usename,
			application_name,
			client_addr::text,
			state,
			query,
			query_start::text,
			state_change::text
		FROM pg_stat_activity
		ORDER BY query_start DESC NULLS LAST`

	rows, err := p.db.Query(sqlStr)
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(p.GenURL(`listDb`))
	}
	defer rows.Close()

	type ProcessItem struct {
		Pid         string
		Database    string
		User        string
		AppName     string
		Client      string
		State       string
		Query       string
		QueryStart  string
		StateChange string
	}

	var processes []ProcessItem
	for rows.Next() {
		var (
			pid         string
			database    sql.NullString
			user        string
			appName     sql.NullString
			client      sql.NullString
			state       sql.NullString
			query       sql.NullString
			queryStart  sql.NullString
			stateChange sql.NullString
		)
		if err := rows.Scan(&pid, &database, &user, &appName, &client, &state, &query, &queryStart, &stateChange); err != nil {
			continue
		}
		processes = append(processes, ProcessItem{
			Pid:         pid,
			Database:    database.String,
			User:        user,
			AppName:     appName.String,
			Client:      client.String,
			State:       state.String,
			Query:       query.String,
			QueryStart:  queryStart.String,
			StateChange: stateChange.String,
		})
	}

	p.Set(`processes`, processes)
	return p.Render(`db/postgres/process_list`, p.checkErr(err))
}

// Privileges handles the privileges management operation
func (p *Postgres) Privileges() error {
	act := p.Form(`act`)

	// Handle create/edit/drop
	switch act {
	case `create`:
		if p.IsPost() {
			roleName := p.Form(`role`)
			password := p.Form(`password`)
			isSuper := p.Formx(`super`).Bool()
			canLogin := p.Formx(`login`).Bool()
			canCreateDB := p.Formx(`createdb`).Bool()
			canCreateRole := p.Formx(`createrole`).Bool()
			if len(roleName) == 0 {
				p.fail(p.T(`角色名不能为空`))
				return p.returnTo(p.GenURL(`privileges`))
			}
			loginStr := `NOLOGIN`
			if canLogin {
				loginStr = `LOGIN`
			}
			superStr := `NOSUPERUSER`
			if isSuper {
				superStr = `SUPERUSER`
			}
			createDBStr := `NOCREATEDB`
			if canCreateDB {
				createDBStr = `CREATEDB`
			}
			createRoleStr := `NOCREATEROLE`
			if canCreateRole {
				createRoleStr = `CREATEROLE`
			}
			passwordStr := ``
			if len(password) > 0 {
				passwordStr = ` PASSWORD ` + QuoteVal(password)
			}
			sqlStr := fmt.Sprintf(`CREATE ROLE %s WITH %s %s %s %s%s`, QuoteCol(roleName), loginStr, superStr, createDBStr, createRoleStr, passwordStr)
			_, err := p.db.Exec(sqlStr)
			if err != nil {
				p.fail(err.Error())
			} else {
				p.ok(p.T(`角色 %s 创建成功`, roleName))
			}
			return p.returnTo(p.GenURL(`privileges`))
		}
		return p.Render(`db/postgres/privilege_edit`, p.checkErr(nil))

	case `edit`:
		roleName := p.Form(`role`)
		if len(roleName) == 0 {
			p.fail(p.T(`角色名不能为空`))
			return p.returnTo(p.GenURL(`privileges`))
		}
		if p.IsPost() {
			password := p.Form(`password`)
			isSuper := p.Formx(`super`).Bool()
			canLogin := p.Formx(`login`).Bool()
			canCreateDB := p.Formx(`createdb`).Bool()
			canCreateRole := p.Formx(`createrole`).Bool()

			loginStr := `NOLOGIN`
			if canLogin {
				loginStr = `LOGIN`
			}
			superStr := `NOSUPERUSER`
			if isSuper {
				superStr = `SUPERUSER`
			}
			createDBStr := `NOCREATEDB`
			if canCreateDB {
				createDBStr = `CREATEDB`
			}
			createRoleStr := `NOCREATEROLE`
			if canCreateRole {
				createRoleStr = `CREATEROLE`
			}

			sqlStr := fmt.Sprintf(`ALTER ROLE %s WITH %s %s %s %s`, QuoteCol(roleName), loginStr, superStr, createDBStr, createRoleStr)
			if len(password) > 0 {
				sqlStr += ` PASSWORD ` + QuoteVal(password)
			}
			_, err := p.db.Exec(sqlStr)
			if err != nil {
				p.fail(err.Error())
			} else {
				p.ok(p.T(`角色 %s 修改成功`, roleName))
			}
			return p.returnTo(p.GenURL(`privileges`))
		}
		// Load existing role info for the edit form
		var super, inherit, createRole, createDB, canLogin, replication bool
		row := p.db.QueryRow(`SELECT rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication FROM pg_roles WHERE rolname = $1`, roleName)
		row.Scan(&super, &inherit, &createRole, &createDB, &canLogin, &replication)
		p.Set(`role`, map[string]interface{}{
			`Name`:       roleName,
			`Super`:      super,
			`CreateRole`: createRole,
			`CreateDB`:   createDB,
			`CanLogin`:   canLogin,
		})
		return p.Render(`db/postgres/privilege_edit`, p.checkErr(nil))

	case `drop`:
		roleName := p.Form(`role`)
		if len(roleName) == 0 {
			p.fail(p.T(`角色名不能为空`))
			return p.returnTo(p.GenURL(`privileges`))
		}
		_, err := p.db.Exec(fmt.Sprintf(`DROP ROLE IF EXISTS %s`, QuoteCol(roleName)))
		if err != nil {
			p.fail(err.Error())
		} else {
			p.ok(p.T(`角色 %s 已删除`, roleName))
		}
		return p.returnTo(p.GenURL(`privileges`))
	}

	// List roles
	sqlStr := `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication FROM pg_roles ORDER BY rolname`
	rows, err := p.db.Query(sqlStr)
	if err != nil {
		p.fail(err.Error())
		return p.returnTo(p.GenURL(`listDb`))
	}
	defer rows.Close()

	type RoleInfo struct {
		Name        string
		Super       bool
		Inherit     bool
		CreateRole  bool
		CreateDB    bool
		CanLogin    bool
		Replication bool
	}

	var roles []RoleInfo
	for rows.Next() {
		var r RoleInfo
		if err := rows.Scan(&r.Name, &r.Super, &r.Inherit, &r.CreateRole, &r.CreateDB, &r.CanLogin, &r.Replication); err != nil {
			continue
		}
		roles = append(roles, r)
	}

	p.Set(`roles`, roles)
	return p.Render(`db/postgres/privileges`, p.checkErr(err))
}
