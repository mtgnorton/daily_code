package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func readDir() {

	// ReadDir 读取指定目录中的所有目录和文件（不包括子目录）。
	// 返回读取到的文件信息列表和遇到的错误，列表是经过排序的。
	rd, err := ioutil.ReadDir("./")

	fmt.Println(err)

	for _, fi := range rd {
		if fi.IsDir() {
			fmt.Printf("[%s]\n", fi.Name())

		} else {
			fmt.Println(fi.Name())
		}
	}
}
func tempFile() {

	// TempDir 功能同 TempFile，只不过创建的是目录，返回目录的完整路径。
	dir, err := ioutil.TempDir("", "Test")
	if err != nil {
		fmt.Println(err)
	}
	defer os.Remove(dir) // 用完删除
	fmt.Printf("%s\n", dir)

	// TempFile 在 dir 目录中创建一个以 prefix 为前缀的临时文件，并将其以读
	// 写模式打开。返回创建的文件对象和遇到的错误。
	// 如果 dir 为空，则在默认的临时目录中创建文件（参见 os.TempDir），多次
	// 调用会创建不同的临时文件，调用者可以通过 f.Name() 获取文件的完整路径。
	// 调用本函数所创建的临时文件，应该由调用者自己删除。
	f, err := ioutil.TempFile(dir, "Test")
	if err != nil {
		fmt.Println(err)
	}
	defer os.Remove(f.Name()) // 用完删除
	fmt.Printf("%s\n", f.Name())
}

func readWrite() {
	//ReadFile 从filename指定的文件中读取数据并返回文件的内容。成功的调用返回的err为nil而非EOF。因为本函数定义为读取整个文件，它不会将读取返回的EOF视为应报告的错误。
	n, err := ioutil.ReadFile("abc.txt")
	fmt.Println(string(n), err)

	// WriteFile 向文件中写入数据，写入前会清空文件。
	// 如果文件不存在，则会以指定的权限创建该文件。
	// 返回遇到的错误。
	ioutil.WriteFile("cde.txt", []byte("hello 321312world"), 0666)

	f, err := os.OpenFile("cde.txt", os.O_RDONLY, 0666)
	//ReadAll从r读取数据直到EOF或遇到error，返回读取的数据和遇到的错误。成功的调用返回的err为nil而非EOF。因为本函数定义为读取r直到EOF，它不会将读取返回的EOF视为应报告的错误。
	content, err := ioutil.ReadAll(f)

	fmt.Println(string(content))

}
func main() {
	//readDir()
	//tempFile()
	readWrite()
}
