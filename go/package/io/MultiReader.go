package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	r1 := strings.NewReader("Hello World!")
	r2 := strings.NewReader("ABCDEFG")
	r3 := strings.NewReader("abcdefg")

	b := make([]byte, 15)

	// MultiReader 将多个 Reader 封装成一个单独的 Reader，多个 Reader 会按顺序读
	// 取，当多个 Reader 都返回 EOF 之后，单独的 Reader 才返回 EOF，否则返回读取
	// 过程中遇到的任何错误。
	mr := io.MultiReader(r1, r2, r3)

	for n, err := 0, error(nil); err == nil; {
		n, err = mr.Read(b)
		fmt.Printf("%q\n", b[:n])
	}
	// "Hello World!"
	// "ABCDEFG"
	// "abcdefg"
	// ""

	r1.Seek(0, 0)
	r2.Seek(0, 0)
	r3.Seek(0, 0)
	mr = io.MultiReader(r1, r2, r3)
	io.Copy(os.Stdout, mr)
	// Hello World!ABCDEFGabcdefg
}
