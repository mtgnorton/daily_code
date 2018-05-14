package main

import (
	"strings"
	"log"

)

func initGn(symbols []string) (ga *KlineAll,g *GN){
	ga = &KlineAll{}
	ga.All = make(map[string]*Kline)
	ga.ChAll = make(map[string]chan bool)
	ga.ChSingle = make(map[string]chan struct{})

	var key = ""
	for _,item := range symbols{
		key  = strings.ToLower(item)
		ga.ChAll[key] = make(chan bool,1)
		ga.ChSingle[key] = make(chan struct{},2)
	}

	g = NewGn()
	g.ConnectGn()

	return ga,g
}

func(ga *KlineAll) subscribeAllGn(g *GN,symbols []string)  {

	if EntrustDebug {
		log.Println("开始订阅公牛深度信息")
		}
		for _,symbol := range symbols{
		go func(symbol string) {
			g.Subscribe(symbol, func(tick *Tick) {

				kline,exist := ga.All[symbol]

					if  !exist{
						kline = new(Kline)
						ga.All[symbol] = kline
					}
					kline.Bids = tick.Bids
					kline.Asks = tick.Asks

					kline.Name = symbol

				select {
					case ga.ChAll[symbol] <- true:
						if EntrustDebug {
							log.Println("公牛交易对"+symbol+"深度更新")
					}
				   default:

				}

			})
		}(symbol)
	}

}

