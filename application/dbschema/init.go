// @generated Do not edit a file, which is automatically generated by the generator.

package dbschema

import (
	"github.com/webx-top/db/lib/factory"
)

var WithPrefix = func(tableName string) string {
	return "" + tableName
}

var DBI = factory.DefaultDBI

func init() {

	DBI.FieldsRegister(map[string]map[string]*factory.FieldInfo{"nging_db_account": {"created": {Name: "created", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "创建时间", GoType: "uint", MyType: "", GoName: "Created"}, "engine": {Name: "engine", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 30, Options: []string{}, DefaultValue: "mysql", Comment: "数据库引擎", GoType: "string", MyType: "", GoName: "Engine"}, "host": {Name: "host", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 200, Options: []string{}, DefaultValue: "localhost:3306", Comment: "服务器地址", GoType: "string", MyType: "", GoName: "Host"}, "id": {Name: "id", DataType: "int", Unsigned: true, PrimaryKey: true, AutoIncrement: true, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "ID", GoType: "uint", MyType: "", GoName: "Id"}, "name": {Name: "name", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 120, Options: []string{}, DefaultValue: "", Comment: "数据库名称", GoType: "string", MyType: "", GoName: "Name"}, "options": {Name: "options", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "其它选项(JSON)", GoType: "string", MyType: "", GoName: "Options"}, "password": {Name: "password", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 128, Options: []string{}, DefaultValue: "", Comment: "密码", GoType: "string", MyType: "", GoName: "Password"}, "title": {Name: "title", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 120, Options: []string{}, DefaultValue: "", Comment: "标题", GoType: "string", MyType: "", GoName: "Title"}, "uid": {Name: "uid", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "UID", GoType: "uint", MyType: "", GoName: "Uid"}, "updated": {Name: "updated", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "修改时间", GoType: "uint", MyType: "", GoName: "Updated"}, "user": {Name: "user", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 100, Options: []string{}, DefaultValue: "", Comment: "用户名", GoType: "string", MyType: "", GoName: "User"}}, "nging_db_sync": {"alter_ignore": {Name: "alter_ignore", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "要忽略的列、索引、外键", GoType: "string", MyType: "", GoName: "AlterIgnore"}, "created": {Name: "created", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "创建时间", GoType: "uint", MyType: "", GoName: "Created"}, "destination_account_id": {Name: "destination_account_id", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "目标数据库账号ID", GoType: "uint", MyType: "", GoName: "DestinationAccountId"}, "drop": {Name: "drop", DataType: "tinyint", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "删除待同步数据库中多余的字段、索引、外键 ", GoType: "uint", MyType: "", GoName: "Drop"}, "dsn_destination": {Name: "dsn_destination", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 255, Options: []string{}, DefaultValue: "", Comment: "目标数据库", GoType: "string", MyType: "", GoName: "DsnDestination"}, "dsn_source": {Name: "dsn_source", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 255, Options: []string{}, DefaultValue: "", Comment: "同步源", GoType: "string", MyType: "", GoName: "DsnSource"}, "id": {Name: "id", DataType: "int", Unsigned: true, PrimaryKey: true, AutoIncrement: true, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "ID", GoType: "uint", MyType: "", GoName: "Id"}, "mail_to": {Name: "mail_to", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 200, Options: []string{}, DefaultValue: "", Comment: "发送邮件", GoType: "string", MyType: "", GoName: "MailTo"}, "name": {Name: "name", DataType: "varchar", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 120, Options: []string{}, DefaultValue: "", Comment: "方案名", GoType: "string", MyType: "", GoName: "Name"}, "skip_tables": {Name: "skip_tables", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "要跳过的表", GoType: "string", MyType: "", GoName: "SkipTables"}, "source_account_id": {Name: "source_account_id", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "源数据库账号ID", GoType: "uint", MyType: "", GoName: "SourceAccountId"}, "tables": {Name: "tables", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "要同步的表", GoType: "string", MyType: "", GoName: "Tables"}, "updated": {Name: "updated", DataType: "int", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: -0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "更新时间", GoType: "int", MyType: "", GoName: "Updated"}}, "nging_db_sync_log": {"change_table_num": {Name: "change_table_num", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "被更改的表的数量", GoType: "uint", MyType: "", GoName: "ChangeTableNum"}, "change_tables": {Name: "change_tables", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "被更改的表", GoType: "string", MyType: "", GoName: "ChangeTables"}, "created": {Name: "created", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "创建时间", GoType: "uint", MyType: "", GoName: "Created"}, "elapsed": {Name: "elapsed", DataType: "bigint", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "总共耗时", GoType: "uint64", MyType: "", GoName: "Elapsed"}, "failed": {Name: "failed", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "0", Comment: "失败次数", GoType: "uint", MyType: "", GoName: "Failed"}, "id": {Name: "id", DataType: "bigint", Unsigned: true, PrimaryKey: true, AutoIncrement: true, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "", GoType: "uint64", MyType: "", GoName: "Id"}, "result": {Name: "result", DataType: "text", Unsigned: false, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "结果报表", GoType: "string", MyType: "", GoName: "Result"}, "sync_id": {Name: "sync_id", DataType: "int", Unsigned: true, PrimaryKey: false, AutoIncrement: false, Min: 0, Max: 0, Precision: 0, MaxSize: 0, Options: []string{}, DefaultValue: "", Comment: "同步方案ID", GoType: "uint", MyType: "", GoName: "SyncId"}}})

	DBI.ColumnsRegister(map[string][]string{"nging_db_account": {"id", "title", "uid", "engine", "host", "user", "password", "name", "options", "created", "updated"}, "nging_db_sync": {"id", "name", "source_account_id", "dsn_source", "destination_account_id", "dsn_destination", "tables", "skip_tables", "alter_ignore", "drop", "mail_to", "created", "updated"}, "nging_db_sync_log": {"id", "sync_id", "created", "result", "change_tables", "change_table_num", "elapsed", "failed"}})

	DBI.ModelsRegister(factory.ModelInstancers{`NgingDbAccount`: factory.NewMI("nging_db_account", func(connID int) factory.Model { return &NgingDbAccount{base: *((&factory.Base{}).SetConnID(connID))} }, "数据库账号"), `NgingDbSync`: factory.NewMI("nging_db_sync", func(connID int) factory.Model { return &NgingDbSync{base: *((&factory.Base{}).SetConnID(connID))} }, "数据表同步方案"), `NgingDbSyncLog`: factory.NewMI("nging_db_sync_log", func(connID int) factory.Model { return &NgingDbSyncLog{base: *((&factory.Base{}).SetConnID(connID))} }, "数据表同步日志")})

}
