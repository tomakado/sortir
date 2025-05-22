package variadic_args

import "fmt"

func main() {
	fmt.Println("zebra", "apple", "banana") // want "variadic arguments are not sorted"
	
	fmt.Printf("%s %s %s\n", "z", "a", "b") // want "variadic arguments are not sorted"
}