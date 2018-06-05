package main

import (
	"bytes"
	"fmt"
)

func main() {
	//将 s 转换为 []byte 后，包装成 bytes.Buffer 对象。
	rd := bytes.NewBufferString("Hello World!")

	buf := make([]byte, 6)

	cap := rd.Cap()
	fmt.Println("容量：", cap)

	// 引用未读取部分的数据切片（不移动读取位置）
	b := rd.Bytes()
	// 读出一部分数据，看看切片有没有变化
	rd.Read(buf)

	fmt.Printf("%s\n", rd.String()) // World!
	fmt.Printf("%s\n\n", b)         // Hello World!

	// 写入一部分数据，看看切片有没有变化
	rd.Write([]byte("abcdefg"))
	fmt.Printf("%s\n", rd.String()) // World!abcdefg
	fmt.Printf("%s\n\n", b)         // Hello World!

	// 再读出一部分数据，看看切片有没有变化
	rd.Read(buf)
	fmt.Printf("%s\n", rd.String()) // abcdefg
	fmt.Printf("%s\n", b)           // Hello World!
}
