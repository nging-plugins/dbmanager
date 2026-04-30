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
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/admpub/errors"
	"github.com/coscms/webcore/library/backend"
	_ "github.com/lib/pq"
	"github.com/webx-top/echo"

	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
)

func init() {
	driver.Register(`postgres`, `PostgreSQL`, New)
}

func New() driver.Driver {
	return &Postgres{
		supportSQL: true,
	}
}

type Postgres struct {
	*driver.BaseDriver
	db         *sql.DB
	dbName     string
	version    string
	supportSQL bool
}

func (p *Postgres) Name() string {
	return `PostgreSQL`
}

func (p *Postgres) Init(ctx echo.Context, auth *driver.DbAuth) {
	p.BaseDriver = driver.NewBaseDriver()
	p.BaseDriver.Init(ctx, auth)
	p.Set(`supportSQL`, p.supportSQL)
}

func (p *Postgres) IsSupported(operation string) bool {
	return true
}

var ErrConnectTimeout = errors.New(`连接超时，请重试`)

func (p *Postgres) buildDSN(dbName ...string) string {
	host := p.DbAuth.Host
	if len(host) == 0 {
		host = `127.0.0.1`
	}
	port := `5432`
	if strings.Contains(host, `:`) {
		parts := strings.SplitN(host, `:`, 2)
		host = parts[0]
		port = parts[1]
	}

	user := p.DbAuth.Username
	if len(user) == 0 {
		user = `postgres`
	}
	password := p.DbAuth.Password

	database := `postgres`
	if len(dbName) > 0 && len(dbName[0]) > 0 {
		database = dbName[0]
	} else if len(p.DbAuth.Db) > 0 {
		database = p.DbAuth.Db
	}

	sslmode := `disable`
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslmode,
	)
	return dsn
}

func (p *Postgres) connect(dbName ...string) error {
	dsn := p.buildDSN(dbName...)
	db, err := sql.Open(`postgres`, dsn)
	if err != nil {
		return err
	}
	if p.db != nil {
		p.db.Close()
	}
	p.db = db
	return db.Ping()
}

func (p *Postgres) Logined() bool {
	return p.db != nil
}

func (p *Postgres) Login() error {
	p.dbName = p.Form(`db`)
	err := p.connect(p.dbName)
	if err != nil {
		return err
	}
	p.Set(`dbName`, p.dbName)
	p.Set(`table`, p.Form(`table`))
	return p.baseInfo()
}

func (p *Postgres) Logout() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// genURL returns a URL generator for the current context
func (p *Postgres) genURL(op string, args ...string) string {
	return p.GenURL(op, args...)
}

func (p *Postgres) newParam() *sql.DB {
	return p.db
}

func (p *Postgres) baseInfo() error {
	if p.Get(`dbList`) == nil {
		dbList, err := p.getDatabases()
		if err != nil {
			p.fail(err.Error())
			return p.returnTo(backend.URLFor(`/db`))
		}
		p.Set(`dbList`, dbList)
	}
	if len(p.dbName) > 0 {
		tableList, err := p.getTables()
		if err != nil {
			p.fail(err.Error())
			return p.returnTo(p.GenURL(`listDb`))
		}
		p.Set(`tableList`, tableList)
	}
	p.Set(`dbVersion`, p.getVersion())
	return nil
}

func (p *Postgres) ok(msg string) {
	p.SetOk(msg)
}

func (p *Postgres) checkErr(err error) interface{} {
	return p.CheckErr(err)
}

func (p *Postgres) fail(msg string) {
	p.SetFail(msg)
}

func (p *Postgres) returnTo(urls ...string) error {
	return p.Goto(urls...)
}

// QuoteCol quotes a column/identifier name for PostgreSQL
func QuoteCol(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// QuoteVal quotes a value string for PostgreSQL
func QuoteVal(v string) string {
	return `'` + strings.ReplaceAll(v, `'`, `''`) + `'`
}

// EncodeDSN encodes the DSN for URL safe transport
func EncodeDSN(dsn string) string {
	return url.QueryEscape(dsn)
}

func (p *Postgres) getVersion() string {
	if len(p.version) > 0 {
		return p.version
	}
	var v string
	p.db.QueryRow(`SELECT version()`).Scan(&v)
	p.version = v
	return v
}
