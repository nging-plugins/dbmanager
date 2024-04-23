package mysql

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webx-top/echo/defaults"
)

func TestDiffIndexes(t *testing.T) {
	/*
		qs, err := url.QueryUnescape(`indexes%5B0%5D%5Btype%5D=PRIMARY&indexes%5B0%5D%5Bcolumns%5D%5B%5D=id&indexes%5B0%5D%5Blengths%5D%5B%5D=&indexes%5B0%5D%5Bdescs%5D%5B%5D=ASC&indexes%5B0%5D%5Bexpressions%5D%5B%5D=&indexes%5B0%5D%5Bcolumns%5D%5B%5D=&indexes%5B0%5D%5Blengths%5D%5B%5D=&indexes%5B0%5D%5Bdescs%5D%5B%5D=&indexes%5B0%5D%5Bexpressions%5D%5B%5D=&indexes%5B0%5D%5Bwith%5D=&indexes%5B0%5D%5Bname%5D=PRIMARY&indexes%5B1%5D%5Btype%5D=INDEX&indexes%5B1%5D%5Bcolumns%5D%5B%5D=customer_id&indexes%5B1%5D%5Blengths%5D%5B%5D=&indexes%5B1%5D%5Bdescs%5D%5B%5D=ASC&indexes%5B1%5D%5Bexpressions%5D%5B%5D=&indexes%5B1%5D%5Bcolumns%5D%5B%5D=&indexes%5B1%5D%5Blengths%5D%5B%5D=&indexes%5B1%5D%5Bdescs%5D%5B%5D=&indexes%5B1%5D%5Bexpressions%5D%5B%5D=&indexes%5B1%5D%5Bwith%5D=&indexes%5B1%5D%5Bname%5D=software_license_customer_id&indexes%5B2%5D%5Btype%5D=INDEX&indexes%5B2%5D%5Bcolumns%5D%5B%5D=package_ident&indexes%5B2%5D%5Blengths%5D%5B%5D=&indexes%5B2%5D%5Bdescs%5D%5B%5D=ASC&indexes%5B2%5D%5Bexpressions%5D%5B%5D=&indexes%5B2%5D%5Bcolumns%5D%5B%5D=&indexes%5B2%5D%5Blengths%5D%5B%5D=&indexes%5B2%5D%5Bdescs%5D%5B%5D=&indexes%5B2%5D%5Bexpressions%5D%5B%5D=&indexes%5B2%5D%5Bwith%5D=&indexes%5B2%5D%5Bname%5D=software_license_package_ident&indexes%5B3%5D%5Btype%5D=INDEX&indexes%5B3%5D%5Bcolumns%5D%5B%5D=product_id&indexes%5B3%5D%5Blengths%5D%5B%5D=&indexes%5B3%5D%5Bdescs%5D%5B%5D=ASC&indexes%5B3%5D%5Bexpressions%5D%5B%5D=&indexes%5B3%5D%5Bcolumns%5D%5B%5D=version&indexes%5B3%5D%5Blengths%5D%5B%5D=&indexes%5B3%5D%5Bdescs%5D%5B%5D=ASC&indexes%5B3%5D%5Bexpressions%5D%5B%5D=&indexes%5B3%5D%5Bcolumns%5D%5B%5D=&indexes%5B3%5D%5Blengths%5D%5B%5D=&indexes%5B3%5D%5Bdescs%5D%5B%5D=&indexes%5B3%5D%5Bexpressions%5D%5B%5D=&indexes%5B3%5D%5Bwith%5D=&indexes%5B3%5D%5Bname%5D=software_license_product_id_version&indexes%5B6%5D%5Btype%5D=INDEX&indexes%5B6%5D%5Bcolumns%5D%5B%5D=sn&indexes%5B6%5D%5Blengths%5D%5B%5D=&indexes%5B6%5D%5Bdescs%5D%5B%5D=ASC&indexes%5B6%5D%5Bexpressions%5D%5B%5D=&indexes%5B6%5D%5Bcolumns%5D%5B%5D=&indexes%5B6%5D%5Blengths%5D%5B%5D=&indexes%5B6%5D%5Bdescs%5D%5B%5D=&indexes%5B6%5D%5Bexpressions%5D%5B%5D=&indexes%5B6%5D%5Bwith%5D=&indexes%5B6%5D%5Bname%5D=sn&indexes%5B7%5D%5Btype%5D=&indexes%5B7%5D%5Bcolumns%5D%5B%5D=&indexes%5B7%5D%5Blengths%5D%5B%5D=&indexes%5B7%5D%5Bdescs%5D%5B%5D=&indexes%5B7%5D%5Bexpressions%5D%5B%5D=&indexes%5B7%5D%5Bwith%5D=&indexes%5B7%5D%5Bname%5D=`)
		assert.NoError(t, err)
		forms, err := url.ParseQuery(qs)
		assert.NoError(t, err)
	*/
	//echo.Dump(forms)
	forms := url.Values{
		"indexes[0][columns][]":     []string{"id", ""}, //id
		"indexes[0][descs][]":       []string{"ASC", ""},
		"indexes[0][expressions][]": []string{"", ""},
		"indexes[0][lengths][]":     []string{"", ""},
		"indexes[0][name]":          []string{"PRIMARY"},
		"indexes[0][type]":          []string{"PRIMARY"},
		"indexes[0][with]":          []string{""},
		"indexes[1][columns][]":     []string{"customer_id", ""}, //customer_id
		"indexes[1][descs][]":       []string{"ASC", ""},
		"indexes[1][expressions][]": []string{"", ""},
		"indexes[1][lengths][]":     []string{"", ""},
		"indexes[1][name]":          []string{"software_license_customer_id"},
		"indexes[1][type]":          []string{"INDEX"},
		"indexes[1][with]":          []string{""},
		"indexes[2][columns][]":     []string{"package_ident", ""}, //package_ident
		"indexes[2][descs][]":       []string{"ASC", ""},
		"indexes[2][expressions][]": []string{"", ""},
		"indexes[2][lengths][]":     []string{"", ""},
		"indexes[2][name]":          []string{"software_license_package_ident"},
		"indexes[2][type]":          []string{"INDEX"},
		"indexes[2][with]":          []string{""},
		"indexes[3][columns][]":     []string{"product_id", "version", ""}, //product_id+version
		"indexes[3][descs][]":       []string{"ASC", "ASC", ""},
		"indexes[3][expressions][]": []string{"", "", ""},
		"indexes[3][lengths][]":     []string{"", "", ""},
		"indexes[3][name]":          []string{"software_license_product_id_version"},
		"indexes[3][type]":          []string{"INDEX"},
		"indexes[3][with]":          []string{""},
		"indexes[6][columns][]":     []string{"sn", ""}, //sn
		"indexes[6][descs][]":       []string{"ASC", ""},
		"indexes[6][expressions][]": []string{"", ""},
		"indexes[6][lengths][]":     []string{"", ""},
		"indexes[6][name]":          []string{"sn"},
		"indexes[6][type]":          []string{"INDEX"},
		"indexes[6][with]":          []string{""},
		"indexes[7][columns][]":     []string{""},
		"indexes[7][descs][]":       []string{""},
		"indexes[7][expressions][]": []string{""},
		"indexes[7][lengths][]":     []string{""},
		"indexes[7][name]":          []string{""},
		"indexes[7][type]":          []string{""},
		"indexes[7][with]":          []string{""}}
	//t.Logf(`%#v`, forms)
	ctx := defaults.NewMockContext()
	alter, err := diffIndexes(ctx, forms, map[string]*Indexes{}, indexTypesWithFulltext)
	assert.NoError(t, err)
	//t.Log(echo.Dump(alter, false))
	expected := []*indexItems{
		{
			Indexes: &Indexes{
				Name: "PRIMARY",
				Type: "PRIMARY",
				Columns: []string{
					"id",
					"",
				},
				Lengths: []string{
					"",
					"",
				},
				Descs: []string{
					"ASC",
					"",
				},
				Expressions: []string{
					"",
					"",
				},
				With: "",
			},
			Set: []string{
				"`id` ASC",
			},
			Operation: "",
		},
		{
			Indexes: &Indexes{
				Name: "software_license_customer_id",
				Type: "INDEX",
				Columns: []string{
					"customer_id",
					"",
				},
				Lengths: []string{
					"",
					"",
				},
				Descs: []string{
					"ASC",
					"",
				},
				Expressions: []string{
					"",
					"",
				},
				With: "",
			},
			Set: []string{
				"`customer_id` ASC",
			},
			Operation: "",
		},
		{
			Indexes: &Indexes{
				Name: "software_license_package_ident",
				Type: "INDEX",
				Columns: []string{
					"package_ident",
					"",
				},
				Lengths: []string{
					"",
					"",
				},
				Descs: []string{
					"ASC",
					"",
				},
				Expressions: []string{
					"",
					"",
				},
				With: "",
			},
			Set: []string{
				"`package_ident` ASC",
			},
			Operation: "",
		},
		{
			Indexes: &Indexes{
				Name: "software_license_product_id_version",
				Type: "INDEX",
				Columns: []string{
					"product_id",
					"version",
					"",
				},
				Lengths: []string{
					"",
					"",
					"",
				},
				Descs: []string{
					"ASC",
					"ASC",
					"",
				},
				Expressions: []string{
					"",
					"",
					"",
				},
				With: "",
			},
			Set: []string{
				"`product_id` ASC",
				"`version` ASC",
			},
			Operation: "",
		},
		{
			Indexes: &Indexes{
				Name: "sn",
				Type: "INDEX",
				Columns: []string{
					"sn",
					"",
				},
				Lengths: []string{
					"",
					"",
				},
				Descs: []string{
					"ASC",
					"",
				},
				Expressions: []string{
					"",
					"",
				},
				With: "",
			},
			Set: []string{
				"`sn` ASC",
			},
			Operation: "",
		},
	}
	assert.Equal(t, expected, alter)
}
