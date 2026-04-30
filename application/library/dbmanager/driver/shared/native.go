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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/admpub/errors"
	"github.com/webx-top/com"

	"github.com/coscms/webcore/library/common"
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

// SupportedExport returns true if the native export tool for the given driver is available.
func SupportedExport(driver string) bool {
	switch driver {
	case `postgres`:
		_, err := LookupPgDump()
		return err == nil
	case `clickhouse`:
		_, err := LookupClickHouseClient()
		return err == nil
	}
	return false
}

// SupportedImport returns true if the native import tool for the given driver is available.
func SupportedImport(driver string) bool {
	switch driver {
	case `postgres`:
		_, err := LookupPsql()
		return err == nil
	case `clickhouse`:
		_, err := LookupClickHouseClient()
		return err == nil
	}
	return false
}

// ========== PostgreSQL tools ==========

var pgBinPaths = []string{
	`/usr/bin`, `/usr/local/bin`, `/usr/lib/postgresql/*/bin`,
}

// LookupPgDump finds pg_dump binary.
func LookupPgDump() (string, error) {
	return common.LookPath(`pg_dump`, pgBinPaths...)
}

// LookupPsql finds psql binary.
func LookupPsql() (string, error) {
	return common.LookPath(`psql`, pgBinPaths...)
}

// NativeExportPG uses pg_dump to export tables.
func NativeExportPG(ctx context.Context, cfg *DBConfig, tables []string, structWriter, dataWriter io.Writer) error {
	pgDump, err := LookupPgDump()
	if err != nil {
		return err
	}
	host, port := SplitHostPort(cfg.Host)
	if len(port) == 0 {
		port = `5432`
	}
	if len(cfg.Database) == 0 {
		return errors.New(`database name is required`)
	}

	// Common args
	baseArgs := []string{
		`-h`, host,
		`-p`, port,
		`-U`, cfg.Username,
		`--no-password`,
		`--no-owner`,
	}

	if structWriter != nil {
		args := append(baseArgs, `-s`, `-f`, `-`, cfg.Database)
		if len(tables) > 0 {
			args = append(args, `-t`, strings.Join(tables, `|`))
		}
		cmd := exec.CommandContext(ctx, pgDump, args...)
		cmd.Env = append(os.Environ(), `PGPASSWORD=`+cfg.Password)
		cmd.Stdout = structWriter
		cmd.Stderr = structWriter
		if err := cmd.Run(); err != nil {
			return fmt.Errorf(`pg_dump struct: %w`, err)
		}
	}

	if dataWriter != nil {
		args := append(baseArgs, `-a`, `--inserts`, `-f`, `-`, cfg.Database)
		if len(tables) > 0 {
			args = append(args, `-t`, strings.Join(tables, `|`))
		}
		cmd := exec.CommandContext(ctx, pgDump, args...)
		cmd.Env = append(os.Environ(), `PGPASSWORD=`+cfg.Password)
		cmd.Stdout = dataWriter
		cmd.Stderr = dataWriter
		if err := cmd.Run(); err != nil {
			return fmt.Errorf(`pg_dump data: %w`, err)
		}
	}
	return nil
}

// NativeImportPG uses psql to import SQL files.
func NativeImportPG(ctx context.Context, cfg *DBConfig, files []string, noticer *notice.NoticeAndProgress) error {
	psql, err := LookupPsql()
	if err != nil {
		return err
	}
	host, port := SplitHostPort(cfg.Host)
	if len(port) == 0 {
		port = `5432`
	}
	if len(cfg.Database) == 0 {
		return errors.New(`database name is required`)
	}

	for _, file := range files {
		noticer.Success(`导入: ` + filepath.Base(file))
		cmd := exec.CommandContext(ctx, psql,
			`-h`, host,
			`-p`, port,
			`-U`, cfg.Username,
			`-d`, cfg.Database,
			`-f`, file,
		)
		cmd.Env = append(os.Environ(), `PGPASSWORD=`+cfg.Password)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf(`psql: %w\n%s`, err, string(com.Bytes2str(output)))
		}
	}
	return nil
}

// ========== ClickHouse tools ==========

var chBinPaths = []string{
	`/usr/bin`, `/usr/local/bin`, `/opt/clickhouse`,
}

// LookupClickHouseClient finds clickhouse-client binary.
func LookupClickHouseClient() (string, error) {
	return common.LookPath(`clickhouse-client`, chBinPaths...)
}

// NativeExportCH uses clickhouse-client to export tables as INSERT statements.
func NativeExportCH(ctx context.Context, cfg *DBConfig, tables []string, structWriter, dataWriter io.Writer) error {
	client, err := LookupClickHouseClient()
	if err != nil {
		return err
	}
	host, port := SplitHostPort(cfg.Host)
	if len(port) == 0 {
		port = `9000`
	}

	for _, table := range tables {
		query := fmt.Sprintf(`SELECT * FROM %s FORMAT TabSeparated`, table)

		if structWriter != nil {
			// Export struct info
			descQuery := fmt.Sprintf(`DESCRIBE TABLE %s FORMAT Pretty`, table)
			cmd := exec.CommandContext(ctx, client,
				`-h`, host,
				`--port`, port,
				`-u`, cfg.Username,
				`--password`, cfg.Password,
				`-d`, cfg.Database,
				`-q`, descQuery,
			)
			cmd.Stdout = structWriter
			cmd.Stderr = structWriter
			if err := cmd.Run(); err != nil {
				return fmt.Errorf(`clickhouse-client describe %s: %w`, table, err)
			}
			fmt.Fprintf(structWriter, "\n-- Table: %s\n", table)
		}

		if dataWriter != nil {
			cmd := exec.CommandContext(ctx, client,
				`-h`, host,
				`--port`, port,
				`-u`, cfg.Username,
				`--password`, cfg.Password,
				`-d`, cfg.Database,
				`-q`, query,
			)
			cmd.Stdout = dataWriter
			cmd.Stderr = dataWriter
			if err := cmd.Run(); err != nil {
				return fmt.Errorf(`clickhouse-client export %s: %w`, table, err)
			}
		}
	}
	return nil
}

// NativeImportCH uses clickhouse-client to import SQL files.
func NativeImportCH(ctx context.Context, cfg *DBConfig, files []string, noticer *notice.NoticeAndProgress) error {
	client, err := LookupClickHouseClient()
	if err != nil {
		return err
	}
	host, port := SplitHostPort(cfg.Host)
	if len(port) == 0 {
		port = `9000`
	}

	for _, file := range files {
		noticer.Success(`导入: ` + filepath.Base(file))
		cmd := exec.CommandContext(ctx, client,
			`-h`, host,
			`--port`, port,
			`-u`, cfg.Username,
			`--password`, cfg.Password,
			`-d`, cfg.Database,
			`-mn`, // multiline query mode
			`<`, file,
		)
		cmd.Stdin, _ = os.Open(file)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf(`clickhouse-client: %w\n%s`, err, string(com.Bytes2str(output)))
		}
	}
	return nil
}
