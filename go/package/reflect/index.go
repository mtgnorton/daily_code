package main

import (
	"fmt"
	"io"
	"os"
	"reflect"
)

type D struct {
	I int
	S string
}

var d = D{
	1, "ss",
}

var f = 1.23

func typeOf() {
	dt := reflect.TypeOf(d)
	ft := reflect.TypeOf(f)
	fmt.Println(dt, ft) //main.D float64

	fmt.Println(dt.Kind()) // 表示底层数据的类型 struct
	fmt.Println("========")

}

func valueOf() {
	dv := reflect.ValueOf(d)
	fv := reflect.ValueOf(f)
	fmt.Printf("%T %v  %T %v \n", dv, dv, fv, fv) //reflect.Value {1 ss}  reflect.Value 1.23

	fmt.Println(dv.Kind())         //struct
	fmt.Printf("%T \n", fv.Kind()) //reflect.Kind
	fmt.Println(dv.Type())         //main.D
	fmt.Println(fv.Float())        //提取数据 1.23
	fmt.Println("========")
}

func useInterface() {
	fv := reflect.ValueOf(f)
	y := fv.Interface().(float64) // y will have type float64.
	fmt.Printf("%T %v\n", y, y)
	fmt.Println("========")
}

func set() {
	var x float64 = 3.4
	p := reflect.ValueOf(&x)                     // Note: take the address of x.
	fmt.Println("type of p:", p.Type())          //type of p: *float64
	fmt.Println("settability of p:", p.CanSet()) //settability of p: false
	v := p.Elem()
	fmt.Println("settability of v:", v.CanSet()) //settability of v: true
	v.SetFloat(7.1)
	fmt.Printf("%T %v \n", v.Interface(), v.Interface()) //float64 7.1
	fmt.Println(x)                                       //7.1
	fmt.Println("========")

}

func reader() {
	var r io.Reader
	var e interface{}
	file, _ := os.OpenFile("c1.txt", os.O_RDONLY, 0666)
	r = file
	e = file
	fmt.Printf("%T %v \n", r, r) //*os.File &{0xc042074780}
	fmt.Printf("%T %v \n", e, e) //*os.File &{0xc042074780}
	rv := reflect.ValueOf(e).Interface()
	fmt.Printf("%T %v \n", rv, rv) //*os.File &{0xc042074780} 空接口值内部包含了具体值的类型信息 Printf 函数会恢复类型信息。
	fmt.Println("========")

}

func useStruct() {
	t := D{23, "skidoo"}
	s := reflect.ValueOf(&t).Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Printf("%d: %s %s = %v\n", i,
			typeOfT.Field(i).Name, f.Type(), f.Interface()) //0: I int = 23  // 1: S string = skidoo
	}
}

func main() {
	typeOf()
	valueOf()
	useInterface()
	set()
	reader()
	useStruct()
}
