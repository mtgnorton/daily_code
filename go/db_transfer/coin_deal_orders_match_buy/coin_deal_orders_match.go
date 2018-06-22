package main

import (
	"database/sql"
	"fmt"
	"sync"

	"bilian/common"

	"github.com/cihub/seelog"

	_ "github.com/go-sql-driver/mysql"
)

type Match struct {
	common.Deliver
}

func (d *Match) Init() {
	db, err := sql.Open("mysql", "wwwbtcblcom:39p2AtWeJNvoXkICDXAy@tcp(47.74.159.140:3306)/wwwbtcblcom")
	if err != nil {
		seelog.Errorf("打开数据库出错%v", err)
	}

	d.ExportDb = db

	db, err = sql.Open("mysql", "ceshi1btcblcom:hFKMObz5WYcT1BgZ0BPE@tcp(47.74.159.140:3306)/ceshi1btcblcom")

	if err != nil {
		seelog.Errorf("打开数据库出错%v", err)
	}

	d.ImportDb = db

	d.ExportTable = "dms_货币交易交易"
	d.ImportTable = "coin_deal_orders_match_bak"

	var a, b, c, h, e, f, g, k, l, m, n []byte
	d.ExportField = map[string]*[]byte{
		"买入委托":    &a,
		"买入编号":    &b,
		"卖出编号":    &c,
		"卖出委托":    &k,
		"价格":      &h,
		"数量":      &e,
		"买入手续费金额": &f,
		"交易时间":    &g,
		"账户":      &l,
		"货币类型":    &m,
		"卖出手续费金额": &n,
	}

	d.InitDefaultData()
	d.GoroutineNumber = 5000.00
	d.ReadyChan = make(chan struct{})
	d.Duration = 1000 //每两秒开启一个进程

	d.AttachChans = make(map[string]chan struct{})
	d.AttachInfo = make(map[string]*sync.Map)

	//d.Test = true
	//d.Predict = true
	d.AttachChans["user"] = make(chan struct{})

	d.AttachInfo["user"] = &sync.Map{}

	d.AttachChans["symbols"] = make(chan struct{})

	d.AttachInfo["symbols"] = &sync.Map{}
}

func (d *Match) InitDefaultData() {

	d.RowDefaultData = map[string]common.RowFunc{
		"symbol_id": func(row map[string]*[]byte) string {

			name := string(*row["账户"]) + string(*row["货币类型"])
			symbolId, ok := d.AttachInfo["symbols"].Load(name)
			if !ok {
				seelog.Errorf("编号未找到，symbols为%s", name)
				d.NoInsertCount++
				return "continue"
			}
			return fmt.Sprintf("%q", symbolId)
		},
		"order_id": common.Def("买入委托", "no"),
		"order_uid": func(row map[string]*[]byte) string {

			userid, ok := d.AttachInfo["user"].Load(string(*row["买入编号"]))
			if !ok {
				seelog.Errorf("编号未找到，username为%s", *row["买入编号"])
				d.NoInsertCount++
				return "continue"
			}
			return fmt.Sprintf("%q", userid)
		},

		"match_id": common.Def("卖出委托", "no"),
		"match_uid": func(row map[string]*[]byte) string {
			userid, ok := d.AttachInfo["user"].Load(string(*row["卖出编号"]))
			if !ok {
				d.NoInsertCount++
				seelog.Errorf("编号未找到，username为%s", *row["卖出编号"])
				return "continue"

			}
			return fmt.Sprintf("%q", userid)
		},
		"price":         common.Def("价格", "no"),
		"filled_amount": common.Def("数量", "no"),
		"filled_fees":   common.Def("买入手续费金额", "no"),
		"create_time":   common.Def("交易时间", "no"),
		"status":        common.Def("", "1"),
		"direction":     common.Def("", "0"),
	}
	d.InitFinish()

}

func (d *Match) GetUserInfo() {
	users, err := d.ExportDb.Query("select `id`,`编号` from dms_会员")
	if err != nil {
		seelog.Errorf("读取会员出错%v", err)
	}
	var count int
	for users.Next() {
		var username string
		var id []byte
		if err := users.Scan(&id, &username); err != nil {
			seelog.Errorf("读取数据出错dms_会员 %v", err)
		}

		count++
		d.AttachInfo["user"].Store(username, string(id))
	}
	seelog.Infof("用户总量为：%v", count)
	close(d.AttachChans["user"])
}

func (d *Match) GetCoin() { //用于coin_address
	users, err := d.ImportDb.Query("select `symbol`,`id` from coin")
	if err != nil {
		seelog.Errorf("读取coin出错%v", err)
	}
	var count int
	for users.Next() {
		var symbol string
		var id []byte
		if err := users.Scan(&symbol, &id); err != nil {
			seelog.Errorf("读取数据出错coin %v", err)
		}

		count++
		d.AttachInfo["coin"].Store(symbol, string(id))
	}
	seelog.Infof("coin总量为：%v", count)
	close(d.AttachChans["coin"])
}
func (d *Match) GetSymbolInfo() {
	users, err := d.ImportDb.Query("select `id`,`name` from coin_symbol")
	if err != nil {
		seelog.Errorf("读取coin_symbol出错%v", err)
	}
	var count int
	for users.Next() {
		var name string
		var id []byte
		if err := users.Scan(&id, &name); err != nil {
			seelog.Errorf("读取coin_symbol数据出错 %v", err)
		}

		count++
		d.AttachInfo["symbols"].Store(name, string(id))
	}
	seelog.Infof("symbol总量为：%v", count)
	close(d.AttachChans["symbols"])
}
func main() {
	common.SetLogger("logConfig.xml")
	defer func() {
		seelog.Flush()
	}()

	seelog.Info("开始插入")
	d := Match{}
	d.Init()
	go d.GetUserInfo()
	go d.GetSymbolInfo()
	go d.Export()
	d.CalCon()
	seelog.Infof("实际插入的记录总量为：%v", d.GetActualCount())
	seelog.Info("插入结束")
}
