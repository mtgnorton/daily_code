package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/cihub/seelog"
)

type Deliver struct {
	ImportDb    *sql.DB
	ExportDb    *sql.DB
	ExportTable string
	ImportTable string
	ExportField map[string]*[]byte
	ImportField []string

	RowDefaultData map[string]RowFunc //新表中一行数据各个字段对应的函数处理

	ImportAllData   []map[string]string //查询出的所有结果
	GoroutineNumber float64             //每个goroutine导入的数量

	ReadyChan chan struct{} //是否准备进行插入操作

	NoInsertCount int
	ActualInsert  int64
	Duration      time.Duration //每个进程的时间间隔 ,启用processChan后无效
	Test          bool          //进行100行的数据插入测试
	Predict       bool          //预处理，不进行实际插入

	AttachInfo  map[string]*sync.Map     //附加的数据，如插入时依赖某个表
	AttachChans map[string]chan struct{} //附加的数据是否完成

	ProcessChan chan struct{} //两个进程可以同时插入,使用channel控制插入频率
}

type Default func(string, string) RowFunc
type RowFunc func(map[string]*[]byte) string

//将导出数据库的数据进行导出，并开始将老数据赋值给相应的新表字段

func (d *Deliver) InitFinish() {
	d.ProcessChan = make(chan struct{}, 2)
	d.ImportField = make([]string, 0, len(d.RowDefaultData))

	for field, _ := range d.RowDefaultData {
		d.ImportField = append(d.ImportField, field)
	}
}
func (d *Deliver) Export() {

	sqlQuery := "select"
	var ExportFieldSlice = make([]interface{}, 0, len(d.ExportField))
	for key, _ := range d.ExportField {
		sqlQuery += (" `" + key + "` ,")
		ExportFieldSlice = append(ExportFieldSlice, d.ExportField[key])
	}
	sqlQueryByte := []byte(sqlQuery)

	sqlQuery = string(sqlQueryByte[:len(sqlQueryByte)-1])

	if d.Test {
		sqlQuery += ("from " + d.ExportTable + " limit 100")
	} else {
		sqlQuery += ("from " + d.ExportTable)

	}

	seelog.Info(sqlQuery)
	exportRs, err := d.ExportDb.Query(sqlQuery)
	if err != nil {
		seelog.Errorf("%v查询出错%v", d.ExportTable, err)
	}
	//待前置操作完成
	for _, item := range d.AttachChans {
		<-item
	}

	//开始进行赋值操作
	for exportRs.Next() {

		if err := exportRs.Scan(ExportFieldSlice...); err != nil {
			seelog.Errorf("%v读取数据出错%v", d.ExportTable, err)
		}
		rowData := make(map[string]string)

		for _, key := range d.ImportField {

			rowData[key] = (d.RowDefaultData[key])(d.ExportField)
			//如果返回continue，将跳过该条记录
			if rowData[key] == "continue" {
				goto stop
			}
		}

		d.ImportAllData = append(d.ImportAllData, rowData)
	stop:
	}
	exportRs.Close()
	if err = exportRs.Err(); err != nil {
		seelog.Errorf("遍历数据库数据时出错")
	}
	close(d.ReadyChan)

}

//赋值给新表字段时的处理
func Def(field string, defaultValue string) RowFunc {
	return func(row map[string]*[]byte) string {
		if defaultValue != "no" {
			return fmt.Sprintf("%q", defaultValue)
		}
		return fmt.Sprintf("%q", *row[field])
	}

}

//开始进行插入预处理，判断开启的进程个数
func (d *Deliver) CalCon() {

	GoroutineNumber := int(d.GoroutineNumber)

	<-d.ReadyChan

	goNumber := math.Ceil(float64(len(d.ImportAllData)) / d.GoroutineNumber)

	theoryTime := float64(d.Duration) / 1000 * goNumber

	seelog.Infof("开启进程数量为：%v", goNumber)
	seelog.Infof("要插入的记录数量为：%v", len(d.ImportAllData))
	seelog.Infof("不完善的记录数量为：%v", d.NoInsertCount)
	seelog.Infof("预估理论时间为：%v秒", theoryTime)

	beginTime := time.Now().Unix()

	var wg sync.WaitGroup

	//go func() {
	//	time.Sleep(time.Second)
	//	d.ProcessChan <- struct{}{}
	//}()

	for i := 0; i < int(goNumber); i++ {
		start := i * GoroutineNumber

		end := (i + 1) * GoroutineNumber

		var tempSlice []map[string]string

		if i == int(goNumber)-1 {

			tempSlice = d.ImportAllData[start:]
		} else {
			tempSlice = d.ImportAllData[start:end]
		}

		wg.Add(1)

		//time.Sleep(d.Duration * time.Millisecond)

		d.ProcessChan <- struct{}{}

		go func(i int) {
			//_ = tempSlice
			seelog.Infof("进程%v开启", i+1)

			if !d.Predict {
				d.Insert(tempSlice, start, end)
			}

			defer wg.Done()
		}(i)

	}

	wg.Wait()
	finishTime := time.Now().Unix()
	seelog.Infof("实际消耗时间为：%v秒", finishTime-beginTime)
}

//进行插入
func (d *Deliver) Insert(transfersSlice []map[string]string, start int, end int) {
	var buffer bytes.Buffer
	buffer.WriteString("INSERT INTO " + d.ImportTable + " (")

	for _, field := range d.ImportField {
		buffer.WriteString(" " + field + ",")
	}
	buffer.Truncate(buffer.Len() - 1)

	buffer.WriteString(")VALUES ")

	for _, item := range transfersSlice {
		buffer.WriteString("(")
		for _, field := range d.ImportField {
			buffer.WriteString(item[field])
			buffer.WriteString(",")
		}
		buffer.Truncate(buffer.Len() - 1)

		buffer.WriteString("),")
	}
	buffer.Truncate(buffer.Len() - 1)

	if d.Test {
		seelog.Info(buffer.String())
	}

	rs, err := d.ImportDb.Exec(buffer.String())

	if err != nil {
		seelog.Errorf("插入数据出错，出错索引为%v-%v,出错原因%v", start, end, err)
	}

	lines, err := rs.RowsAffected()

	if err != nil {
		seelog.Errorf("获取插入行数出错%v", err)
	}
	d.ActualInsert += lines
	seelog.Infof("插入的行数为%v", lines)
	<-d.ProcessChan
}

func (d *Deliver) GetActualCount() int64 {
	return d.ActualInsert
}

func SetLogger(fileName string) {
	if _, err := os.Stat(fileName); err == nil {
		logger, err := seelog.LoggerFromConfigAsFile(fileName)
		if err != nil {
			panic(err)
		}

		seelog.ReplaceLogger(logger)
	} else {
		configString := `<seelog>
                        <outputs formatid="main">
                            <filter levels="info,error,critical">
                                <rollingfile type="date" filename="log/AppLog.log" namemode="prefix" datepattern="02.01.2006"/>
                            </filter>
                            <console/>
                        </outputs>
                        <formats>
                            <format id="main" format="%Date %Time [%LEVEL] %Msg%n"/>
                        </formats>
                        </seelog>`
		logger, err := seelog.LoggerFromConfigAsString(configString)
		if err != nil {
			panic(err)
		}

		seelog.ReplaceLogger(logger)
	}

}
