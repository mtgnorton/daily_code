package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {

	//NewBuffer使用buf作为初始内容创建并初始化一个Buffer。本函数用于创建一个用于读取已存在数据的buffer；也用于指定用于写入的内部缓冲的大小，此时，buf应为一个具有指定容量但长度为0的切片。buf会被作为返回值的底层缓冲切片。大多数情况下，new(Buffer)（或只是声明一个Buffer类型变量）就足以初始化一个Buffer了。

	buf := bytes.NewBuffer([]byte("Hello World!"))
	b := make([]byte, buf.Len())

	n, err := buf.Read(b) //将buf中内容读取到b中，此时buf为空

	fmt.Printf("%s   %v\n", b[:n], err)

	buf.WriteString("ABCDEFG\n")

	//WriteTo从缓冲中读取数据直到缓冲内没有数据或遇到错误，并将这些数据写入w。返回值n为从b读取并写入w的字节数
	buf.WriteTo(os.Stdout)

	//Write将b的内容写入缓冲中，如必要会增加缓冲容量。返回值n为len(p)，err总是nil。如果缓冲变得太大，Write会采用错误值ErrTooLarge引发panic。
	n, err = buf.Write(b)

	fmt.Printf("%d   %s   %v\n", n, buf.String(), err)

	//ReadByte读取并返回缓冲中的下一个字节。如果没有数据可用，返回值err为io.EOF。
	c, err := buf.ReadByte()
	fmt.Printf("%c   %s   %v\n", c, buf.String(), err)

	c, err = buf.ReadByte()
	fmt.Printf("%c   %s   %v\n", c, buf.String(), err)

	err = buf.UnreadByte()
	fmt.Printf("%s   %v\n", buf.String(), err)

	//UnreadByte吐出最近一次读取操作读取的最后一个字节。如果最后一次读取操作之后进行了写入，本方法会返回错误。
	err = buf.UnreadByte()
	fmt.Printf("%s   %v\n", buf.String(), err)
	// ello World!   bytes.Buffer: UnreadByte: previous operation was not a read
}
