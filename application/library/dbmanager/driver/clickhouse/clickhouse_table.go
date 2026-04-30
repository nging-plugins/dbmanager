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
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/coscms/webcore/library/backend"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/shared"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
	"github.com/webx-top/pagination"
)

// getVersion returns the ClickHouse server version
func (c *ClickHouse) getVersion() string {
	if len(c.version) > 0 {
		return c.version
	}
	var v string
	c.db.QueryRow(`SELECT version()`).Scan(&v)
	c.version = v
	return v
}

// getTables returns a list of tables in the current database
func (c *ClickHouse) getTables() ([]string, error) {
	sqlStr := `SHOW TABLES`
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

// TableStatus holds table status information for ClickHouse
type TableStatus struct {
	Name    string
	Engine  string
	Rows    int64
	Size    int64
	Comment string
}

// getTableStatus returns detailed table status information
func (c *ClickHouse) getTableStatus(tableName string) (map[string]*TableStatus, error) {
	sqlStr := `
		SELECT
			name,
			engine,
			total_rows,
			total_bytes,
			comment
		FROM system.tables
		WHERE database = currentDatabase()`

	if len(tableName) > 0 {
		sqlStr += ` AND name = '` + strings.ReplaceAll(tableName, `'`, `''`) + `'`
	}
	sqlStr += ` ORDER BY name`

	rows, err := c.db.Query(sqlStr)
	if err != nil {
		return nil, fmt.Errorf(`%v: %v`, sqlStr, err)
	}
	defer rows.Close()

	ret := map[string]*TableStatus{}
	for rows.Next() {
		var (
			name    string
			engine  sql.NullString
			rowsN   sql.NullInt64
			sizeN   sql.NullInt64
			comment sql.NullString
		)
		if err := rows.Scan(&name, &engine, &rowsN, &sizeN, &comment); err != nil {
			return nil, err
		}
		v := &TableStatus{
			Name:    name,
			Engine:  engine.String,
			Rows:    rowsN.Int64,
			Size:    sizeN.Int64,
			Comment: comment.String,
		}
		ret[v.Name] = v
	}
	return ret, rows.Err()
}

// FieldInfo holds column information for ClickHouse
type FieldInfo struct {
	ColumnName  string
	DataType    string
	DefaultType string
	DefaultExpr string
	Comment     string
}

// getTableFields returns column information for a table
func (c *ClickHouse) getTableFields(tableName string) ([]*FieldInfo, error) {
	sqlStr := `
		SELECT
			name,
			type,
			default_kind,
			default_expression,
			comment
		FROM system.columns
		WHERE database = currentDatabase() AND table = $1
		ORDER BY position`

	rows, err := c.db.Query(sqlStr, tableName)
	if err != nil {
		// Try with ? placeholder for older versions
		rows, err = c.db.Query(
			`SELECT name, type, default_kind, default_expression, comment
			FROM system.columns
			WHERE database = currentDatabase() AND table = ?
			ORDER BY position`,
			tableName,
		)
		if err != nil {
			return nil, fmt.Errorf(`getTableFields: %v`, err)
		}
	}
	defer rows.Close()

	var ret []*FieldInfo
	for rows.Next() {
		var (
			colName    string
			dataType   string
			defaultTyp sql.NullString
			defaultExp sql.NullString
			comment    sql.NullString
		)
		if err := rows.Scan(&colName, &dataType, &defaultTyp, &defaultExp, &comment); err != nil {
			return nil, err
		}
		v := &FieldInfo{
			ColumnName:  colName,
			DataType:    dataType,
			DefaultType: defaultTyp.String,
			DefaultExpr: defaultExp.String,
			Comment:     comment.String,
		}
		ret = append(ret, v)
	}
	return ret, rows.Err()
}

// getTableDDL returns the CREATE TABLE DDL statement for ClickHouse
func (c *ClickHouse) getTableDDL(tableName string) (string, error) {
	// ClickHouse doesn't have SHOW CREATE TABLE, so we'll construct it
	fields, err := c.getTableFields(tableName)
	if err != nil {
		return "", err
	}

	var colDefs []string
	for _, f := range fields {
		def := QuoteCol(f.ColumnName) + ` ` + f.DataType
		if len(f.DefaultExpr) > 0 {
			def += ` DEFAULT ` + f.DefaultExpr
		}
		if len(f.Comment) > 0 {
			def += ` COMMENT ` + QuoteVal(f.Comment)
		}
		colDefs = append(colDefs, def)
	}

	engine := `MergeTree()`
	// Try to get the actual engine
	var tblEngine sql.NullString
	c.db.QueryRow(
		`SELECT engine FROM system.tables WHERE database = currentDatabase() AND name = ?`,
		tableName,
	).Scan(&tblEngine)
	if tblEngine.Valid && len(tblEngine.String) > 0 {
		engine = tblEngine.String
	}

	ddl := fmt.Sprintf("CREATE TABLE %s (\n  %s\n) ENGINE = %s",
		QuoteCol(tableName),
		strings.Join(colDefs, ",\n  "),
		engine,
	)

	// Get ORDER BY / PRIMARY KEY info
	var sortKey sql.NullString
	c.db.QueryRow(
		`SELECT sorting_key FROM system.tables WHERE database = currentDatabase() AND name = ?`,
		tableName,
	).Scan(&sortKey)
	if sortKey.Valid && len(sortKey.String) > 0 {
		ddl += " ORDER BY " + sortKey.String
	}

	return ddl + ";", nil
}

// ListDb handles the list databases operation
func (c *ClickHouse) ListDb() error {
	dbList, err := c.getDatabases()
	if err != nil {
		c.fail(err.Error())
		return c.returnTo(backend.URLFor(`/db`))
	}
	c.Set(`dbList`, dbList)
	return c.Render(`db/clickhouse/list_db`, c.checkErr(err))
}

// CreateDb handles the create database operation
func (c *ClickHouse) CreateDb() error {
	if c.IsPost() {
		data := c.Data()
		dbName := c.Form(`db`)
		if len(dbName) == 0 {
			data.SetInfo(c.T(`数据库名不能为空`), 0)
			return c.JSON(data)
		}

		err := c.createDatabase(dbName)
		if err != nil {
			data.SetError(err)
		} else {
			data.SetInfo(c.T(`数据库创建成功`), 1)
		}
		return c.JSON(data)
	}
	return c.Render(`db/clickhouse/create_db`, c.checkErr(nil))
}

// ModifyDb handles the modify/delete database operation
func (c *ClickHouse) ModifyDb() error {
	if c.IsPost() {
		data := c.Data()
		operate := c.Form(`operate`)

		switch operate {
		case `drop`:
			dbName := c.Form(`db`)
			if len(dbName) == 0 {
				data.SetInfo(c.T(`数据库名不能为空`), 0)
				return c.JSON(data)
			}
			err := c.dropDatabase(dbName)
			if err != nil {
				data.SetError(err)
			} else {
				data.SetInfo(c.T(`数据库删除成功`), 1)
			}
		default:
			data.SetInfo(c.T(`不支持的操作`), 0)
		}
		return c.JSON(data)
	}
	return c.Render(`db/clickhouse/modify_db`, c.checkErr(nil))
}

// ListTable handles the list tables operation
func (c *ClickHouse) ListTable() error {
	tables, err := c.getTables()
	if err != nil {
		c.fail(err.Error())
		return c.returnTo(c.GenURL(`listDb`))
	}
	c.Set(`tableList`, tables)

	// Also get table statuses
	statuses, _ := c.getTableStatus(``)
	c.Set(`tableStatuses`, statuses)

	return c.Render(`db/clickhouse/list_table`, c.checkErr(err))
}

// ViewTable handles the view table structure operation
func (c *ClickHouse) ViewTable() error {
	tableName := c.Form(`table`)
	if len(tableName) == 0 {
		c.fail(c.T(`请选择表`))
		return c.returnTo(c.GenURL(`listTable`))
	}

	fields, err := c.getTableFields(tableName)
	if err != nil {
		c.fail(err.Error())
		return c.returnTo(c.GenURL(`listTable`))
	}
	c.Set(`fields`, fields)

	ddl, err := c.getTableDDL(tableName)
	if err == nil {
		c.Set(`ddl`, ddl)
	}

	// Get indexes (sorting key / primary key)
	var sortKey, primaryKey sql.NullString
	c.db.QueryRow(
		`SELECT sorting_key, primary_key FROM system.tables WHERE database = currentDatabase() AND name = ?`,
		tableName,
	).Scan(&sortKey, &primaryKey)
	var idx interface{}
	if sortKey.Valid || primaryKey.Valid {
		idx = &struct {
			SortingKey string
			PrimaryKey string
		}{
			SortingKey: sortKey.String,
			PrimaryKey: primaryKey.String,
		}
	}
	c.Set(`indexes`, idx)

	return c.Render(`db/clickhouse/view_table`, c.checkErr(err))
}

// RunCommand handles running SQL commands
func (c *ClickHouse) RunCommand() error {
	err := shared.RunSQL(c.BaseDriver, c.db, nil)
	c.SetFunc(`getFieldsByTable`, c.getFieldsByTable)
	return c.Render(`db/clickhouse/sql`, c.checkErr(err))
}

func (c *ClickHouse) getFieldsByTable(table string) []string {
	columns, _ := shared.GetTableColumns(c.db, table)
	if columns == nil {
		return []string{}
	}
	return columns
}

// ListData handles displaying table data (paginated)
func (c *ClickHouse) ListData() error {
	tableName := c.Form(`table`)
	if len(tableName) == 0 {
		c.fail(c.T(`请选择表`))
		return c.returnTo(c.GenURL(`listTable`))
	}

	limit := c.Formx(`limit`).Int()
	page := c.Formx(`page`).Int()
	totalRows := c.Formx(`rows`).Int()

	if limit < 1 {
		limit = 50
		c.Request().Form().Set(`limit`, strconv.Itoa(limit))
	}
	if page < 1 {
		page = 1
	}

	// Get fields for column display
	fields, err := c.getTableFields(tableName)
	if err != nil {
		c.fail(err.Error())
		return c.returnTo(c.GenURL(`listTable`))
	}
	c.Set(`fields`, fields)

	// Get total row count if not provided
	if totalRows < 1 {
		var count sql.NullInt64
		countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s`, QuoteCol(tableName))
		c.db.QueryRow(countSQL).Scan(&count)
		totalRows = int(count.Int64)
	}

	offset := (page - 1) * limit

	// Query data
	sqlStr := fmt.Sprintf(`SELECT * FROM %s LIMIT %d OFFSET %d`, QuoteCol(tableName), limit, offset)
	rows, err := c.db.Query(sqlStr)
	if err != nil {
		c.fail(err.Error())
		return c.returnTo(c.GenURL(`listTable`))
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		c.fail(err.Error())
		return c.returnTo(c.GenURL(`listTable`))
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

	c.Set(`columns`, columns)
	c.Set(`values`, results)
	c.Set(`total`, totalRows)
	c.Set(`table`, tableName)

	// Pagination - same pattern as MySQL
	q := c.Request().URL().Query()
	q.Del(`page`)
	q.Del(`rows`)
	q.Del(`_pjax`)
	c.Set(`pagination`, pagination.New(c.Context).SetURL(backend.URLFor(`/db`)+`?`+q.Encode()+`&page={page}&rows={rows}`).SetPage(page).SetRows(totalRows))

	return c.Render(`db/clickhouse/list_data`, c.checkErr(err))
}

// CreateTable handles the create table operation
func (c *ClickHouse) CreateTable() error {
	if c.IsPost() {
		tableName := c.Form(`table`)
		if len(tableName) == 0 {
			c.fail(c.T(`表名不能为空`))
			return c.returnTo(c.GenURL(`createTable`, c.dbName))
		}

		columnNames := c.FormValues(`column_name[]`)
		columnTypes := c.FormValues(`column_type[]`)
		columnComments := c.FormValues(`column_comment[]`)

		if len(columnNames) == 0 {
			c.fail(c.T(`请至少定义一个字段`))
			return c.returnTo(c.GenURL(`createTable`, c.dbName))
		}

		engine := c.Form(`engine`, `MergeTree()`)
		orderBy := c.Form(`order_by`)

		var colDefs []string
		for i, colName := range columnNames {
			if len(colName) == 0 || i >= len(columnTypes) {
				continue
			}
			def := QuoteCol(colName) + ` ` + columnTypes[i]
			if i < len(columnComments) && len(columnComments[i]) > 0 {
				def += ` COMMENT ` + QuoteVal(columnComments[i])
			}
			colDefs = append(colDefs, def)
		}

		sqlStr := fmt.Sprintf(
			`CREATE TABLE %s (%s) ENGINE = %s`,
			QuoteCol(tableName),
			strings.Join(colDefs, `, `),
			engine,
		)
		if len(orderBy) > 0 {
			sqlStr += ` ORDER BY ` + orderBy
		}

		_, err := c.db.Exec(sqlStr)
		if err != nil {
			c.fail(err.Error())
		} else {
			c.ok(c.T(`表创建成功`))
		}
		return c.returnTo(c.GenURL(`listTable`, c.dbName))
	}

	// Provide engine suggestions
	c.Set(`engines`, []string{
		`MergeTree()`,
		`ReplacingMergeTree()`,
		`SummingMergeTree()`,
		`AggregatingMergeTree()`,
		`CollapsingMergeTree(sign)`,
		`VersionedCollapsingMergeTree(sign, version)`,
		`Log`,
		`TinyLog`,
		`StripeLog`,
		`Memory`,
		`Distributed(cluster, db, table, rand())`,
	})
	return c.Render(`db/clickhouse/create_table`, c.checkErr(nil))
}

// ModifyTable handles the modify table operation
func (c *ClickHouse) ModifyTable() error {
	tableName := c.Form(`table`)
	if len(tableName) == 0 {
		c.fail(c.T(`请选择表`))
		return c.returnTo(c.GenURL(`listTable`))
	}

	if c.IsPost() || len(c.Form("operate")) > 0 {
		operate := c.Form(`operate`)

		switch operate {
		case `drop`:
			sqlStr := fmt.Sprintf(`DROP TABLE %s`, QuoteCol(tableName))
			_, err := c.db.Exec(sqlStr)
			if err != nil {
				c.fail(err.Error())
			} else {
				c.ok(c.T(`表删除成功`))
			}
			return c.returnTo(c.GenURL(`listTable`, c.dbName))
		case `truncate`:
			sqlStr := fmt.Sprintf(`TRUNCATE TABLE %s`, QuoteCol(tableName))
			_, err := c.db.Exec(sqlStr)
			if err != nil {
				c.fail(err.Error())
			} else {
				c.ok(c.T(`表已清空`))
			}
			return c.returnTo(c.GenURL(`listTable`, c.dbName))
		case `rename`:
			newName := c.Form(`newName`)
			if len(newName) == 0 {
				c.fail(c.T(`新表名不能为空`))
				return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
			}
			sqlStr := fmt.Sprintf(`RENAME TABLE %s TO %s`, QuoteCol(tableName), QuoteCol(newName))
			_, err := c.db.Exec(sqlStr)
			if err != nil {
				c.fail(err.Error())
			} else {
				c.ok(c.T(`表重命名成功`))
			}
			return c.returnTo(c.GenURL(`listTable`, c.dbName))
		case `modify`:
			newName := c.Form(`new_name`)
			comment := c.Form(`comment`)
			if len(newName) > 0 {
				sqlStr := fmt.Sprintf(`RENAME TABLE %s TO %s`, QuoteCol(tableName), QuoteCol(newName))
				_, err := c.db.Exec(sqlStr)
				if err != nil {
					c.fail(err.Error())
					return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
				}
				tableName = newName
			}
			if len(comment) > 0 {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s MODIFY COMMENT %s`, QuoteCol(tableName), QuoteVal(comment))
				_, err := c.db.Exec(sqlStr)
				if err != nil {
					c.fail(err.Error())
					return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
				}
			}
			c.ok(c.T(`修改成功`))
			return c.returnTo(c.GenURL(`listTable`, c.dbName))
		case `modifyColumn`:
			colName := c.Form(`colName`)
			newColName := c.Form(`newColName`)
			colType := c.Form(`colType`)
			colComment := c.Form(`colComment`)
			if len(colName) == 0 {
				c.fail(c.T(`列名不能为空`))
				return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
			}
			targetCol := newColName
			if len(targetCol) == 0 {
				targetCol = colName
			}
			if newColName != colName && len(newColName) > 0 {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s RENAME COLUMN %s TO %s`, QuoteCol(tableName), QuoteCol(colName), QuoteCol(newColName))
				_, err := c.db.Exec(sqlStr)
				if err != nil {
					c.fail(err.Error())
					return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
				}
			}
			if len(colType) > 0 {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s MODIFY COLUMN %s %s`, QuoteCol(tableName), QuoteCol(targetCol), colType)
				_, err := c.db.Exec(sqlStr)
				if err != nil {
					c.fail(err.Error())
					return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
				}
			}
			if len(colComment) > 0 {
				sqlStr := fmt.Sprintf(`ALTER TABLE %s COMMENT COLUMN %s %s`, QuoteCol(tableName), QuoteCol(targetCol), QuoteVal(colComment))
				_, err := c.db.Exec(sqlStr)
				if err != nil {
					c.fail(err.Error())
					return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
				}
			}
			c.ok(c.T(`修改成功`))
			return c.returnTo(c.GenURL(`viewTable`, c.dbName, tableName))
		case `dropColumn`:
			colName := c.Form(`colName`)
			if len(colName) == 0 {
				c.fail(c.T(`列名不能为空`))
				return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
			}
			sqlStr := fmt.Sprintf(`ALTER TABLE %s DROP COLUMN %s`, QuoteCol(tableName), QuoteCol(colName))
			_, err := c.db.Exec(sqlStr)
			if err != nil {
				c.fail(err.Error())
			} else {
				c.ok(c.T(`列删除成功`))
			}
			return c.returnTo(c.GenURL(`viewTable`, c.dbName, tableName))
		default:
			c.fail(c.T(`不支持的操作`))
			return c.returnTo(c.GenURL(`modifyTable`, c.dbName, tableName))
		}
	}

	fields, err := c.getTableFields(tableName)
	if err == nil {
		c.Set(`fields`, fields)
	}
	return c.Render(`db/clickhouse/modify_table`, c.checkErr(err))
}

