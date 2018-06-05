package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// http get by proxy
func GetByProxy(url_addr, proxy_addr string) (*http.Response, error) {
	request, _ := http.NewRequest("GET", url_addr, nil)
	proxy, err := url.Parse(proxy_addr)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxy),
		},
	}
	return client.Do(request)
}

func main() {
	proxy := "http://118.171.27.77:3128/"
	url := "http://httpbin.org/get/"
	resp, err := GetByProxy(url, proxy)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
