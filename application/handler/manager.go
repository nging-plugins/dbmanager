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

package handler

import (
	"fmt"
	"net/url"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/common"
	"github.com/webx-top/db"
	"github.com/webx-top/echo"

	"github.com/admpub/log"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/mysql"   //mysql
	_ "github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/redis" //redis
	"github.com/nging-plugins/dbmanager/application/model"
)

var (
	defaultGenURL = func(_ string, _ ...string) string {
		return ``
	}
	defaultGenBaseURL = func(auth *driver.DbAuth) string {
		baseURL := backend.URLFor(`/db`)
		if auth.AccountID > 0 {
			return baseURL + `?accountId=` + fmt.Sprint(auth.AccountID)
		}
		return baseURL + `?driver=` + auth.Driver + `&host=` + url.QueryEscape(auth.Host) + `&username=` + url.QueryEscape(auth.Username)
	}
	driverColors = map[string]string{
		`mysql`: `primary`,
		`redis`: `info`,
	}
)

func Manager(ctx echo.Context) error {
	user := backend.User(ctx)
	m := model.NewDbAccount(ctx)
	var err error
	driverName := ctx.Form(`driver`)
	operation := ctx.Form(`operation`)
	var accountID uint
	if user != nil {
		accountID = ctx.Formx(`accountId`).Uint()
		if accountID > 0 {
			err = m.Get(nil, db.And(
				db.Cond{`id`: accountID},
				db.Cond{`uid`: user.Id},
			))
			if err != nil && db.ErrNoMoreRows != err {
				log.Error(err)
			}
			accountID = m.Id
			driverName = m.Engine
		}
	}
	auth := &driver.DbAuth{
		Driver:    driverName,
		AccountID: accountID,
		Username:  ctx.Form(`username`),
		Host:      ctx.Form(`host`),
		Db:        ctx.Form(`db`),
	}
	mgr := dbmanager.New(ctx, auth)
	genURL := defaultGenURL
	switch operation {
	case `login`:
		if accountID > 0 {
			err, _ = authentication(mgr, m)
			if err != nil {
				deleteAuth(ctx, auth)
				common.SendFail(ctx, err.Error())
				err = nil
			}
		} else if err = getLoginInfo(mgr, accountID, m, user); err != nil {
			deleteAuth(ctx, auth)
			return err
		}

	case `logout`:
		_, _ = authentication(mgr, m)

	case `logoutAll`:
		clearAuth(ctx)

	default:
		var signedIn bool
		err, signedIn = authentication(mgr, m)
		ctx.Set(`signedIn`, signedIn)
		ctx.Set(`dbUsername`, auth.Username)
		ctx.Set(`dbHost`, auth.Host)
		ctx.Set(`accountTitle`, auth.AccountTitle)
		if signedIn {
			driverName = auth.Driver
			if len(operation) == 0 {
				operation = `listDb`
			}
			genURL = func(op string, args ...string) string {
				if len(op) == 0 {
					op = operation
				}
				var p string
				switch len(args) {
				case 2:
					p += `&table=` + args[1]
					fallthrough
				case 1:
					p += `&db=` + args[0]
				}
				return defaultGenBaseURL(auth) + `&operation=` + op + p
			}
			defer mgr.Run(`logout`)
		} else {
			if err != nil { //登录失败
				deleteAuth(ctx, auth)
				common.SendFail(ctx, err.Error())
				err = nil
			}
			driverName = ``
		}
	}
	mgr.SetURLGenerator(genURL)
	ctx.SetFunc(`dbMgrURL`, genURL)
	ctx.Set(`operation`, operation)
	if len(driverName) > 0 {
		ctx.Set(`driver`, driverName)
		if err == nil {
			err = mgr.Run(operation)
		}
		if err == nil {
			switch operation {
			case `login`: //登录成功
				addAuth(ctx, auth)
				mgr.Run(`logout`)
				return ctx.Redirect(defaultGenBaseURL(auth))
			case `logout`: //退出登录
				deleteAuth(ctx, auth)
			default:
				return err
			}
		} else {
			if operation != `login` {
				return err
			}
		}
	}
	ret := common.Err(ctx, err)
	driverList := map[string]string{}
	for driverName, driver := range driver.GetAll() {
		driverList[driverName] = driver.Name()
	}

	ctx.Set(`driverList`, driverList)
	ctx.Set(`dbType`, ctx.T(`数据库`))
	ctx.Set(`charsetList`, mysql.Charsets)
	ctx.Set(`accounts`, getAccounts(ctx))
	ctx.SetFunc(`dbMgrURLByAccount`, defaultGenBaseURL)
	ctx.SetFunc(`colorByDriver`, func(driver string) string {
		if color, ok := driverColors[driver]; ok {
			return color
		}
		return `default`
	})
	return ctx.Render(`db/index`, ret)
}
