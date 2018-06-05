package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	r := strings.NewReader("Hello World!")
	b := make([]byte, 15)

	// TeeReader 对 r 进行封装，使 r 在读取数据的同时，自动向 w 中写入数据。
	// 它是一个无缓冲的 Reader，所以对 w 的写入操作必须在 r 的 Read 操作结束
	// 之前完成。所有写入时遇到的错误都会被作为 Read 方法的 err 返回。
	tr := io.TeeReader(r, os.Stdout)

	n, err := tr.Read(b)                  // Hello World!
	fmt.Printf("\n%s   %v\n", b[:n], err) // Hello World!   <nil>
}
