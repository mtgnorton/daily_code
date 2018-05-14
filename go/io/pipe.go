package main

import (
	"errors"
	"fmt"
	"io"
	"time"
)

func readClose() {
	// Pipe 在内存中创建一个同步管道，用于不同区域的代码之间相互传递数据。
	// 返回的 *PipeReader 用于从管道中读取数据，*PipeWriter 用于向管道中写入数据。
	// 管道没有缓冲区，读写操作可能会被阻塞。可以安全的对管道进行并行的读、写或关闭
	// 操作，读写操作会依次执行，Close 会在被阻塞的 I/O 操作结束之后完成。

	r, w := io.Pipe()
	// 启用一个例程进行读取
	go func() {
		buf := make([]byte, 5)
		for n, err := 0, error(nil); err == nil; {
			n, err = r.Read(buf)

			// 向管道中写入数据，如果管道被关闭，则会返会一个错误信息：
			// 1、如果读取端通过 CloseWithError 方法关闭了管道，则返回关闭时传入的错误信息。
			// 2、如果读取端通过 Close 方法关闭了管道，则返回 io.ErrClosedPipe。
			// 3、如果是写入端关闭了管道，则返回 io.ErrClosedPipe。

			r.CloseWithError(errors.New("管道被读取端关闭"))
			fmt.Printf("读取：%d, %v, %s\n", n, err, buf[:n])
		}
	}()
	// 主例程进行写入
	n, err := w.Write([]byte("Hello World !"))
	fmt.Printf("写入：%d, %v\n", n, err)
}

func writeClose() {
	r, w := io.Pipe()
	// 启用一个例程进行读取
	go func() {
		buf := make([]byte, 5)
		for n, err := 0, error(nil); err == nil; {
			n, err = r.Read(buf)
			fmt.Printf("读取：%d, %v, %s\n", n, err, buf[:n])
		}
	}()
	// 主例程进行写入
	n, err := w.Write([]byte("Hello World !"))
	fmt.Printf("写入：%d, %v\n", n, err)

	// 从管道中读取数据，如果管道被关闭，则会返会一个错误信息：
	// 1、如果写入端通过 CloseWithError 方法关闭了管道，则返回关闭时传入的错误信息。
	// 2、如果写入端通过 Close 方法关闭了管道，则返回 io.EOF。
	// 3、如果是读取端关闭了管道，则返回 io.ErrClosedPipe。
	w.CloseWithError(errors.New("管道被写入端关闭"))
	n, err = w.Write([]byte("Hello World !"))
	fmt.Printf("写入：%d, %v\n", n, err)
	time.Sleep(time.Second * 1)
}

func main() {
	//readClose()
	writeClose()
}
