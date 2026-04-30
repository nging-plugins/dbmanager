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
	"context"
	"io"

	"github.com/webx-top/com"

	"github.com/coscms/webcore/library/notice"
)

// ExportDriver describes capabilities of a native export tool.
type ExportDriver struct {
	Name       string
	BinName    string
	BinPaths   []string
	Lookup     func() (string, error)
	ExportFunc func(ctx context.Context, cfg *DBConfig, tables []string, structWriter, dataWriter io.Writer) error
}

// ImportDriver describes capabilities of a native import tool.
type ImportDriver struct {
	Name       string
	BinName    string
	BinPaths   []string
	Lookup     func() (string, error)
	ImportFunc func(ctx context.Context, cfg *DBConfig, files []string, noticer *notice.NoticeAndProgress) error
}

// DBConfig holds connection parameters for native tools.
type DBConfig struct {
	Driver   string
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// SplitHostPort splits host:port into separate values.
func SplitHostPort(host string) (string, string) {
	return com.SplitHostPort(host)
}
