package constants

// Correctly sorted group
const (
	A = 1
	B = 2
)

// Unsorted group: D comes before C
const (
	D = 2
	C = 1 // want "constant declarations are not sorted"
)

// Test group, this comment is added to ensure correct line numbers in tests
const (
	A1 = 1
	// This comment is to keep line numbers stable
)