package mysql

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultSQLToHTML(t *testing.T) {
	r := &Result{}
	result := r.ToHTML(``, "SELECT * FROM `a` WHERE")
	assert.Equal(t, template.HTML("SELECT * FROM `<a href=\"&table=a\">a</a>` WHERE"), result)

	result = r.ToHTML(``, "DELETE from `a` WHERE")
	assert.Equal(t, template.HTML("DELETE from `<a href=\"&table=a\">a</a>` WHERE"), result)

	result = r.ToHTML(``, "INSERT INTO\r\n`a` WHERE")
	assert.Equal(t, template.HTML("INSERT INTO\r\n`<a href=\"&table=a\">a</a>` WHERE"), result)

	result = r.ToHTML(``, "UPDATE `a` SET")
	assert.Equal(t, template.HTML("UPDATE `<a href=\"&table=a\">a</a>` SET"), result)
}
