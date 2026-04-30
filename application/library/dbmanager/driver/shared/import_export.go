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
	"database/sql"
	"fmt"
	"io"
	"strings"

	"github.com/webx-top/echo"
)

// ExportTableToSQL streams all rows from a table as INSERT statements.
func ExportTableToSQL(ctx echo.Context, db *sql.DB, tableName, quoteChar string) error {
	cols, err := GetTableColumns(db, tableName)
	if err != nil {
		return err
	}
	qTable := QuoteIdentifier(tableName, quoteChar)
	colsQuoted := make([]string, len(cols))
	for i, c := range cols {
		colsQuoted[i] = QuoteIdentifier(c, quoteChar)
	}
	colList := strings.Join(colsQuoted, `, `)

	rows, err := db.Query(fmt.Sprintf(`SELECT * FROM %s`, qTable))
	if err != nil {
		return err
	}
	defer rows.Close()

	ctx.Response().Header().Set(echo.HeaderContentType, `text/plain; charset=utf-8`)
	ctx.Response().Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.sql"`, tableName))

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		line := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s);`,
			qTable, colList, FormatValues(values))
		ctx.Response().Write([]byte(line + "\n"))
	}
	return rows.Err()
}

// ExportTableToCSV streams all rows from a table as CSV.
func ExportTableToCSV(ctx echo.Context, db *sql.DB, tableName, quoteChar string) error {
	cols, err := GetTableColumns(db, tableName)
	if err != nil {
		return err
	}
	qTable := QuoteIdentifier(tableName, quoteChar)

	rows, err := db.Query(fmt.Sprintf(`SELECT * FROM %s`, qTable))
	if err != nil {
		return err
	}
	defer rows.Close()

	ctx.Response().Header().Set(echo.HeaderContentType, `text/csv; charset=utf-8`)
	ctx.Response().Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.csv"`, tableName))
	ctx.Response().Write([]byte(strings.Join(cols, `,`) + "\n"))

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}
		ctx.Response().Write([]byte(FormatCSVRow(values) + "\n"))
	}
	return rows.Err()
}

// ExportMultipleTablesToWriter exports multiple tables to a writer.
func ExportMultipleTablesToWriter(w io.Writer, db *sql.DB, tables []string, quoteChar string, withStruct, withData bool, version string) error {
	for _, table := range tables {
		if withStruct {
			ddl, _ := GetTableDDL(db, table, quoteChar)
			if ddl != "" {
				fmt.Fprintf(w, "\n-- Table: %s\n-- %s\n\n", table, ddl)
			}
		}
		if withData {
			cols, err := GetTableColumns(db, table)
			if err != nil {
				continue
			}
			qTable := QuoteIdentifier(table, quoteChar)
			colsQuoted := make([]string, len(cols))
			for i, c := range cols {
				colsQuoted[i] = QuoteIdentifier(c, quoteChar)
			}

			rows, err := db.Query(fmt.Sprintf(`SELECT * FROM %s`, qTable))
			if err != nil {
				continue
			}
			for rows.Next() {
				values := make([]interface{}, len(cols))
				valuePtrs := make([]interface{}, len(cols))
				for i := range cols {
					valuePtrs[i] = &values[i]
				}
				if err := rows.Scan(valuePtrs...); err != nil {
					continue
				}
				line := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s);\n`,
					qTable, strings.Join(colsQuoted, `, `), FormatValues(values))
				fmt.Fprint(w, line)
			}
			rows.Close()
		}
	}
	return nil
}

// FormatValues formats a row of values for SQL INSERT.
func FormatValues(values []interface{}) string {
	parts := make([]string, len(values))
	for i, v := range values {
		switch val := v.(type) {
		case []byte:
			parts[i] = QuoteValue(string(val))
		case nil:
			parts[i] = `NULL`
		default:
			parts[i] = QuoteValue(fmt.Sprintf(`%v`, val))
		}
	}
	return strings.Join(parts, `, `)
}

// FormatCSVRow formats a row values as CSV.
func FormatCSVRow(values []interface{}) string {
	parts := make([]string, len(values))
	for i, v := range values {
		switch val := v.(type) {
		case []byte:
			parts[i] = `"` + strings.ReplaceAll(string(val), `"`, `""`) + `"`
		case nil:
			parts[i] = ``
		default:
			parts[i] = `"` + fmt.Sprintf(`%v`, val) + `"`
		}
	}
	return strings.Join(parts, `,`)
}

// GetTableColumns returns column names for a table.
func GetTableColumns(db *sql.DB, tableName string) ([]string, error) {
	rows, err := db.Query(fmt.Sprintf(`SELECT * FROM %s WHERE 1=0`, QuoteIdentifier(tableName, `"`)))
	if err != nil {
		rows, err = db.Query(fmt.Sprintf("SELECT * FROM `%s` WHERE 1=0", tableName))
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()
	return rows.Columns()
}

// GetTableDDL attempts to retrieve table DDL.
func GetTableDDL(db *sql.DB, tableName, quoteChar string) (string, error) {
	rows, err := db.Query(fmt.Sprintf(`SELECT * FROM %s WHERE 1=0`, QuoteIdentifier(tableName, quoteChar)))
	if err != nil {
		return "", err
	}
	cols, _ := rows.Columns()
	rows.Close()

	var defs []string
	for _, c := range cols {
		defs = append(defs, QuoteIdentifier(c, quoteChar)+` ...`)
	}
	return fmt.Sprintf(`CREATE TABLE %s (\n  %s\n);`, QuoteIdentifier(tableName, quoteChar), strings.Join(defs, ",\n  ")), nil
}

// SQLStats holds result of SQL execution.
type SQLStats struct {
	Total   int
	Success int
	Failed  int
	Errors  []string
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}

// QuoteIdentifier quotes a database identifier.
func QuoteIdentifier(name, quote string) string {
	return quote + strings.ReplaceAll(name, quote, quote+quote) + quote
}

// QuoteValue quotes a string value for SQL.
func QuoteValue(v string) string {
	return `'` + strings.ReplaceAll(v, `'`, `''`) + `'`
}

// IsQuerySQL returns true if SQL is a query.
func IsQuerySQL(sql string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	return strings.HasPrefix(upper, `SELECT`) ||
		strings.HasPrefix(upper, `SHOW`) ||
		strings.HasPrefix(upper, `DESCRIBE`) ||
		strings.HasPrefix(upper, `EXPLAIN`) ||
		strings.HasPrefix(upper, `WITH`)
}
