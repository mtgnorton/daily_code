package main

import (
	//"fmt"
	//"fmt"
	"net/http"
	"fmt"
	"log"
	"strings"

)

func main()  {

	binance,biSymbos := initBiQuotation();
	binance.BiConnect(biSymbos)


	huobi,huoSymbos := initHuoBiQuotation();

	huobi.huoBiConnect(huoSymbos)



	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		var searchMany = ""

		for k, v := range r.Form {
			if k=="name" && strings.Join(v, "")=="huobi" {
				fmt.Fprintf(w, huobi.getResultJson())
			}else if  k=="name" && strings.Join(v, "")=="binance"  {
				fmt.Fprintf(w, binance.getResultJson())

			}else if k=="search" {
				searchValue := strings.Join(v, "")

				r1 := 	huobi.searchCoin(searchValue,"huobi");
				r2 :=  binance.searchCoin(searchValue,"binance")
				r3 := append(r1,r2...)
				if len(r3) ==0  {
					fmt.Fprintf(w,"{}")
					return
				}
				fmt.Fprintf(w,transferJson(r3))
			}else if k== "searchMany" {
				searchMany = strings.Join(v, "")
			}else if k=="exchange"{
				exchange := strings.Join(v,"")
				if exchange == "huobi" {
					rm :=huobi.searchManyCoin(searchMany,"huobi")
					fmt.Fprintf(w,transferJson(rm))

				}else{
					rm :=binance.searchManyCoin(searchMany,"binance")
					fmt.Fprintf(w,transferJson(rm))

				}
			}
		}
	}) //设置访问的路由

	err := http.ListenAndServe(":8081", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	select {

	}
}

