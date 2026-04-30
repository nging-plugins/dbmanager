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
	"net/url"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/admpub/errors"
	"github.com/coscms/webcore/library/backend"
	"github.com/webx-top/echo"

	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/shared"
)

func init() {
	driver.Register(`clickhouse`, `ClickHouse`, New)
}

func New() driver.Driver {
	return &ClickHouse{
		supportSQL: true,
	}
}

type ClickHouse struct {
	*driver.BaseDriver
	db         *sql.DB
	dbName     string
	version    string
	supportSQL bool
}

func (c *ClickHouse) Name() string {
	return `ClickHouse`
}

func (c *ClickHouse) Init(ctx echo.Context, auth *driver.DbAuth) {
	c.BaseDriver = driver.NewBaseDriver()
	c.BaseDriver.Init(ctx, auth)
	c.Set(`supportSQL`, c.supportSQL)
}

func (c *ClickHouse) IsSupported(operation string) bool {
	return true
}

var ErrConnectTimeout = errors.New(`连接超时，请重试`)

func (c *ClickHouse) buildDSN(dbName ...string) string {
	host := c.DbAuth.Host
	if len(host) == 0 {
		host = `127.0.0.1`
	}
	port := `9000`
	if strings.Contains(host, `:`) {
		parts := strings.SplitN(host, `:`, 2)
		host = parts[0]
		port = parts[1]
	}

	user := c.DbAuth.Username
	if len(user) == 0 {
		user = `default`
	}
	password := c.DbAuth.Password

	database := `default`
	if len(dbName) > 0 && len(dbName[0]) > 0 {
		database = dbName[0]
	} else if len(c.DbAuth.Db) > 0 {
		database = c.DbAuth.Db
	}

	// ClickHouse DSN format: clickhouse://username:password@host:port/database
	dsn := fmt.Sprintf(
		"clickhouse://%s:%s@%s:%s/%s",
		url.QueryEscape(user), url.QueryEscape(password), host, port, database,
	)
	return dsn
}

func (c *ClickHouse) connect(dbName ...string) error {
	dsn := c.buildDSN(dbName...)
	db, err := sql.Open(`clickhouse`, dsn)
	if err != nil {
		return err
	}
	if c.db != nil {
		c.db.Close()
	}
	c.db = db
	return db.Ping()
}

func (c *ClickHouse) Logined() bool {
	return c.db != nil
}

func (c *ClickHouse) Login() error {
	c.dbName = c.Form(`db`)
	err := c.connect(c.dbName)
	if err != nil {
		return err
	}
	c.Set(`dbName`, c.dbName)
	c.Set(`table`, c.Form(`table`))
	return c.baseInfo()
}

func (c *ClickHouse) Logout() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *ClickHouse) baseInfo() error {
	if c.Get(`dbList`) == nil {
		dbList, err := c.getDatabases()
		if err != nil {
			c.fail(err.Error())
			return c.returnTo(backend.URLFor(`/db`))
		}
		c.Set(`dbList`, dbList)
	}
	if len(c.dbName) > 0 {
		tableList, err := c.getTables()
		if err != nil {
			c.fail(err.Error())
			return c.returnTo(c.GenURL(`listDb`))
		}
		c.Set(`tableList`, tableList)
	}
	c.Set(`dbVersion`, c.getVersion())
	return nil
}

func (c *ClickHouse) ok(msg string) {
	c.SetOk(msg)
}

func (c *ClickHouse) checkErr(err error) interface{} {
	return c.CheckErr(err)
}

func (c *ClickHouse) fail(msg string) {
	c.SetFail(msg)
}

func (c *ClickHouse) returnTo(urls ...string) error {
	return c.Goto(urls...)
}

func (m *ClickHouse) bgExecManage(op string) error {
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

// QuoteCol quotes a column/identifier name for ClickHouse
func QuoteCol(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

// QuoteVal quotes a value string for ClickHouse
func QuoteVal(v string) string {
	return `'` + strings.ReplaceAll(v, `'`, `''`) + `'`
}
