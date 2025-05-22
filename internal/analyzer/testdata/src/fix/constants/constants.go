package constants

const (
	C = 3
	A = 1 // want "variable/constant declarations are not sorted"
	B = 2
)

const (
	Z = "z"
	X = "x" // want "variable/constant declarations are not sorted"
	Y = "y"
)