// Indexes handles showing indexes for a table (ClickHouse uses ORDER BY / PRIMARY KEY)
func (c *ClickHouse) Indexes() error {
	tableName := c.Form(`table`)
	if len(tableName) == 0 {
		c.fail(c.T(`请选择表`))
		return c.returnTo(c.GenURL(`listTable`))
	}

	// Handle POST to modify ORDER BY / PRIMARY KEY / SAMPLE BY
	if c.IsPost() {
		sortingKey := c.Form(`sorting_key`)
		primaryKey := c.Form(`primary_key`)
		sampleBy := c.Form(`sample_by`)

		var sqls []string
		if len(sortingKey) > 0 {
			sqls = append(sqls, fmt.Sprintf(`ALTER TABLE %s MODIFY ORDER BY %s`, QuoteCol(tableName), sortingKey))
		}
		if len(primaryKey) > 0 {
			sqls = append(sqls, fmt.Sprintf(`ALTER TABLE %s MODIFY PRIMARY KEY %s`, QuoteCol(tableName), primaryKey))
		}
		if len(sampleBy) > 0 {
			sqls = append(sqls, fmt.Sprintf(`ALTER TABLE %s MODIFY SAMPLE BY %s`, QuoteCol(tableName), sampleBy))
		}
		if len(sqls) == 0 {
			c.fail(c.T(`请至少填写一项`))
			return c.returnTo(c.GenURL(`indexes`, c.dbName, tableName))
		}
		var lastErr error
		for _, s := range sqls {
			_, lastErr = c.db.Exec(s)
			if lastErr != nil {
				break
			}
		}
		if lastErr != nil {
			c.fail(lastErr.Error())
		} else {
			c.ok(c.T(`索引修改成功`))
		}
		return c.returnTo(c.GenURL(`indexes`, c.dbName, tableName))
	}

	// Get sorting key and primary key from system.tables
	var sortKey, primaryKey, sampleBy sql.NullString
	c.db.QueryRow(
		`SELECT sorting_key, primary_key, sampling_key FROM system.tables WHERE database = currentDatabase() AND name = ?`,
		tableName,
	).Scan(&sortKey, &primaryKey, &sampleBy)

	var indexes interface{}
	if sortKey.Valid || primaryKey.Valid || sampleBy.Valid {
		indexes = &struct {
			Name       string
			SortingKey string
			PrimaryKey string
			SampleBy   string
		}{
			Name:       tableName,
			SortingKey: sortKey.String,
			PrimaryKey: primaryKey.String,
			SampleBy:   sampleBy.String,
		}
	}

	c.Set(`indexes`, indexes)
	c.Set(`table`, tableName)
	return c.Render(`db/clickhouse/indexes`, c.checkErr(nil))
}

