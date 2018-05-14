package main

import (
	"net/http"
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"github.com/leizongmin/huobiapi"

	//"github.com/astaxie/beego/utils"
	"strings"
	"log"

	"strconv"
	"time"
)
const HuoBiMarketUrl  =  "https://api.huobipro.com/v1/common/symbols"

type coinItem struct{
	Base string `json:"base-currency"`
	Quote string `json:"quote-currency"`
}
type coinAll struct{
	Status string
	List []coinItem `json:"data"`
}



func initHbKline(symbols []coinItem)  *KlineAll {

	ka := &KlineAll{}
	ka.All = make(map[string]*Kline)
	ka.ChAll = make(map[string]chan bool)
	ka.Symbos = symbols
	var key string
	for _,item := range ka.Symbos{
		key  = strings.ToLower(item.Base +item.Quote)
		ka.ChAll[key] = make(chan bool,1)
	}
	return ka;
}

func (k *KlineAll) subscribeAllHb()  {

	market, err := huobiapi.NewMarket("wss://api.huobi.pro/ws")
	if err != nil {
		fmt.Println("市场初始化错误",err)
	}
	if EntrustDebug {
		log.Println("开始订阅火币深度信息")
	}

	for _, item := range k.Symbos  {

		go func(item coinItem ) {
			market.Subscribe("market."+ item.Base +item.Quote+".depth.step0", func(topic string, data *huobiapi.JSON)  {

				symbol := item.Base+item.Quote

				bJson,_ := data.Get("tick").Get("bids").Encode()
				bids := new([][2]float64)
				json.Unmarshal(bJson,bids)

				aJson,_ := data.Get("tick").Get("asks").Encode()
				asks := new([][2]float64)
				json.Unmarshal(aJson,asks)

						kline,exist := k.All[symbol]
						if  !exist{
							kline = new(Kline)
							k.All[symbol] = kline
						}
						//if EntrustDebug && item.Base=="btcusdt"{
						//	log.Println("btcusdt深度获取成功")
						//}
						kline.Bids = *bids
						kline.Asks = *asks
						kline.Name = item.Base

						select {
						case k.ChAll[symbol] <- true:
							if EntrustDebug {
								log.Println("火币交易对 "+symbol+" 深度更新")
							}
						default:

						}


			})
		}(item)

	}
}



func (k *KlineAll) getOneResult(symbol string) *Kline {

	return k.All[symbol]
}

func (k *KlineAll) getChan(symbol string) chan bool{
	return k.ChAll[symbol]
}

func (k *KlineAll) getHuoBiSymbols() []coinItem {
	resp,err := http.Get(HuoBiMarketUrl);
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
	err = json.Unmarshal(symbolsByte,&ca);
	if err != nil {
		fmt.Println("解码出错",err)
	}
	if EntrustDebug {
		//log.Println("火币网货币信息获取完成",ca.List)
	}
		return ca.List

}

func  (gn *KlineAll) compareAll(hb *KlineAll,symbols map[string]string) {
	for key,item := range  symbols  {
		go gn.compareWait(hb,key,item)
	}
}

func (gn *KlineAll)compareWait(hb *KlineAll,symbol string,hbSymbol string) {

	c1 := hb.getChan(hbSymbol)
	c2 := gn.getChan(symbol)

	var r1,r2 bool

	if EntrustDebug {
		log.Println("等待获取两个交易所的对应交易对 "+symbol+" 的深度信息")
	}
	for  {
		time.Sleep(EntrustDuration*time.Second)
		select {
		case <- c1:
			r1 = true
			if r2 {
				if EntrustDebug {
					log.Println("火币网交易对深度信息 "+hbSymbol+" 获取成功")
				}
				gn.compare(hb,symbol,hbSymbol)
				gn.judgeFinish(symbol)
				r2,r1 = false,false

			}
		case <-c2:
			r2 = true
			if EntrustDebug {
				log.Println("公牛交易对深度信息 "+symbol+" 获取成功")
			}
			if r1 {
				gn.compare(hb,symbol,hbSymbol)
				gn.judgeFinish(symbol)
				r2,r1 = false,false

			}
		}

	}


}

func (gn *KlineAll) judgeFinish(symbol string )  {
	var i int
	for {
		select {
		case <- gn.ChSingle[symbol]:
			i++
			if i == 2 {
				if EntrustDebug {
					fmt.Println("买入和卖出全部完成")
				}
				goto quit
			}
		}
	}
	quit:
}
var ca *Kline


func  (gn *KlineAll) compare(hb *KlineAll,symbol string,hbSymbol string)  {

	r1 := hb.getOneResult(hbSymbol)

	r2 := gn.getOneResult(symbol)

	if EntrustDebug {
		log.Println("火币和公牛交易对  "+symbol+"  的深度信息分别为",r1,"-----------",r2)
	}
	rsBuy  := CalDiff(r1.Bids,r2.Bids)
	rsSell := CalDiff(r1.Asks,r2.Asks)

	rsBuy.Symbol = symbol
	rsBuy.Ch = make(chan int,EntrustNumber)
	rsBuy.EType = "buy-limit"

	rsSell.Symbol = symbol
	rsSell.Ch = make(chan int,EntrustNumber)
	rsSell.EType = "sell-limit"
	if EntrustDebug {
		log.Println(symbol,"开始挂单",rsBuy,rsSell)

	}

	go gn.EntrustMany(rsBuy)
	go gn.EntrustMany(rsSell)

}

func (gn *KlineAll) EntrustMany(rs *EntrustS)  {
	for price,number := range rs.Diff {
		go Entrust(rs.Symbol,1,strconv.FormatFloat(number, 'f', -1, 32),strconv.FormatFloat(price, 'f', -1, 32),rs.EType,"",rs.Ch)
	}
	var i int
	ids := make([]int,EntrustNumber)
	for id := range rs.Ch {
		ids = append(ids,id)
		i++
		if i == EntrustNumber{
			if EntrustDebug {
				close(rs.Ch)
				gn.ChSingle[rs.Symbol] <- struct{}{}
				log.Println(rs.Symbol,rs.Diff,ids,rs.EType,"成功")
			}
		}
	}
}

func CalDiff(r1,r2 [][2]float64)  *EntrustS{

	es := &EntrustS{}

	es.Diff = make(map[float64]float64)

	for _, i1 := range r1 {

		price,_ := strconv.ParseFloat(strconv.FormatFloat(i1[0]/PriceMultiple,'f',5,64),64)
		number,_ := strconv.ParseFloat(strconv.FormatFloat(i1[1],'f',5,64),64)

		if len(r2) == 0 {
			es.Diff[price] = number

		} else {
			var exist bool
			for _, i2 := range r2 {

				//如果公牛中存在但数量不足
				if  price == i2[0] {
					exist = true
					rest := i1[1] - i2[1]
					if rest > 0 {
						//判断结果中是否存在
						number,_:= strconv.ParseFloat(strconv.FormatFloat(rest,'f',5,64),64)

						es.Diff[price] += number

						//es.Diff[i1[0]] += rest
					}

				}
			}
			if !exist {
				es.Diff[price] = number
				exist = false
			}

		}

		if len(es.Diff) >= EntrustNumber {
			return es
		}
	}
	return es
}




func Abs(f float64) float64 {
	if f < 0 {
		return float64(-f)
	}
	return float64(f)
}