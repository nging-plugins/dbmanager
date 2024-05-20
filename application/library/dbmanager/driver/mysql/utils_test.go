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

func TestCreateTable(t *testing.T) {
	/*
		qs, err := url.QueryUnescape(`name=shopx_order_service_remark&engine=InnoDB&collation=utf8mb4_general_ci&ai_start_val=&comment=%E5%94%AE%E5%90%8E%E6%9C%8D%E5%8A%A1%E5%A4%87%E6%B3%A8%E4%BF%A1%E6%81%AF&fieldIndexes%5B%5D=1716201973366&fields%5B0%5D%5Bfield%5D=id&fields%5B0%5D%5Borig%5D=&fields%5B0%5D%5Btype%5D=bigint&fields%5B0%5D%5Blength%5D=20&fields%5B0%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B0%5D%5Bunsigned%5D=unsigned&fields%5B0%5D%5Bon_update%5D=&fields%5B0%5D%5Bon_delete%5D=&auto_increment=0&fields%5B0%5D%5Bdefault%5D=&fields%5B0%5D%5Bcomment%5D=ID&fieldIndexes%5B%5D=1716202438538&fields%5B1716202438538%5D%5Bfield%5D=order_id&fields%5B1716202438538%5D%5Borig%5D=&fields%5B1716202438538%5D%5Btype%5D=bigint&fields%5B1716202438538%5D%5Blength%5D=20&fields%5B1716202438538%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B1716202438538%5D%5Bunsigned%5D=unsigned&fields%5B1716202438538%5D%5Bon_update%5D=&fields%5B1716202438538%5D%5Bon_delete%5D=&fields%5B1716202438538%5D%5Bhas_default%5D=1&fields%5B1716202438538%5D%5Bdefault%5D=0&fields%5B1716202438538%5D%5Bcomment%5D=%E8%AE%A2%E5%8D%95ID&fieldIndexes%5B%5D=0&fields%5B1%5D%5Bfield%5D=service_id&fields%5B1%5D%5Borig%5D=&fields%5B1%5D%5Btype%5D=bigint&fields%5B1%5D%5Blength%5D=20&fields%5B1%5D%5Bcollation%5D=&fields%5B1%5D%5Bunsigned%5D=unsigned&fields%5B1%5D%5Bon_update%5D=&fields%5B1%5D%5Bon_delete%5D=&fields%5B1%5D%5Bhas_default%5D=1&fields%5B1%5D%5Bdefault%5D=0&fields%5B1%5D%5Bcomment%5D=%E5%94%AE%E5%90%8E%E6%9C%8D%E5%8A%A1ID&fieldIndexes%5B%5D=1716202175567&fields%5B1716202175567%5D%5Bfield%5D=owner_type&fields%5B1716202175567%5D%5Borig%5D=&fields%5B1716202175567%5D%5Btype%5D=enum&fields%5B1716202175567%5D%5Blength%5D='user'%2C'customer'&fields%5B1716202175567%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B1716202175567%5D%5Bunsigned%5D=&fields%5B1716202175567%5D%5Bon_update%5D=&fields%5B1716202175567%5D%5Bon_delete%5D=&fields%5B1716202175567%5D%5Bhas_default%5D=1&fields%5B1716202175567%5D%5Bdefault%5D=user&fields%5B1716202175567%5D%5Bcomment%5D=%E5%A4%87%E6%B3%A8%E4%BA%BA%E7%B1%BB%E5%9E%8B(user-%E5%90%8E%E5%8F%B0%E7%94%A8%E6%88%B7%3Bcustomer-%E5%89%8D%E5%8F%B0%E7%94%A8%E6%88%B7)&fieldIndexes%5B%5D=1716202174999&fields%5B1716202174999%5D%5Bfield%5D=owner_id&fields%5B1716202174999%5D%5Borig%5D=&fields%5B1716202174999%5D%5Btype%5D=bigint&fields%5B1716202174999%5D%5Blength%5D=&fields%5B1716202174999%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B1716202174999%5D%5Bunsigned%5D=unsigned&fields%5B1716202174999%5D%5Bon_update%5D=&fields%5B1716202174999%5D%5Bon_delete%5D=&fields%5B1716202174999%5D%5Bhas_default%5D=1&fields%5B1716202174999%5D%5Bdefault%5D=0&fields%5B1716202174999%5D%5Bcomment%5D=%E5%A4%87%E6%B3%A8%E4%BA%BAID&fieldIndexes%5B%5D=1716202174211&fields%5B1716202174211%5D%5Bfield%5D=owner_trade_role&fields%5B1716202174211%5D%5Borig%5D=&fields%5B1716202174211%5D%5Btype%5D=enum&fields%5B1716202174211%5D%5Blength%5D='buyer'%2C'seller'%2C'none'&fields%5B1716202174211%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B1716202174211%5D%5Bunsigned%5D=&fields%5B1716202174211%5D%5Bon_update%5D=&fields%5B1716202174211%5D%5Bon_delete%5D=&fields%5B1716202174211%5D%5Bhas_default%5D=1&fields%5B1716202174211%5D%5Bdefault%5D=seller&fields%5B1716202174211%5D%5Bcomment%5D=%E5%A4%87%E6%B3%A8%E4%BA%BA%E4%BA%A4%E6%98%93%E8%A7%92%E8%89%B2(buyer-%E4%B9%B0%E5%AE%B6%3Bseller-%E5%8D%96%E5%AE%B6%3Bnone-%E6%9C%AA%E5%8F%82%E4%B8%8E%E4%BA%A4%E6%98%93%E7%9A%84%E5%85%B6%E4%BB%96%E8%A7%92%E8%89%B2)&fieldIndexes%5B%5D=1716202019330&fields%5B1716202019330%5D%5Bfield%5D=remark&fields%5B1716202019330%5D%5Borig%5D=&fields%5B1716202019330%5D%5Btype%5D=text&fields%5B1716202019330%5D%5Blength%5D=&fields%5B1716202019330%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B1716202019330%5D%5Bunsigned%5D=&fields%5B1716202019330%5D%5Bon_update%5D=&fields%5B1716202019330%5D%5Bon_delete%5D=&fields%5B1716202019330%5D%5Bdefault%5D=&fields%5B1716202019330%5D%5Bcomment%5D=%E5%A4%87%E6%B3%A8&fieldIndexes%5B%5D=1716202357502&fields%5B1716202357502%5D%5Bfield%5D=viewed&fields%5B1716202357502%5D%5Borig%5D=&fields%5B1716202357502%5D%5Btype%5D=int&fields%5B1716202357502%5D%5Blength%5D=&fields%5B1716202357502%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B1716202357502%5D%5Bunsigned%5D=unsigned&fields%5B1716202357502%5D%5Bon_update%5D=&fields%5B1716202357502%5D%5Bon_delete%5D=&fields%5B1716202357502%5D%5Bhas_default%5D=1&fields%5B1716202357502%5D%5Bdefault%5D=0&fields%5B1716202357502%5D%5Bcomment%5D=%E9%A6%96%E6%AC%A1%E6%9F%A5%E7%9C%8B%E6%97%B6%E9%97%B4&fieldIndexes%5B%5D=1716202088169&fields%5B1716202088169%5D%5Bfield%5D=created&fields%5B1716202088169%5D%5Borig%5D=&fields%5B1716202088169%5D%5Btype%5D=int&fields%5B1716202088169%5D%5Blength%5D=&fields%5B1716202088169%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B1716202088169%5D%5Bunsigned%5D=unsigned&fields%5B1716202088169%5D%5Bon_update%5D=&fields%5B1716202088169%5D%5Bon_delete%5D=&fields%5B1716202088169%5D%5Bhas_default%5D=1&fields%5B1716202088169%5D%5Bdefault%5D=0&fields%5B1716202088169%5D%5Bcomment%5D=%E5%88%9B%E5%BB%BA%E6%97%B6%E9%97%B4&fieldIndexes%5B%5D=1716202109068&fields%5B1716202109068%5D%5Bfield%5D=updated&fields%5B1716202109068%5D%5Borig%5D=&fields%5B1716202109068%5D%5Btype%5D=int&fields%5B1716202109068%5D%5Blength%5D=&fields%5B1716202109068%5D%5Bcollation%5D=utf8mb4_general_ci&fields%5B1716202109068%5D%5Bunsigned%5D=unsigned&fields%5B1716202109068%5D%5Bon_update%5D=&fields%5B1716202109068%5D%5Bon_delete%5D=&fields%5B1716202109068%5D%5Bhas_default%5D=1&fields%5B1716202109068%5D%5Bdefault%5D=0&fields%5B1716202109068%5D%5Bcomment%5D=%E4%BF%AE%E6%94%B9%E6%97%B6%E9%97%B4&partition_method=&partition_expression=&partition_position=&partition_names%5B%5D=&partition_values%5B%5D=`)
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
