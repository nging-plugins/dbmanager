package utils

import (
	"context"
	"testing"

	"github.com/admpub/nging/v5/application/library/notice"
	"github.com/stretchr/testify/assert"
	"github.com/webx-top/echo/defaults"
)

func TestAllSqlFileNum(t *testing.T) {
	a := &ImportFile{
		StructFiles: []string{`a`},
		DataFiles:   []string{`b`},
	}
	assert.Equal(t, 2, a.AllSqlFileNum())
}
