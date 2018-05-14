package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
)

//获取美元汇率
func GetRate() (rate float64) {

	var exchangeUrl = "http://web.juhe.cn:8080/finance/exchange/rmbquot?key=e950f2756be34232c7f278abda1ab07a"
	resp, err := http.Get(exchangeUrl)

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", exchangeUrl, err)
		os.Exit(1)
	}

	rs := make(map[string]interface{})
	err = json.Unmarshal(b, &rs)

	if err != nil {
		fmt.Println(err)
	} else {
		result := rs["result"]

		temp := result.([]interface{})
		if len(temp) == 0 {
			return 6.40
		}

		v := result.([]interface{})[0].(map[string]interface{})

		for _, item := range v {
			value := item.(map[string]interface{})
			if value["name"] == "美元" {
				rate := value["bankConversionPri"].(string)
				frate, _ := strconv.ParseFloat(rate, 64)
				return Round(frate/100, 2)

			}
		}
	}
	return 6.40
}

func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}