// Export handles data export
func (c *ClickHouse) Export() error {
	process := c.Queryx(`process`).Bool()
	if process {
		return c.bgExecManage(shared.OpExport)
	}
	var err error
	if c.IsPost() {
		if len(c.dbName) == 0 {
			c.fail(c.T(`请选择数据库`))
			return c.returnTo(c.GenURL(`listDb`))
		}
		var tables []string
		if c.Formx(`all`).Bool() {
			tables, _ = c.getTables()
		} else {
			tables = c.FormValues(`table`)
		}
		if len(tables) == 0 {
			c.fail(c.T(`请选择要导出的表`))
			return c.returnTo(c.GenURL(`listTable`, c.dbName))
		}
		host, port := shared.SplitHostPort(c.DbAuth.Host)
		if len(port) == 0 {
			port = `9000`
		}
		cfg := &shared.DBConfig{
			Driver:   `clickhouse`,
			Host:     host,
			Port:     port,
			Username: c.DbAuth.Username,
			Password: c.DbAuth.Password,
			Database: c.dbName,
		}
		output := c.Form(`output`, `down`)
		types := c.FormValues(`type`)

		switch output {
		case `down`, `open`:
			if output == `down` {
				c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
				c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", c.dbName+".sql"))
			} else {
				c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
			}
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
			if SupportedCmdExport() {
				var structWriter, dataWriter io.Writer
				if hasStruct {
					structWriter = c.Response()
				}
				if hasData {
					dataWriter = c.Response()
				}
				return NativeExportCH(c.Request().Context(), cfg, tables, structWriter, dataWriter)
			}
			return shared.ExportMultipleTablesToWriter(c.Response(), c.db, tables, "`", hasStruct, hasData, c.getVersion())
		default: // send
			saveDir := filepath.Join(os.TempDir(), `dbmanager/cache/export`, c.dbName)
			com.MkdirAll(saveDir, os.ModePerm)
			nowTime := time.Now().Format("20060102150405")
			file := filepath.Join(saveDir, nowTime+`.sql`)
			f, err := os.Create(file)
			if err != nil {
				c.fail(err.Error())
				return c.returnTo(c.GenURL(`export`))
			}
			defer f.Close()
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
			if SupportedCmdExport() {
				var structWriter, dataWriter io.Writer
				if hasStruct {
					structWriter = f
				}
				if hasData {
					dataWriter = f
				}
				err = NativeExportCH(c.Request().Context(), cfg, tables, structWriter, dataWriter)
			} else {
				err = shared.ExportMultipleTablesToWriter(f, c.db, tables, "`", hasStruct, hasData, c.getVersion())
			}
			if err != nil {
				c.fail(err.Error())
			} else {
				c.ok(c.T(`导出成功，文件: %s`, file))
			}
			return c.returnTo(c.GenURL(`export`))
		}
	}
	c.Set(`tableList`, c.Get(`tableList`))
	c.Set(`dbName`, c.dbName)
	return c.Render(`db/clickhouse/export`, c.checkErr(err))
}

