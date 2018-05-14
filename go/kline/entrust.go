package main

import (
	"log"
	"github.com/shopspring/decimal"
	"github.com/gorilla/rpc/json"
	//"net/rpc/jsonrpc"
	"net/http"
	"bytes"

	"io/ioutil"
	js "encoding/json"
	"fmt"
	"strconv"
)


type EntrustRequest struct {
	Userid    int64           `json:"userid,omitempty"`
	Symbol    string          `json:"symbol,omitempty"`
	Amount    decimal.Decimal `json:"amount,omitempty"`
	Price     decimal.Decimal `json:"price,omitempty"`
	Type      string          `json:"type,omitempty"`//sell-limit buy-limit
	UseCoin   string          `json:"dollar,omitempty"`
	OrderType string           `json:"ordertype"` //normal,leverage,blowing-up
}

type EntrustRs struct {
	Id string `json:"result"`
}

func Entrust(symbol string,userid int64,number string,price string,t string,usecoin string,ch chan int)interface{}{
	url := GnEntrustAddress
	var n,_=decimal.NewFromString(number)
	var p,_=decimal.NewFromString(price)

	args := EntrustRequest{
		Userid:         userid,
		Symbol:        symbol,
		Amount:        n,
		Price:         p,
		Type:          t,
		UseCoin:       usecoin,
		OrderType:   "normal",
	}
	message, err := json.EncodeClientRequest("Entrust.Entrust", args)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))


	if err != nil {
		log.Fatalf("%s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error in sending request to %s. %s", url, err)
	}
	defer resp.Body.Close()


	jsonstr,_:=ioutil.ReadAll(resp.Body)


	es := &EntrustRs{}

	js.Unmarshal(jsonstr,es)
	//fmt.Println(string(jsonstr))
	id ,err:= strconv.Atoi(es.Id)
	if err != nil{
		fmt.Println(string(jsonstr))
	}


	if id > 0 {
		ch <- id
	}else {
		ch <- 1
	}

	return nil
}

