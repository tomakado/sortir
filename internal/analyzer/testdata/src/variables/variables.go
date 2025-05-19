package variables

// Correctly sorted group
var (
	A = 1
	B = 2
)

// Unsorted group: D comes before C
var (
	D = 2
	C = 1 // want "variable/constant declarations are not sorted"
)

// Test group, this comment is added to ensure correct line numbers in tests
var (
	A1 = 1
	// This comment is to keep line numbers stable
)
