package mysql

import (
	"testing"

	"github.com/webx-top/echo"
	"github.com/webx-top/echo/testing/test"
)

func TestEnumOptions(t *testing.T) {
	s := `'ab','bc'`
	test.True(t, reContainsEnumLength.MatchString(s))
	matches := reEnumLength.FindAllStringSubmatch(s, -1)
	echo.Dump(matches)
	var r string
	for index, values := range matches {
		if index > 0 {
			r += `,`
		}
		r += values[0]
	}
	r = "(" + r + ")"
	test.Eq(t, "('ab','bc')", r)
}

func TestMatchGeneratedColumn(t *testing.T) {
	s := "CREATE TABLE `locations` (" + "\n" +
		"  `id` int(11) NOT NULL AUTO_INCREMENT," + "\n" +
		"  `food_type` varchar(25) GENERATED ALWAYS AS (json_value(`attr`,'$.details.foodType')) VIRTUAL," + "\n" +
		"  `food_type2` varchar(25) GENERATED ALWAYS AS (json_value(`attr`,'$.details.foodType')) VIRTUAL," + "\n" +
		"  PRIMARY KEY (`id`)," + "\n" +
		"  KEY `foodtypes` (`food_type`)" + "\n" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci"
	test.True(t, reGeneratedColumn.MatchString(s))
	matches := reGeneratedColumn.FindAllStringSubmatch(s, -1)
	test.Eq(t, [][]string{
		{
			"`food_type` varchar(25) GENERATED ALWAYS AS (json_value(`attr`,'$.details.foodType')) VIRTUAL",
			"food_type",
			"json_value(`attr`,'$.details.foodType')",
		},
		{
			"`food_type2` varchar(25) GENERATED ALWAYS AS (json_value(`attr`,'$.details.foodType')) VIRTUAL",
			"food_type2",
			"json_value(`attr`,'$.details.foodType')",
		},
	}, matches, false)
}
