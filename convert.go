package tableToStruct

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

const selectSql = `SELECT COLUMN_NAME,DATA_TYPE,TABLE_NAME,COLUMN_COMMENT
FROM information_schema.COLUMNS 
WHERE table_schema = DATABASE()`

type Record struct {
	COLUMN_NAME    string
	DATA_TYPE      string
	TABLE_NAME     string
	COLUMN_COMMENT string
	IS_NULLABLE    string
}
type TableToStruct struct {
	savePath      string   //生成文件保存路径
	tables        []string //设置导入哪些表 空时为数据库全部的表
	db            *sql.DB
	dsn           string //数据库连接配置
	selectSql     string
	structContext chan string
	mux           *sync.Mutex
	wg            sync.WaitGroup
}

func NewTableToStruct() *TableToStruct {
	t2s := &TableToStruct{}
	t2s.db = &sql.DB{}
	t2s.selectSql = selectSql
	t2s.wg = sync.WaitGroup{}
	t2s.mux = &sync.Mutex{}
	t2s.structContext = make(chan string)
	return t2s
}

//设置保存路径
func (t *TableToStruct) SetSavePath(path string) {
	t.savePath = path
}

func (t *TableToStruct) SetDbDsn(dsn string) {
	t.dsn = dsn
	var err error
	t.db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln("数据库连接失败:" + err.Error())
	}
	log.Println("数据库连接成功")
}

//设置要查询哪些表
func (t *TableToStruct) SetTable(tables []string) {
	t.tables = tables
	var whereParams string
	switch len(tables) {
	case 0:
	case 1:
		whereParams = fmt.Sprintf(" AND TABLE_NAME = '%s' ", tables[0])
	default:
		var ts []string
		for _, t := range tables {
			ts = append(ts, fmt.Sprintf("'%s'", strings.TrimSpace(t)))
		}
		whereParams = fmt.Sprintf(" AND TABLE_NAME IN (%s) ", strings.Join(ts, ","))
	}
	t.selectSql += whereParams
}

//获取表记录信息
func (t *TableToStruct) getTablesColumns() map[string][]Record {
	tablesColumns := make(map[string][]Record)
	rows, err := t.db.Query(t.selectSql)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		record := Record{}
		err = rows.Scan(&record.COLUMN_NAME, &record.DATA_TYPE, &record.TABLE_NAME, &record.COLUMN_COMMENT)
		if err != nil {
			log.Fatalln(err.Error())
		}
		tablesColumns[record.TABLE_NAME] = append(tablesColumns[record.TABLE_NAME], record)
	}
	return tablesColumns
}

//导出操作
func (t *TableToStruct) exportStructText(tc chan *Table) {
	var structContext string
	var importContext string
	for {
		select {
		case table := <-tc:
			if table == nil {
				goto Loop
			}
			t.mux.Lock()
			res := table.handler()
			if !res {
				log.Printf("数据表%s处理失败\n", table.Name)
			}
			if len(table.Tstruct) > 0 {
				structContext += "\n" + table.Tstruct
			}
			if len(table.ImportTag) > 0 && (strings.Index(importContext, table.ImportTag) == -1) {
				importContext += table.ImportTag + "\n"
			}
			t.mux.Unlock()
		default:
		}
	}
Loop:
	t.wg.Done()
	t.structContext <- fmt.Sprintf("%s \n %s \n\n %s", "package mode", importContext, structContext)
	close(t.structContext)
}

//保存内容
func (t *TableToStruct) saveContext(text string) bool {
	if len(t.savePath) == 0 {
		t.savePath = "tableStruct.go"
	}
	log.Println("写入文:" + t.savePath)
	filePath := t.savePath
	f, err := os.Create(filePath)
	if err != nil {
		log.Println("写入文件失败:" + err.Error())
		return false
	}
	defer f.Close()
	f.WriteString(text)
	cmd := exec.Command("gofmt", "-w", filePath)
	cmd.Run()
	log.Printf("数据保存完成")
	return true
}

func (t *TableToStruct) Run() {
	tablesColumns := t.getTablesColumns()
	tableChan := make(chan *Table, len(tablesColumns))
	for _tablename, _column := range tablesColumns {
		table := &Table{Name: _tablename, Columns: _column}
		tableChan <- table
	}
	t.wg.Add(1)
	close(tableChan)
	go t.exportStructText(tableChan)
	t.wg.Wait()
	context := <-t.structContext
	if len(context) == 0 {
		log.Println("无表内容需导出")
		return
	}
	s := t.saveContext(context)
	if s {
		log.Println("导出完成")
	}
}
