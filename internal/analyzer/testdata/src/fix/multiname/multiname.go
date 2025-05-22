package multiname

var (
	z, x, y int
	c, a, b string // want "variable/constant declarations are not sorted"
)

const (
	Z, X, Y = 3, 1, 2
	C, A, B = "c", "a", "b" // want "variable/constant declarations are not sorted"
)