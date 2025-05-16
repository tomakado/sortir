package structs

type Struct1 struct {
	PrefC error
	PrefB int // want "struct fields are not sorted"
	PrefA string
}

type Struct2 struct {
	E float64
	D bool
}
