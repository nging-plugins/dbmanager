package common

import (
	"testing"

	"github.com/webx-top/echo/testing/test"
)

func TestSortStrings(t *testing.T) {
	ss := []string{`a_1`, `a_11`, `a_2`}
	SortStrings(ss)
	test.Eq(t, []string{`a_1`, `a_2`, `a_11`}, ss)
}