func (c *ClickHouse) Import() error {
	process := c.Queryx(`process`).Bool()
	if process {
		return c.bgExecManage(shared.OpImport)
	}
	if c.IsPost() {
		if len(c.dbName) == 0 {
			c.fail(c.T(`请选择数据库`))
			return c.returnTo(c.GenURL(`listDb`))
		}
		sqlFiles, err := shared.GetSQLFiles(c.Context)
		if err != nil {
			return shared.ResponseDropzone(err, c.Context)
		}
		async := c.Formx(`async`, `true`).Bool()
		executor := shared.NewSQLExecutor(c.db)
		err = shared.ImportSQLFiles(c.Context, *c.DbAuth, c.dbName, ``, sqlFiles, async, executor.ImportExecutor)
		return shared.ResponseDropzone(err, c.Context)
	}
	return c.Render(`db/clickhouse/import`, c.checkErr(nil))
}

// Info handles the server info operation
func (c *ClickHouse) Info() error {
	infos := map[string]string{}

	var version string
	c.db.QueryRow(`SELECT version()`).Scan(&version)
	infos[`version`] = version

	var uptime string
	c.db.QueryRow(`SELECT formatReadableTimeDelta(uptime())`).Scan(&uptime)
	infos[`uptime`] = uptime

	var totalDatabases int
	c.db.QueryRow(`SELECT count() FROM system.databases`).Scan(&totalDatabases)
	infos[`total_databases`] = fmt.Sprintf(`%d`, totalDatabases)

	var totalTables int
	c.db.QueryRow(`SELECT count() FROM system.tables WHERE database = currentDatabase()`).Scan(&totalTables)
	infos[`total_tables`] = fmt.Sprintf(`%d`, totalTables)

	c.Set(`infos`, infos)
	return c.Render(`db/clickhouse/info`, c.checkErr(nil))
}

