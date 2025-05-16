package variadic

import "fmt"

const (
	A = "A"
	B = "B"
	C = "C"
)

const (
	PrefA = "A"
	PrefB = "B"
	PrefC = "C"
)

// Sorted variadic arguments with Println
func sortedPrintln() {
	fmt.Println(A, B, C)
}

// Unsorted variadic arguments with Println
func unsortedPrintln() {
	fmt.Println(C, A, B)
}

// Sorted variadic arguments with Println
func sortedPrintlnPref() {
	fmt.Println(PrefA, PrefB, PrefC)
}

// Unsorted variadic arguments with Println
func unsortedPrintlnPref() {
	fmt.Println(PrefC, PrefA, PrefB) // want "variadic arguments are not sorted"
}
