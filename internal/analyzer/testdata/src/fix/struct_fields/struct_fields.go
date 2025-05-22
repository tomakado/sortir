package struct_fields

type UnsortedStruct struct {
	Zebra string
	Apple int // want "struct fields are not sorted"
	Banana bool `json:"banana"`
}

type AnotherStruct struct {
	Z int
	A int // want "struct fields are not sorted"
	B int
}