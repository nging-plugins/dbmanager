package shared

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/admpub/errors"
	"github.com/coscms/webcore/library/nsql"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/webx-top/com"
)

func NewDataTable() *DataTable {
	return &DataTable{
		Columns: []string{},
		Values:  []map[string]*sql.NullString{},
	}
}

type DataTable struct {
	Columns []string
	Values  []map[string]*sql.NullString
}

type SelectData struct {
	Result  *Result
	Data    *DataTable
	Explain *DataTable
}

type TableQuerier func(rows *sql.Rows, limit int, textLength ...int) (columns []string, r []map[string]*sql.NullString, err error)

func RunSQL(drv *driver.BaseDriver, db *sql.DB, selectTable TableQuerier) error {
	if selectTable == nil {
		selectTable = QueryTable
	}
	c := drv.Context
	driver := drv.Driver
	var err error
	selects := []*SelectData{}
	var errs []error
	if c.IsPost() {
		query := c.Form(`query`)
		query = strings.TrimSpace(query)
		errorStops := c.Formx(`error_stops`).Bool()
		onlyErrors := c.Formx(`only_errors`).Bool()
		limit := c.Formx(`limit`).Int()
		if limit <= 0 {
			limit = 50
		}
		reader := bytes.NewReader([]byte(query))
		space := "(?:\\s|/\\*[\\s\\S]*?\\*/|(?:#|-- )[^\\n]*\\n?|--\\r?\\n)"
		delimiter := ";"
		parse := `['"`
		empty := true
		switch driver {
		case `sqlite`:
			parse += "`["
		case `mssql`:
			parse += "["
		default:
			if strings.Contains(driver, `sql`) {
				parse += "`#"
			}
		}
		parse += "]|/\\*|-- |$"
		switch driver {
		case `sqlite`:
			parse += "|\\$[^$]*\\$"
		}
		buf := make([]byte, 1e6)
		query = ``
		offset := 0
		for {
			n, e := reader.Read(buf)
			if e != nil {
				if e == io.EOF {
					break
				}
				c.Logger().Error(err)
				errs = append(errs, err)
			}
			q := string(buf[0:n])
			if offset == 0 {
				if match := regexp.MustCompile("(?i)^" + space + "*DELIMITER\\s+(\\S+)").FindStringSubmatch(q); len(match) > 1 {
					delimiter = match[1]
					q = q[len(match[0]):]
					query += q
					offset += n
					continue
				}
			}
			query += q
			offset += n

			/*/ 跳过注释和空白
			match := regexp.MustCompile("(" + regexp.QuoteMeta(delimiter) + "\\s*|" + parse + ")").FindStringSubmatch(query)
			com.Dump(match)
			if len(match) > 1 {
				found := match[1]
				if strings.TrimRight(query, " \t\n\r") != delimiter {
					rule := `(?s)`
					switch found {
					case `/*`:
						rule += "\\*\/"
					case `[`:
						rule += `]`
					default:
						match := regexp.MustCompile("^-- |^#").FindStringSubmatch(found)
						if len(match) > 1 {
							rule += "\n"
						} else {
							rule += regexp.QuoteMeta(found) + "|\\\\."
						}
					}
					pos := strings.Index(query, found)
					query = query[:pos]
					rule += `|$`
					match := regexp.MustCompile(rule).FindStringSubmatch(query)
					for len(match) > 0 {
						n, e := reader.Read(buf)
						if e != nil {
							if e == io.EOF {
								break
							}
							m.Logger().Error(err)
						}
						q := string(buf[0:n])
						if len(match) > 1 && len(match[1]) > 0 && match[1][0] != '\\' {
							break
						}
						match = regexp.MustCompile(rule).FindStringSubmatch(q)
					}
				}
			}
			// */

			empty = false
			if driver == `sqlite` && regexp.MustCompile(`(?i)^`+space+`*ATTACH\b`).MatchString(query) {
				if errorStops {
					err = errors.New(c.T(`ATTACH queries are not supported.`))
					break
				}
			}

			if regexp.MustCompile(`(?i)^` + space + `*USE\b`).MatchString(query) {
				_, err = db.Exec(query)
				if err != nil {
					c.Logger().Error(err, query)
					if onlyErrors {
						return err
					}
					errs = append(errs, err)
				}
				continue
			}

			if regexp.MustCompile(`(?i)^` + space + `*(CREATE|DROP|ALTER)` + space + `+(DATABASE|SCHEMA)\b`).MatchString(query) {
				_, err = db.Exec(query)
				if err != nil {
					c.Logger().Error(err, query)
					if onlyErrors {
						return err
					}
					errs = append(errs, err)
				}
				continue
			}

			if !regexp.MustCompile(`(?i)^(` + space + `|\()*(SELECT|SHOW|EXPLAIN|DESC|DESCRIBE)\b`).MatchString(query) {
				execute := nsql.SQLLineParser(func(sqlStr string) error {
					r := &Result{
						SQL: sqlStr,
					}
					r.SetDB(db).Exec(nil)
					drv.AddResults(r)
					return r.Error()
				})
				if !strings.HasSuffix(strings.TrimSpace(query), `;`) {
					query += `;`
				}
				for _, line := range strings.Split(query, "\n") {
					line = strings.TrimSpace(line)
					if len(line) == 0 {
						continue
					}
					err = execute(line)
					if err != nil {
						c.Logger().Error(err, line)
						if onlyErrors {
							return err
						}
						errs = append(errs, err)
					}
				}
				continue
			}
			execute := nsql.SQLLineParser(func(sqlStr string) error {
				r := &Result{
					SQL: sqlStr,
				}
				dt := &DataTable{}
				r.SetDB(db).Query(nil, func(rows *sql.Rows) error {
					dt.Columns, dt.Values, err = selectTable(rows, limit)
					return err
				})
				if r.Error() != nil {
					return fmt.Errorf(`%w: %s`, r.Error(), sqlStr)
				}
				selectData := &SelectData{Result: r, Data: dt}
				if regexp.MustCompile(`(?i)^(` + space + `|\()*SELECT\b`).MatchString(sqlStr) {
					var rows *sql.Rows
					sqlStr = `EXPLAIN ` + sqlStr
					rows, err = db.Query(sqlStr)
					if err != nil {
						return fmt.Errorf(`%w: %s`, r.Error(), sqlStr)
					}
					dt := &DataTable{}
					dt.Columns, dt.Values, err = selectTable(rows, limit)
					if err != nil {
						return fmt.Errorf(`%w: %s`, r.Error(), sqlStr)
					}
					selectData.Explain = dt
				}
				selects = append(selects, selectData)
				return nil
			})
			if !strings.HasSuffix(strings.TrimSpace(query), `;`) {
				query += `;`
			}
			for _, line := range strings.Split(query, "\n") {
				line = strings.TrimSpace(line)
				if len(line) == 0 {
					continue
				}
				err = execute(line)
				if err != nil {
					c.Logger().Error(err.Error())
					if onlyErrors {
						return err
					}
					errs = append(errs, err)
				}
			}

			/*
				com.Dump(columns)
				com.Dump(values)
			// */
		}
		_ = delimiter
		_ = empty
	}
	c.Set(`selects`, selects)
	if len(errs) > 0 {
		errMessages := make([]string, len(errs))
		for i, e := range errs {
			errMessages[i] = e.Error()
		}
		err = c.E(strings.Join(errMessages, "\n"))
	}
	return err
}

