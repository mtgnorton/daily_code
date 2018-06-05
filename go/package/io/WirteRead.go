package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// 示例：WriteString、ReadAtLeast、ReadFull
func main() {

	// WriteString 将字符串 s 写入到 w 中，返回写入的字节数和遇到的错误。
	// 如果 w 实现了 WriteString 方法，则优先使用该方法将 s 写入 w 中。
	// 否则，将 s 转换为 []byte，然后调用 w.Write 方法将数据写入 w 中。
	io.WriteString(os.Stdout, "Hello World!\n")
	// Hello World!

	r := strings.NewReader("Hello World!")
	b := make([]byte, 15)

	// ReadAtLeast 从 r 中读取数据到 b 中，要求至少读取 min 个字节。
	// 返回读取的字节数和遇到的错误。
	// 如果 min 超出了 buf 的容量，则 err 返回 io.ErrShortBuffer，否则：
	// 1、读出的数据长度 == 0  ，则 err 返回 EOF。
	// 2、读出的数据长度 <  min，则 err 返回 io.ErrUnexpectedEOF。
	// 3、读出的数据长度 >= min，则 err 返回 nil。
	n, err := io.ReadAtLeast(r, b, 20)
	fmt.Printf("%q   %d   %v\n", b[:n], n, err)
	// ""   0   short buffer

	//Seek方法设定下一次读写的位置：偏移量为offset，校准点由whence确定：0表示相对于文件起始；1表示相对于当前位置；2表示相对于文件结尾。Seek方法返回新的位置以及可能遇到的错误。
	r.Seek(0, 0)
	b = make([]byte, 15)

	//ReadFull从r精确地读取len(buf)字节数据填充进buf。函数返回写入的字节数和错误（如果没有读取足够的字节）。只有没有读取到字节时才可能返回EOF；如果读取了有但不够的字节时遇到了EOF，函数会返回ErrUnexpectedEOF。 只有返回值err为nil时，返回值n才会等于len(buf)。
	n, err = io.ReadFull(r, b)
	fmt.Printf("%q   %d   %v\n", b[:n], n, err)
	// "Hello World!"   12   unexpected EOF
}
