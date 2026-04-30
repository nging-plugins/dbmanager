package postgres

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/admpub/errors"
	"github.com/coscms/webcore/library/common"
	"github.com/coscms/webcore/library/notice"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/shared"
	"github.com/webx-top/com"
)

// ========== PostgreSQL tools ==========

var pgBinPaths = []string{
	`/usr/bin`, `/usr/local/bin`, `/usr/lib/postgresql/*/bin`,
}

// SupportedExport returns true if the native export tool for the given driver is available.
func SupportedCmdExport() bool {
	_, err := LookupPgDump()
	return err == nil
}

// SupportedImport returns true if the native import tool for the given driver is available.
func SupportedCmdImport() bool {
	_, err := LookupPsql()
	return err == nil
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
func NativeExportPG(ctx context.Context, cfg *shared.DBConfig, tables []string, structWriter, dataWriter io.Writer) error {
	pgDump, err := LookupPgDump()
	if err != nil {
		return err
	}
	host, port := shared.SplitHostPort(cfg.Host)
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
func NativeImportPG(ctx context.Context, cfg *shared.DBConfig, files []string, noticer *notice.NoticeAndProgress) error {
	psql, err := LookupPsql()
	if err != nil {
		return err
	}
	host, port := shared.SplitHostPort(cfg.Host)
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
