package main

import (
	//"github.com/leizongmin1/huobiapi"
	"golang.org/x/net/websocket"

	"fmt"
	//"github.com/leek-box/sheep/util"
	"encoding/json"
	"bytes"
	"encoding/binary"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	//"learn/treasure/ws"
	"sync"
	"strings"
)
func ParseGzip(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, data)
	r, err := gzip.NewReader(b)
	if err != nil  && err!=io.EOF{
		//logger.Info("[ParseGzip] NewReader error: %v, maybe data is ungzip", err)
		return nil, err
	} else {
		defer r.Close()
		undatas, err := ioutil.ReadAll(r)
		if err != nil && err!=io.EOF{
			log.Println(err.Error())
			//logger.Warn("[ParseGzip]  ioutil.ReadAll error: %v", err)
			return nil, err
		}
		return undatas, nil
	}
}

type Han func (*Tick)

type GN struct {
	mu sync.RWMutex
	KlineHan map[string]Han
	ws  *websocket.Conn
}
func  NewGn() *GN {
	return &GN{
		KlineHan:make(map[string]Han),
	}
}
func (g *GN) ConnectGn()  {
	servAddr := GnDepthAddress
	ws, err := websocket.Dial("ws://"+servAddr+"/ws", "", "http://"+servAddr)
	g.ws = ws
	if err != nil {
		fmt.Print("公牛连接出错",err)
		return
	}
	go g.Receive()

}


func (g *GN) Subscribe(symbol string,handle Han)  {

	g.ws.Write([]byte(`{"sub": "market.`+symbol+`.depth.step0","pick":["bids.100","asks.100"],"id": "id1"}`))

	g.ws.Write([]byte(`{"req": "market.`+symbol+`.depth.step0","pick":["bids.100","asks.100"],"id": "id2"}`))


	g.mu.Lock()
	g.KlineHan[strings.ToLower("market."+symbol+".depth.step0")]= handle
	g.mu.Unlock()

}
func (g *GN) Receive()  {
	for {

		if g.ws !=nil{
			var data string
			err := websocket.Message.Receive(g.ws, &data)

			if err!=nil{
				fmt.Println(err)
				return
			}
			b,err:=ParseGzip([]byte(data))

			if err!=nil{
				fmt.Println(err)
				return
			}
			//判断PING
			resultInterface := make(map[string]interface{})
			// 不推荐使用json.Unmarshal
			decoder := json.NewDecoder(bytes.NewBuffer(b))
			decoder.UseNumber() // 此处能够保证bigint的精度
			decoder.Decode(&resultInterface)
			ping,ok := resultInterface["ping"]

			if ok{
				str := "{\"pong\": "+ping.(json.Number).String()+"}"
				g.ws.Write([]byte(str))
				continue
			}
			symbol,ok := resultInterface["rep"];
			if ok {
				g.mu.RLock()
				if handle,exist :=  g.KlineHan[symbol.(string)];exist{
					g.mu.RUnlock()
					t := &Tick{}
					tickJson,_ := json.Marshal(resultInterface["tick"])
					err = json.Unmarshal(tickJson,t)
					handle(t)
				}
			}


			symbol,ok =  resultInterface["ch"];


			if ok {
				g.mu.RLock()
				if handle,exist :=  g.KlineHan[symbol.(string)];exist{
					g.mu.RUnlock()
					t := &Tick{}
					tickJson,_ := json.Marshal(resultInterface["tick"])
					err = json.Unmarshal(tickJson,t)
					handle(t)
				}
			}


			if err!=nil{
				fmt.Println(err)
				break
			}

		}
	}

}

