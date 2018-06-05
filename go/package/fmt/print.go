package main

import (
	"fmt"
	"os"
)

type Data struct {
	a int
	b string
}

var d1, d2 = Data{1, "fff"}, Data{2, "ssss"}

//Print采用默认格式将其参数格式化并写入标准输出。如果两个相邻的参数都不是字符串，会在它们的输出之间添加空格。返回写入的字节数和遇到的任何错误。
func t_print() {
	fmt.Print(d1, d2)       //{1 fff} {2 ssss}
	fmt.Print("aaa", "bbb") //aaabbb
}

//通用
//%v	值的默认格式表示
//%+v	类似%v，但输出结构体时会添加字段名
//%#v	值的Go语法表示
//%T	值的类型的Go语法表示
//布尔值
//%t	单词true或false
//整数
//%b	表示为二进制
//%c	该值对应的unicode码值
//%d	表示为十进制
//%o	表示为八进制
//%q	该值对应的单引号括起来的go语法字符字面值，必要时会采用安全的转义表示
//浮点数
//%b	无小数部分、二进制指数的科学计数法，如-123456p-78；参见strconv.FormatFloat
//%e	科学计数法，如-1234.456e+78
//%E	科学计数法，如-1234.456E+78
//%f	有小数部分但无指数部分，如123.456
//字符串
//%s	直接输出字符串或者[]byte
//%q	该值对应的双引号括起来的go语法字符串字面值，必要时会采用安全的转义表示
//%x	每个字节用两字符十六进制数表示（使用a-f）
//%X	每个字节用两字符十六进制数表示（使用A-F）

//%f:    默认宽度，默认精度
//%9f    宽度9，默认精度
//%.2f   默认宽度，精度2
//%9.2f  宽度9，精度2
//%9.f   宽度9，精度0

//Printf根据format参数生成格式化的字符串并写入标准输出

func t_printf() {
	fmt.Printf("%#v", d1, d2) //main.Data{a:1, b:"fff"}%!(EXTRA main.Data={2 ssss})
	fmt.Printf("%q", 123)     //'{'
	fmt.Println()
	fmt.Printf("%8.2f  %3.2f", 123.0, 44.555)
}

//Fprintf根据format参数生成格式化的字符串并写入w
func t_fprintf() {
	file, _ := os.OpenFile("abc.txt", os.O_WRONLY, 0666)
	fmt.Fprintf(file, "%+v", d1)
}

func main() {
	t_printf()
}
