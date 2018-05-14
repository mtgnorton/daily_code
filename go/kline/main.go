package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var d = flag.Bool("d", false, "is debug")

func main() {

	flag.Parse()

	EntrustDebug = *d
	//EntrustDebug = true
	m, _ := readFile("./coin.json")

	gaSymbols := make([]string, len(m))
	hbSymbols := make([]coinItem, len(m))

	i := 0
	for key, value := range m {

		gaSymbols[i] = key
		temp := strings.Split(value, "-")
		hbSymbols[i] = coinItem{temp[0], temp[1]}
		m[key] = temp[0] + temp[1]
		i++
	}

	ga, g := initGn(gaSymbols)

	ga.subscribeAllGn(g, gaSymbols)

	hb := initHbKline(hbSymbols)
	hb.subscribeAllHb()

	ga.compareAll(hb, m)

	select {}
}

func readFile(filename string) (map[string]string, error) {

	execpath, err := os.Executable() // 获得程序路径
	// handle err .// .

	_filepath := filepath.Join(filepath.Dir(execpath), filename)

	fmt.Println(_filepath)
	bytes, err := ioutil.ReadFile(_filepath)
	if err != nil {
		fmt.Println("ReadFile: ", err.Error())
		return nil, err
	}
	m := new(map[string]string)
	if err := json.Unmarshal(bytes, m); err != nil {
		fmt.Println("Unmarshal: ", err.Error())
		return nil, err
	}
	return *m, nil
}

func test() {
	symbols := []string{"bhtusdt"}
	ga, g := initGn(symbols)

	ga.subscribeAllGn(g, symbols)
	go func() {
		for {
			time.Sleep(5 * time.Second)

			Entrust("bhtusdt", 1, "0.0556", "94.3184", "sell-limit", "", make(chan int))

		}
	}()
}
