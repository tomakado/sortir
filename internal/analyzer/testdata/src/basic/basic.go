package basic

import "fmt"

const (
	A = 1
	B = 2
)

var (
	AVar = 1
	BVar = 2
)

type S struct {
	AField int
	BField int
}

type I interface {
	AMethod()
	BMethod()
}

func Print(a, b int, args ...int) {
	fmt.Println(a, b, args)
}

var m = map[string]int{
	"a": 1,
	"b": 2,
}