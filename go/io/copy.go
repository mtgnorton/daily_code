package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// 示例：CopyN、Copy、CopyBuffer
func main() {
	r := strings.NewReader("Hello World!")
	buf := make([]byte, 32)

	// CopyN 从 src 中复制 n 个字节的数据到 dst 中，返回复制的字节数和遇到的错误。
	// 只有当 written = n 时，err 才返回 nil。
	n, err := io.CopyN(os.Stdout, r, 5) // Hello
	fmt.Printf("\n%d   %v\n\n", n, err) // 5   <nil>

	r.Seek(0, 0)

	// Copy 从 src 中复制数据到 dst 中，直到所有数据都复制完毕，返回复制的字节数和
	// 遇到的错误。如果复制过程成功结束，则 err 返回 nil，而不是 EOF，因为 Copy 的
	// 定义为“直到所有数据都复制完毕”，所以不会将 EOF 视为错误返回。
	n, err = io.Copy(os.Stdout, r)      // Hello World!
	fmt.Printf("\n%d   %v\n\n", n, err) // 12   <nil>

	r.Seek(0, 0)

	r2 := strings.NewReader("ABCDEFG")
	r3 := strings.NewReader("abcdefg")

	// CopyBuffer 相当于 Copy，只不过 Copy 在执行的过程中会创建一个临时的缓冲区来中
	// 转数据，而 CopyBuffer 则可以单独提供一个缓冲区，让多个复制操作共用同一个缓
	// 冲区，避免每次复制操作都创建新的缓冲区。如果 buf == nil，则 CopyBuffer 会
	// 自动创建缓冲区。
	n, err = io.CopyBuffer(os.Stdout, r, buf) // Hello World!
	fmt.Printf("\n%d   %v\n", n, err)         // 12   <nil>

	n, err = io.CopyBuffer(os.Stdout, r2, buf) // ABCDEFG
	fmt.Printf("\n%d   %v\n", n, err)          // 7   <nil>

	n, err = io.CopyBuffer(os.Stdout, r3, buf) // abcdefg
	fmt.Printf("\n%d   %v\n", n, err)          // 7   <nil>
}
