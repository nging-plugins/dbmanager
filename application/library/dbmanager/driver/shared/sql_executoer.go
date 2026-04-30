package shared

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/coscms/webcore/library/notice"
	"github.com/nging-plugins/dbmanager/application/library/dbmanager/driver"
	"github.com/webx-top/com"
)

// NewSQLExecutor creates a line-based SQL executor for use with SeekFileLines.
// It accumulates lines into a statement buffer and executes when a complete
// statement (ending with ';') is detected. Returns final stats.
func NewSQLExecutor(db *sql.DB) *SQLExecutor {
	return &SQLExecutor{db: db}
}

// SQLExecutor executes SQL statements incrementally as they are read line by line.
type SQLExecutor struct {
	db      *sql.DB
	buf     strings.Builder
	seenSQL func(string, error)
	stats   SQLStats
}

// WriteLine processes one line of SQL content. If it completes a statement,
// the statement is executed immediately.
func (e *SQLExecutor) WriteLine(line string) error {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
		return nil
	}
	e.buf.WriteString(line)
	if strings.HasSuffix(line, ";") {
		stmt := strings.TrimSuffix(e.buf.String(), ";")
		e.buf.Reset()
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			return nil
		}
		_, err := e.db.Exec(stmt)
		e.stats.Total++
		if err != nil {
			e.stats.Failed++
			if len(e.stats.Errors) >= 100 {
				e.stats.Errors[99] = `...`
			} else {
				e.stats.Errors = append(e.stats.Errors, fmt.Sprintf("%v: %s", err, truncate(stmt, 120)))
			}
		} else {
			e.stats.Success++
		}
		if e.seenSQL != nil {
			e.seenSQL(stmt, err)
		}
		return err
	}
	e.buf.WriteByte(' ')
	return nil
}

// Flush executes any remaining un-terminated statement.
func (e *SQLExecutor) Flush() {
	remainder := strings.TrimSpace(e.buf.String())
	if remainder == "" {
		return
	}
	e.buf.Reset()
	_, err := e.db.Exec(remainder)
	e.stats.Total++
	if err != nil {
		e.stats.Failed++
		e.stats.Errors = append(e.stats.Errors, fmt.Sprintf("%v: %s", err, truncate(remainder, 120)))
	} else {
		e.stats.Success++
	}
	if e.seenSQL != nil {
		e.seenSQL(remainder, err)
	}
}

func (e *SQLExecutor) ExecuteUploadedSQLFile(file multipart.File, hdr *multipart.FileHeader) (SQLStats, error) {
	// Save to temp and execute line-by-line via SeekFileLines
	tmpFile := filepath.Join(TempDir(OpImport), hdr.Filename)
	com.MkdirAll(filepath.Dir(tmpFile), os.ModePerm)
	dst, err := os.Create(tmpFile)
	if err != nil {
		return e.Stats(), err
	}
	io.Copy(dst, file)
	dst.Close()
	defer os.Remove(tmpFile)
	com.SeekFileLines(tmpFile, e.WriteLine)
	e.Flush()
	return e.Stats(), err
}

func (e *SQLExecutor) ExecuteSQLFiles(ctx context.Context, noticer *notice.NoticeAndProgress, sqlFiles []string, ignoreErr bool) (err error) {
	exec := func(sqlFile string, callback func(strLen int)) func(string) error {
		e.seenSQL = func(sqlStr string, err error) {
			if err != nil {
				noticer.Failure(`[FAILURE] ` + err.Error() + `: ` + com.HTMLEncode(sqlStr) + `: ` + filepath.Base(sqlFile))
			} else {
				noticer.Success(`[SUCCESS] ` + filepath.Base(sqlFile))
			}
			callback(len(sqlStr))
		}
		return e.WriteLine
	}
	for _, sqlFile := range sqlFiles {
		err = executeSQLFileWithProgress(noticer, sqlFile, exec)
		if !ignoreErr && err != nil {
			return
		}
	}
	e.Flush()
	if ignoreErr {
		return nil
	}
	return err
}

// Stats returns the accumulated execution statistics.
func (e *SQLExecutor) Stats() SQLStats {
	return e.stats
}

func (e *SQLExecutor) ImportExecutor(ctx context.Context, noticer *notice.NoticeAndProgress, cfg *driver.DbAuth, cacheDir string, files []string) error {
	names := make([]string, len(files))
	for i, file := range files {
		names[i] = filepath.Base(file)
	}
	noticer.Success(`开始导入: ` + strings.Join(names, ", "))
	ifi, err := ParseImportFile(cacheDir, files)
	if err != nil {
		return err
	}
	defer ifi.Close()
	noticer.Add(int64(ifi.AllSqlFileNum()) * 100)
	if len(ifi.StructFiles) > 0 {
		err = e.ExecuteSQLFiles(ctx, noticer, ifi.StructFiles, false)
	}
	if err == nil && len(ifi.DataFiles) > 0 {
		err = e.ExecuteSQLFiles(ctx, noticer, ifi.DataFiles, true)
	}
	return err
}

func executeSQLFileWithProgress(
	noticer *notice.NoticeAndProgress,
	sqlFile string,
	exec func(sqlFile string, callback func(strLen int)) func(string) error,
) error {
	fileSize, _ := com.FileSize(sqlFile)
	return noticer.Callback(fileSize, func(callback func(int)) error {
		return com.SeekFileLines(sqlFile, exec(sqlFile, callback))
	})
}
