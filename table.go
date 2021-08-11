package tableToStruct

import (
	"fmt"
	"strings"
)

var typeForMysqlToGo = map[string]string{
	"int":                "int",
	"integer":            "int",
	"tinyint":            "int",
	"smallint":           "int",
	"mediumint":          "int",
	"bigint":             "int64",
	"int unsigned":       "int",
	"integer unsigned":   "int",
	"tinyint unsigned":   "int",
	"smallint unsigned":  "int",
	"mediumint unsigned": "int",
	"bigint unsigned":    "int64",
	"bit":                "int",
	"bool":               "bool",
	"enum":               "string",
	"set":                "string",
	"varchar":            "string",
	"char":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"text":               "string",
	"longtext":           "string",
	"blob":               "string",
	"tinyblob":           "string",
	"mediumblob":         "string",
	"longblob":           "string",
	"date":               "time.Time", // time.Time or string
	"datetime":           "time.Time", // time.Time or string
	"timestamp":          "time.Time", // time.Time or string
	"time":               "time.Time", // time.Time or string
	"float":              "float64",
	"double":             "float64",
	"decimal":            "float64",
	"binary":             "string",
	"varbinary":          "string",
}

type Table struct {
	Columns   []Record
	Name      string
	Tstruct   string //结构体
	ImportTag string //导入包
}

//处理表数据
func (t *Table) handler() bool {
	var context string
	tableName := t.Name
	// // l := len(tableName)
	// // if l > 1 {
	// // 	tableName = strings.ToUpper(tableName[0:1]) + tableName[1:]
	// // } else if l == 1 {
	// // 	tableName = strings.ToUpper(tableName[0:1])
	// // }
	l := len(tableName)
	if l > 0 {
		tableNameOptions := t.columns(t.Name)
		tableName = strings.Join(tableNameOptions, "_")
	}

	context = fmt.Sprintf(" type %s struct {\n ", tableName)
	for _, c := range t.Columns {
		if c.TABLE_NAME != t.Name {
			continue
		}
		list := t.columns(c.COLUMN_NAME)
		field := strings.Join(list, "")
		fieldType := typeForMysqlToGo[c.DATA_TYPE]
		var comment string
		if len(c.COLUMN_COMMENT) != 0 {
			comment = strings.ReplaceAll(c.COLUMN_COMMENT, "\n", "")
			comment = fmt.Sprintf(" // %s", comment)
		}
		context += fmt.Sprintf("%s %s  %s \n", field, fieldType, comment)
	}
	context += "}\n\n"
	var importContent string
	if strings.Contains(context, "time.Time") {
		importContent = "import \"time\"\n\n"
	}
	t.Tstruct = context
	t.ImportTag = importContent
	return true
}

//字段名处理
func (t *Table) columns(c string) []string {
	var textList []string
	for _, f := range strings.Split(c, "_") {
		switch len(f) {
		case 0:
		case 1:
			//text += strings.ToUpper(f[0:1])
			textList = append(textList, strings.ToUpper(f[0:1]))
		default:
			//text += strings.ToUpper(f[0:1]) + f[1:]
			textList = append(textList, strings.ToUpper(f[0:1])+f[1:])

		}
	}
	return textList
}