// ProcessList handles showing the process list
func (c *ClickHouse) ProcessList() error {
	sqlStr := `
		SELECT
			query_id,
			user,
			query,
			elapsed,
			read_rows,
			memory_usage
		FROM system.processes
		ORDER BY elapsed DESC`

	rows, err := c.db.Query(sqlStr)
	if err != nil {
		c.fail(err.Error())
		return c.returnTo(c.GenURL(`listDb`))
	}
	defer rows.Close()

	type ProcessItem struct {
		QueryID   string
		User      string
		Query     string
		Elapsed   float64
		ReadRows  int64
		ReadBytes int64
	}

	var processes []ProcessItem
	for rows.Next() {
		var (
			queryID  string
			user     sql.NullString
			query    sql.NullString
			elapsed  sql.NullFloat64
			readRows sql.NullInt64
			memory   sql.NullInt64
		)
		if err := rows.Scan(&queryID, &user, &query, &elapsed, &readRows, &memory); err != nil {
			continue
		}
		item := ProcessItem{
			QueryID:   queryID,
			User:      user.String,
			Query:     query.String,
			Elapsed:   elapsed.Float64,
			ReadRows:  readRows.Int64,
			ReadBytes: memory.Int64,
		}
		processes = append(processes, item)
	}

	c.Set(`processes`, processes)
	return c.Render(`db/clickhouse/process_list`, c.checkErr(err))
}

