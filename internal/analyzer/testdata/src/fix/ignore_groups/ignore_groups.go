package ignore_groups

const (
	Group1Z = 3
	Group1A = 1 // want "variable/constant declarations are not sorted"
	Group1B = 2

	Group2Z = "z"
	Group2A = "a"
	Group2B = "b"
)

var (
	z1 = "z1"
	a1 = "a1" // want "variable/constant declarations are not sorted"
	b1 = "b1"

	z2 = "z2"
	a2 = "a2"
	b2 = "b2"
)

type MyStruct struct {
	Z1 int
	A1 int // want "struct fields are not sorted"
	B1 int

	Z2 string
	A2 string
	B2 string
}

type MyInterface interface {
	ZMethod()
	AMethod() // want "interface methods are not sorted"
	BMethod()

	ZMethod2()
	AMethod2()
	BMethod2()
}