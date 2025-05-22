package variables

var (
	zebra = "z"
	apple = "a" // want "variable/constant declarations are not sorted"
	banana = "b"
)

var (
	z int
	x int // want "variable/constant declarations are not sorted"
	y int
)