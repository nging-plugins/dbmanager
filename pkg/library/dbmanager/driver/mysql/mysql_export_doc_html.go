package mysql

import (
	"github.com/admpub/nging/v4/application/library/common"
	"github.com/webx-top/echo"
)

func newHTMLDocExportor(dbName string) DocExportor {
	return &mysqlExportHTMLDoc{
		dbName: dbName,
	}
}

type mysqlExportHTMLDoc struct {
	dbName string
}

func (a *mysqlExportHTMLDoc) Open(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEEventStream)
	encodedName := echo.URLEncode(a.dbName+`_doc.html`, true)
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+encodedName+"; filename*=utf-8''"+encodedName)
	c.Response().Write([]byte(`<doctype html><html><head><title>Table Documention</title><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="stylesheet" href="` + common.BackendURL(c) + `/public/assets/backend/js/bootstrap/dist/css/bootstrap.min.css?t=20220807115920" /></head><body><div class="container"><div class="row"><div class="col-md-12">`))
	return nil
}

func (a *mysqlExportHTMLDoc) Write(c echo.Context, table *TableStatus, fields []*Field) error {
	c.Response().Write([]byte(`<h2>` + table.Name.String + `</h2>`))
	c.Response().Write([]byte(`<em>` + table.Comment.String + `</em>`))
	c.Response().Write([]byte(`<table class="table table-bordered table-hover table-condensed">`))
	c.Response().Write([]byte(`<thead><tr><th>` + c.T(`字段名`) + `</th><th>` + c.T(`数据类型`) + `</th><th>` + c.T(`说明`) + `</th></tr></thead>`))
	c.Response().Write([]byte(`<tbody>`))
	for _, v := range fields {
		dataType := v.Full_type
		if v.Null {
			dataType += ` <span>NULL</span>`
		}
		if v.AutoIncrement.Valid {
			dataType += ` <em>` + c.T("自动增量") + `</em>`
		}
		if v.Default.Valid {
			dataType += ` [<b>` + v.Default.String + `</b>]`
		}
		if len(v.On_update) > 0 {
			dataType += ` ON UPDATE <b>` + v.On_update + `</b>`
		}
		c.Response().Write([]byte(`<tr><td>` + v.Field + `</td><td>` + dataType + `</td><td>` + v.Comment + `</td></tr>`))
	}
	c.Response().Write([]byte(`</tbody>`))
	c.Response().Write([]byte(`</table>`))
	return nil
}

func (a *mysqlExportHTMLDoc) Close(c echo.Context) error {
	c.Response().Write([]byte(`</div></div></div></body></html>`))
	return nil
}
