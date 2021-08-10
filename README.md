# tableToStruct

将连接数据库中的表自动创建为结构体

### 引入包使用

```go

package main

import (
 "tableToStruct"
)
func main() {
 t2s := tableToStruct.NewTableToStruct()
 t2s.SetDbDsn(dsn) //设置 数据库dsn name:pwd@tcp(dbserver)/dbname?charset=utf8"
 t2s.SetSavePath(filePath) //设置保存的文件
 t2s.SetTable(tableList) //设置要导出的表  不设置时 导出数据库全部的表
 t2s.Run()
}
```

### 直接编译成执行文件使用

| 参数 | 说明                                 |
| ---- | ------------------------------------ |
| -dsn | 数据库连接配置                       |
| -f   | 设置生成的结构体保存的文件路径       |
| -t   | 设置需要导出的表（未设置时导出全部,设置格式 table1,table2） |


``` go
package main

import (
 "flag"
 "fmt"
 "os"
 "strings"
 "tableToStruct"
)
//-f --file  文件路径或文件名
//-dsn  name:pwd@dbserver/dbname?charset=utf8
//-t --table  a,b,c
func main() {
 var filePath string
 var dsn string
 var table string
 path, _ := os.Getwd()
 flag.StringVar(&filePath, "f", path+"/mode.go", path+"/mode.go")
 flag.StringVar(&dsn, "dsn", "", "name:pwd@dbserver/dbname?charset=utf8")
 flag.StringVar(&table, "t", "", "table1,table2")
 flag.Parse()
 var tableList []string
 if table != "" {
  tableList = strings.Split(table, ",")
 }
 tag := strings.Index(dsn, "@")
 tag1 := strings.Index(dsn, "/")
 dsnUP := dsn[0:tag]
 dsnTcp := dsn[tag+1 : tag1]
 dsnLast := dsn[tag1:]
 dsntr := fmt.Sprintf("%s@tcp(%s)%s", dsnUP, dsnTcp, dsnLast)
 t2s := tableToStruct.NewTableToStruct()
 t2s.SetDbDsn(dsntr)
 t2s.SetSavePath(filePath)
 t2s.SetTable(tableList)
 t2s.Run()
}
```

### 可执行文件执行  `./convert.bin -dsn username:pwd@server/dbname?charset=utf8  -f dir  -t table1,table2`