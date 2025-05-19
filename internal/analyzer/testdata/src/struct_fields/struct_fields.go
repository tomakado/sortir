package structs

// Correctly sorted struct
type SortedStruct struct {
	A int
	B int
}

// Unsorted struct
type UnsortedStruct struct {
	B int
	A int // want "struct fields are not sorted"
}

// Correctly sorted struct with embedded fields
type SortedWithEmbedded struct {
	A
	B
	CField int
	DField int
}

// Unsorted struct with embedded fields
type UnsortedWithEmbedded struct {
	B
	A      // want "struct fields are not sorted"
	CField int
	DField int
}

// Correctly sorted structs separated by empty lines
type SortedWithGroups struct {
	A int
	B int

	D int
	E int
}

// Unsorted struct with groups
type UnsortedWithGroups struct {
	D int
	E int

	A int
	B int
}

// Unsorted fields in different groups
type UnsortedAcrossGroups struct {
	B int
	A int // want "struct fields are not sorted"

	E int
	D int // want "struct fields are not sorted"
}

// For anonymous field name extraction
type A struct{}
type B struct{}

// Placeholder for filter test
type PlaceholderStruct struct {
	Field int
}

// Struct init, sorted
var sortedStruct = SortedStruct{
	A: 1,
	B: 2,
}

// Struct init, unsorted
func foo() {
	_ = SortedStruct{
		B: 2,
		A: 1, // want "composite literal elements are not sorted"
	}
}
