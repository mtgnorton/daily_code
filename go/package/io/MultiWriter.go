package main

import (
	"io"
	"os"
	"strings"
)

func main() {
	r := strings.NewReader("Hello World!\n")

	//MultiWriter将向自身写入的数据同步写入到所有 writers 中
	mw := io.MultiWriter(os.Stdout, os.Stdout, os.Stdout)

	r.WriteTo(mw)
	// Hello World!
	// Hello World!
	// Hello World!
}
