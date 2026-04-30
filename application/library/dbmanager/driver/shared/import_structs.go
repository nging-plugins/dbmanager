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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/admpub/collate"
	"github.com/admpub/log"
	"github.com/webx-top/com"
)

type ImportFile struct {
	delDirs     []string
	StructFiles []string
	DataFiles   []string
}

func (a *ImportFile) Close() error {
	for _, delDir := range a.delDirs {
		os.RemoveAll(delDir)
	}
	for _, sqlFile := range a.StructFiles {
		if !com.FileExists(sqlFile) {
			continue
		}
		os.Remove(sqlFile)
	}
	return nil
}

func (a *ImportFile) AllSqlFiles() []string {
	files := make([]string, 0, len(a.StructFiles)+len(a.DataFiles))
	files = append(files, a.StructFiles...)
	files = append(files, a.DataFiles...)
	return files
}

func (a *ImportFile) AllSqlFileNum() int {
	return len(a.StructFiles) + len(a.DataFiles)
}

func ParseImportFile(cacheDir string, files []string) (*ImportFile, error) {
	var (
		delDirs        []string
		sqlStructFiles []string
		sqlDataFiles   []string
	)
	nowTime := com.String(time.Now().Unix())
	for index, sqlFile := range files {
		extension := strings.ToLower(filepath.Ext(sqlFile))
		switch extension {
		case `.sql`:
			if strings.Contains(filepath.Base(sqlFile), `struct`) {
				sqlStructFiles = append(sqlStructFiles, sqlFile)
			} else {
				sqlDataFiles = append(sqlDataFiles, sqlFile)
			}
		case `.gz`:
			if !strings.HasSuffix(sqlFile, `.tar.gz`) {
				continue
			}
			fallthrough
		case `.zip`:
			dir := filepath.Join(cacheDir, fmt.Sprintf("upload-"+nowTime+"-%d", index))
			err := com.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return nil, err
			}
			if extension == `.gz` {
				_, err = com.UnTarGz(sqlFile, dir)
			} else {
				err = com.Unzip(sqlFile, dir)
			}
			if err != nil {
				log.Error(err)
				continue
			}
			delDirs = append(delDirs, dir)
			err = os.Remove(sqlFile)
			if err != nil {
				log.Error(err)
			}
			err = filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return err
				}
				if strings.ToLower(filepath.Ext(fpath)) != `.sql` {
					return nil
				}
				if strings.Contains(info.Name(), `struct`) {
					sqlStructFiles = append(sqlStructFiles, fpath)
					return nil
				}
				sqlDataFiles = append(sqlDataFiles, fpath)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	collate.SortStrings(sqlStructFiles)
	collate.SortStrings(sqlDataFiles)
	return &ImportFile{
		delDirs:     delDirs,
		StructFiles: sqlStructFiles,
		DataFiles:   sqlDataFiles,
	}, nil
}
