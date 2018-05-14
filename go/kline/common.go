package main

type KlineItem struct {
	Price  string `json:"price"`
	Amount string `json:"amount"`
}
type Tick struct {
	Bids [][2]float64 `json:"bids"`
	Asks [][2]float64 `json:"asks"`
}

type Kline struct {
	Bids [][2]float64
	Asks [][2]float64
	Name string
}
type KlineAll struct {
	All      map[string]*Kline
	ChAll    map[string]chan bool
	Symbos   []coinItem
	ChSingle map[string]chan struct{} //保证挂单完成之后才开始下一轮

}

type EntrustItem struct {
	Price  float64
	Amount float64
}

type EntrustS struct {
	Symbol string
	Diff   map[float64]float64
	Ch     chan int
	EType  string
}

const (
	EntrustDuration  = 2   //下单间隔
	EntrustNumber    = 10  //每次下单数量
	PriceMultiple    = 100 //价格倍数
	GnDepthAddress   = "192.168.0.200:8091"
	GnEntrustAddress = "http://200-coin.dsceshi.cn/rpc"
)

var EntrustDebug = false
