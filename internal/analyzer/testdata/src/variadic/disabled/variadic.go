package disabled

import "fmt"

// Sorted variadic arguments with Println
func sortedPrintln() {
	fmt.Println("a", "b", "c")
}

// Unsorted variadic arguments with Println
func unsortedPrintln() {
	fmt.Println("c", "a", "b")
}

// Helper function for testing
func printArgs(args ...int) {
	fmt.Println(args)
}

// Call with sorted arguments
func sortedPrintArgs() {
	printArgs(1, 2, 3)
}

// Call with unsorted arguments
func unsortedPrintArgs() {
	printArgs(3, 1, 2)
}
