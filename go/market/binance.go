package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"strings"
	"time"
)

const BiMarketUrl = "https://api.binance.com/api/v1/exchangeInfo"

type BiItem struct {
	Base  string `json:"baseAsset"`
	Quote string `json:"quoteAsset"`
}
type BiSymbos struct {
	List []BiItem `json:"symbols"`
}

func initBiQuotation() (q *Quotation, s []BiItem) {

	symbols := getBiSymbos()

	quotation := Quotation{

		//result: make([]QuotationItem,),

		rsExist: make(map[string]bool),

		c1: make(chan QuotationItem, len(symbols)),

		resultJson:   `{"message":"初始化中，请稍后"}`,
		updateTime:   0,
		symbolNumber: int64(len(symbols)),
		priceMap:     make(map[string]float64),
	}

	return &quotation, symbols
}
func (quotation *Quotation) BiConnect(symbos []BiItem) {

	dollarRate := GetRate()

	go func() {
		MinuteCon([]string{"bnbusdt", "btcusdt", "ethusdt"}, func(event *KlineItem, item BiItem) {
			open, _ := strconv.ParseFloat(event.Kline.Open, 64)

			quotation.mu.Lock()
			quotation.priceMap[strings.ToLower(strings.Replace(event.Symbol, "USDT", "", -1))] = open

			quotation.mu.Unlock()
		})
	}()

	time.Sleep(time.Second * 1)

	biLibCon(symbos, func(event *KlineItem, item BiItem) {

		closeTemp, _ := strconv.ParseFloat(event.Kline.Close, 64)
		open, _ := strconv.ParseFloat(event.Kline.Open, 64)
		vary := fmt.Sprintf("%.2f", (closeTemp-open)/open)

		var tempRate float64 = 1
		quotation.mu.Lock()
		if item.Quote == "BTC" {
			tempRate = quotation.priceMap["btc"]
		}
		if item.Quote == "ETH" {
			tempRate = quotation.priceMap["eth"]
		}
		if item.Quote == "BNB" {
			tempRate = quotation.priceMap["bnb"]
		}
		quotation.mu.Unlock()

		cny := fmt.Sprintf("%.6f", closeTemp*tempRate*dollarRate)
		closePrice := fmt.Sprintf("%.6f", closeTemp)
		topic := event.Symbol
		quotation.mu.RLock()
		isExist := quotation.rsExist[topic]
		quotation.mu.RUnlock()

		if !isExist {
			quotation.mu.Lock()
			quotation.rsExist[topic] = true
			quotation.mu.Unlock()

			quotation.c1 <- QuotationItem{closePrice, cny, vary, topic, item.Base, item.Quote}
		}

	},
	)

	go quotation.collectResult(100, 50, 20)

}

func getBiSymbos() []BiItem {
	resp, err := http.Get(BiMarketUrl)
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
	var bi BiSymbos
	err = json.Unmarshal(symbolsByte, &bi)

	if err != nil {
		fmt.Println("解码出错", err)
	}
	return bi.List

}