// Privileges handles the privileges management operation
func (c *ClickHouse) Privileges() error {
	act := c.Form(`act`)

	switch act {
	case `create`:
		if c.IsPost() {
			userName := c.Form(`user`)
			password := c.Form(`password`)
			if len(userName) == 0 {
				c.fail(c.T(`用户名不能为空`))
				return c.returnTo(c.GenURL(`privileges`))
			}
			sqlStr := fmt.Sprintf(`CREATE USER %s`, QuoteCol(userName))
			if len(password) > 0 {
				sqlStr += ` IDENTIFIED BY ` + QuoteVal(password)
			}
			_, err := c.db.Exec(sqlStr)
			if err != nil {
				c.fail(err.Error())
			} else {
				c.ok(c.T(`用户 %s 创建成功`, userName))
			}
			return c.returnTo(c.GenURL(`privileges`))
		}
		return c.Render(`db/clickhouse/privilege_edit`, c.checkErr(nil))

	case `edit`:
		userName := c.Form(`user`)
		if len(userName) == 0 {
			c.fail(c.T(`用户名不能为空`))
			return c.returnTo(c.GenURL(`privileges`))
		}
		if c.IsPost() {
			password := c.Form(`password`)
			if len(password) > 0 {
				_, err := c.db.Exec(fmt.Sprintf(`ALTER USER %s IDENTIFIED BY %s`, QuoteCol(userName), QuoteVal(password)))
				if err != nil {
					c.fail(err.Error())
				} else {
					c.ok(c.T(`用户 %s 密码修改成功`, userName))
				}
			}
			return c.returnTo(c.GenURL(`privileges`))
		}
		c.Set(`editUser`, map[string]string{`Name`: userName})
		return c.Render(`db/clickhouse/privilege_edit`, c.checkErr(nil))

	case `drop`:
		userName := c.Form(`user`)
		if len(userName) == 0 {
			c.fail(c.T(`用户名不能为空`))
			return c.returnTo(c.GenURL(`privileges`))
		}
		_, err := c.db.Exec(fmt.Sprintf(`DROP USER IF EXISTS %s`, QuoteCol(userName)))
		if err != nil {
			c.fail(err.Error())
		} else {
			c.ok(c.T(`用户 %s 已删除`, userName))
		}
		return c.returnTo(c.GenURL(`privileges`))
	}

	// List users
	sqlStr := `SELECT name, storage FROM system.users ORDER BY name`
	rows, err := c.db.Query(sqlStr)
	if err != nil {
		c.fail(err.Error())
		return c.returnTo(c.GenURL(`listDb`))
	}
	defer rows.Close()

	type UserInfo struct {
		Name    string
		Storage string
	}

	var users []UserInfo
	for rows.Next() {
		var (
			name    string
			storage sql.NullString
		)
		if err := rows.Scan(&name, &storage); err != nil {
			continue
		}
		users = append(users, UserInfo{Name: name, Storage: storage.String})
	}
	c.Set(`users`, users)
	return c.Render(`db/clickhouse/privileges`, c.checkErr(err))
}
