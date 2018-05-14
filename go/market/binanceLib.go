package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var baseUrl = "wss://stream.binance.com:9443/stream?streams="

type WsKlineHandler func(event *KlineItem, item BiItem)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
)

type WsKline struct {
	StartTime            int64  `json:"t"`
	EndTime              int64  `json:"T"`
	Symbol               string `json:"s"`
	Interval             string `json:"i"`
	FirstTradeID         int64  `json:"f"`
	LastTradeID          int64  `json:"L"`
	Open                 string `json:"o"`
	Close                string `json:"c"`
	High                 string `json:"h"`
	Low                  string `json:"l"`
	Volume               string `json:"v"`
	TradeNum             int64  `json:"n"`
	IsFinal              bool   `json:"x"`
	QuoteVolume          string `json:"q"`
	ActiveBuyVolume      string `json:"V"`
	ActiveBuyQuoteVolume string `json:"Q"`
}
type KlineItem struct {
	Event  string  `json:"e"`
	Time   int64   `json:"E"`
	Symbol string  `json:"s"`
	Kline  WsKline `json:"k"`
}
type WsKlineEvent struct {
	Topic string    `json:"stream"`
	Data  KlineItem `json:"data"`
}

func biLibCon(queryMap []BiItem, handler WsKlineHandler) {

	m := make(map[string]BiItem)

	for _, item := range queryMap {
		m[strings.ToLower(item.Base+item.Quote)] = item
		queryItem := fmt.Sprintf("%s@kline_%s/", strings.ToLower(item.Base+item.Quote), "1d")

		baseUrl += queryItem
	}

	c, _, err := websocket.DefaultDialer.Dial(baseUrl, nil)
	if err != nil {
		fmt.Println("连接出错")
	}

	go func() {
		defer func() {
			c.Close()
		}()

		for {

			_, message, err := c.ReadMessage()
			//err = errors.New("this is a new error")
			if err != nil {
				fmt.Println("读取消息出错", err)
				c.Close()
				time.Sleep(time.Second)
				c, _, err = websocket.DefaultDialer.Dial(baseUrl, nil)

				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Println("error: %v", err)
				}
				//break
			}
			event := new(WsKlineEvent)
			err = json.Unmarshal(message, event)
			item := m[strings.ToLower(event.Data.Symbol)]
			//fmt.Println(string(message))
			go handler(&event.Data, item)
		}
	}()

}

func MinuteCon(queryMap []string, handler WsKlineHandler) {

	for _, item := range queryMap {
		queryItem := fmt.Sprintf("%s@kline_%s/", strings.ToLower(item), "1m")

		baseUrl += queryItem
	}

	c, _, err := websocket.DefaultDialer.Dial(baseUrl, nil)

	if err != nil {
		fmt.Println("连接出错")
	}
	go func() {
		defer func() {
			c.Close()
		}()
		for {

			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println("读取消息出错", err)
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Println("error: %v", err)
				}
				break
			}

			event := new(WsKlineEvent)
			err = json.Unmarshal(message, event)
			//fmt.Println(string(message))
			go handler(&event.Data, BiItem{})
		}
	}()
}
