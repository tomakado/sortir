package composite_literals

type Person struct {
	Name    string
	Age     int
	Country string
	City    string
	Street  string
}

func main() {
	// Multi-line struct literal with unsorted fields
	p1 := &Person{
		Street:  "Main St", // want `composite literal elements are not sorted`
		City:    "New York",
		Country: "USA",
		Age:     30,
		Name:    "John",
	}

	// Another multi-line struct literal
	p2 := Person{
		Country: "Canada", // want `composite literal elements are not sorted`
		City:    "Toronto",
		Street:  "Queen St",
		Name:    "Jane",
		Age:     25,
	}

	// Single-line struct literal (should remain single-line)
	p3 := Person{Country: "UK", City: "London", Age: 35, Name: "Bob", Street: "Baker St"} // want `composite literal elements are not sorted`

	// Map literal with multi-line format
	m := map[string]int{
		"zebra": 1, // want `composite literal elements are not sorted`
		"apple": 2,
		"mango": 3,
		"banana": 4,
	}

	_, _, _, _ = p1, p2, p3, m
}