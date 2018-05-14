package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/leizongmin/huobiapi"

	"strings"
	"sync"
	"time"
)

const HuoBiMarketUrl = "https://api.huobipro.com/v1/common/symbols"

type coinItem struct {
	Base  string `json:"base-currency"`
	Quote string `json:"quote-currency"`
}
type coinAll struct {
	Status string
	List   []coinItem `json:"data"`
}

type QuotationItem struct {
	Close string
	Cny   string
	Vary  string
	Name  string
	Base  string
	Quote string
}

type Quotation struct {
	result       []QuotationItem //存储结果
	lastResult   []QuotationItem
	mu           sync.RWMutex
	rsExist      map[string]bool    //结果是否已经存在
	c1           chan QuotationItem //将结果通过c1传输到result
	resultJson   string
	updateTime   int64
	symbolNumber int64
	priceMap     map[string]float64
}

func initHuoBiQuotation() (q *Quotation, s []coinItem) {

	symbols := getHuoBiSymbols()

	quotation := Quotation{

		rsExist: make(map[string]bool),

		c1: make(chan QuotationItem, len(symbols)),

		resultJson:   `{"message":"初始化中，请稍后"}`,
		updateTime:   0,
		symbolNumber: int64(len(symbols)),
		priceMap:     make(map[string]float64),
	}

	return &quotation, symbols
}

func (quotation *Quotation) huoBiConnect(symbos []coinItem) {

	dollarRate := GetRate()

	market, err := huobiapi.NewMarket("wss://api.huobi.pro/ws")

	go func() {
		for _, query := range []string{"btcusdt", "ethusdt"} {

			market.Subscribe("market."+query+".kline.1min", func(topic string, data *huobiapi.JSON) {
				open := data.Get("tick").Get("open").MustFloat64()
				area := strings.Replace(topic, "market.", "", -1)
				area = strings.Replace(area, "usdt.kline.1min", "", -1)

				quotation.mu.Lock()
				quotation.priceMap[area] = open
				quotation.mu.Unlock()
			})

		}
	}()

	time.Sleep(time.Second * 1)

	if err != nil {
		fmt.Println("初始化market出错")
	}

	for _, item := range symbos {

		go func(item coinItem) {
			market.Subscribe("market."+item.Base+item.Quote+".kline.1day", func(topic string, data *huobiapi.JSON) {

				open := data.Get("tick").Get("open").MustFloat64()

				closeTemp := data.Get("tick").Get("close").MustFloat64()

				vary := fmt.Sprintf("%.2f", (closeTemp-open)/open*100)

				topic = strings.Replace(topic, "market.", "", -1)

				topic = strings.Replace(topic, "kline.1day", "", -1)

				var tempRate float64 = 1
				quotation.mu.Lock()
				if item.Quote == "btc" {
					tempRate = quotation.priceMap["btc"]
				}
				if item.Quote == "eth" {
					tempRate = quotation.priceMap["eth"]
				}
				quotation.mu.Unlock()
				cny := fmt.Sprintf("%.6f", closeTemp*tempRate*dollarRate)

				closePrice := fmt.Sprintf("%.6f", closeTemp)

				quotation.mu.Lock()

				if !quotation.rsExist[topic] {
					quotation.rsExist[topic] = true
					quotation.c1 <- QuotationItem{closePrice, cny, vary, topic, item.Base, item.Quote}
				}
				quotation.mu.Unlock()
			})
		}(item)

	}
	go quotation.collectResult(30, 1, 20)

}

func (quotation *Quotation) collectResult(ignoreNumberOne int64, ignoreNumberTwo int64, collectTime int64) {

	var i int64 = 0
	for item := range quotation.c1 {

		quotation.result = append(quotation.result, item)
		i++

		if (quotation.symbolNumber-ignoreNumberOne <= i && time.Now().Unix()-quotation.updateTime > collectTime) || (quotation.symbolNumber-1 <= i && time.Now().Unix()-quotation.updateTime < collectTime) {
			i = 0
			jsonData, err := json.Marshal(quotation.result)
			fmt.Println(ignoreNumberOne, len(quotation.result), time.Now())
			if err != nil {
				fmt.Println("结果转换成json出错")
			}
			quotation.mu.Lock()
			quotation.lastResult = quotation.result
			quotation.result = nil
			quotation.updateTime = time.Now().Unix()
			quotation.resultJson = fmt.Sprintf("%s", jsonData)
			quotation.rsExist = make(map[string]bool)
			quotation.mu.Unlock()
		}
	}
}

func (quotation *Quotation) getResultJson() string {
	return quotation.resultJson
}

func (quotation *Quotation) searchCoin(coinName string, exchange string) []QuotationItem {
	if len(quotation.lastResult) == 0 {
		return []QuotationItem{}
	}
	var searchRs []QuotationItem
	for _, item := range quotation.lastResult {
		if strings.ToLower(item.Base) == strings.ToLower(coinName) {
			item.Name = exchange
			searchRs = append(searchRs, item)
		}
	}
	return searchRs
}

func (quotation *Quotation) searchManyCoin(coinName string, exchange string) []QuotationItem {
	coins := strings.Split(strings.ToLower(coinName), ",")

	if len(quotation.lastResult) == 0 {
		return nil
	}

	var searchRs []QuotationItem
	for _, item := range quotation.lastResult {

		for _, search := range coins {

			if strings.ToLower(item.Base) == strings.ToLower(search) {
				item.Name = exchange
				searchRs = append(searchRs, item)
			}
		}

	}
	return searchRs
}

func transferJson(data []QuotationItem) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("transferJson出错", err)

	}
	return fmt.Sprintf("%s", jsonData)

}

func getHuoBiSymbols() []coinItem {
	resp, err := http.Get(HuoBiMarketUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求出错: %v\n", err)
		os.Exit(1)
	}
	symbolsByte, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "读取内容出错 %v\n", err)
		os.Exit(1)
	}

	var ca coinAll
	err = json.Unmarshal(symbolsByte, &ca)
	if err != nil {
		fmt.Println("解码出错", err)
	}
	return ca.List

}
