package clickhouse

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/coscms/webcore/library/common"
	"github.com/coscms/webcore/library/notice"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver/shared"
	"github.com/webx-top/com"
)

// ========== ClickHouse tools ==========

var chBinPaths = []string{
	`/usr/bin`, `/usr/local/bin`, `/opt/clickhouse`,
}

// SupportedExport returns true if the native export tool for the given driver is available.
func SupportedCmdExport() bool {
	_, err := LookupClickHouseClient()
	return err == nil
}

// SupportedImport returns true if the native import tool for the given driver is available.
func SupportedCmdImport() bool {
	_, err := LookupClickHouseClient()
	return err == nil
}

// LookupClickHouseClient finds clickhouse-client binary.
func LookupClickHouseClient() (string, error) {
	return common.LookPath(`clickhouse-client`, chBinPaths...)
}

// NativeExportCH uses clickhouse-client to export tables as INSERT statements.
func NativeExportCH(ctx context.Context, cfg *shared.DBConfig, tables []string, structWriter, dataWriter io.Writer) error {
	client, err := LookupClickHouseClient()
	if err != nil {
		return err
	}
	host, port := shared.SplitHostPort(cfg.Host)
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
func NativeImportCH(ctx context.Context, cfg *shared.DBConfig, files []string, noticer *notice.NoticeAndProgress) error {
	client, err := LookupClickHouseClient()
	if err != nil {
		return err
	}
	host, port := shared.SplitHostPort(cfg.Host)
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
