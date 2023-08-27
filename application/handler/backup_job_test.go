package handler

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/webx-top/echo"
)

func TestFileGlob(t *testing.T) {
	absPath, _ := filepath.Abs(`.`)
	matches, err := filepath.Glob(absPath + echo.FilePathSeparator + `*.go`)
	if err != nil {
		t.Error(err)
	}

	for _, file := range matches {
		fmt.Println(file)
	}
}