func QueryTable(rows *sql.Rows, limit int, textLength ...int) (columns []string, r []map[string]*sql.NullString, err error) {
	r = []map[string]*sql.NullString{}
	columns, err = QueryNext(rows, func(_ []string, row map[string]*sql.NullString) error {
		r = append(r, row)
		return nil
	}, limit, textLength...)
	return
}

func QueryNext(rows *sql.Rows, callback func(columns []string, row map[string]*sql.NullString) error, limit int, textLength ...int) (columns []string, err error) {
	columns, err = rows.Columns()
	if err != nil {
		return
	}
	size := len(columns)
	var maxLen int
	if len(textLength) > 0 {
		maxLen = textLength[0]
	}
	for i := 0; (limit < 0 || i < limit) && rows.Next(); i++ {
		values := make([]interface{}, size)
		for k := range columns {
			values[k] = &sql.NullString{}
		}
		err = rows.Scan(values...)
		if err != nil {
			return
		}
		val := map[string]*sql.NullString{}
		for k, colName := range columns {
			val[colName] = values[k].(*sql.NullString)
			if maxLen > 0 {
				val[colName].String = com.Substr(val[colName].String, ` ...`, maxLen)
			}
		}
		err = callback(columns, val)
		if err != nil {
			return
		}
	}
	return
}
