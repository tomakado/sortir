package interfaces

// Correctly sorted interface
type SortedInterface interface {
	A()
	B()
}

// Unsorted interface
type UnsortedInterface interface {
	B()
	A() // want "interface methods are not sorted"
}

// Correctly sorted interface with groups
type SortedWithGroups interface {
	A()
	B()
	
	C()
	D()
}

// First group unsorted
type UnsortedGroup1 interface {
	B()
	A() // want "interface methods are not sorted"
}

// Second group unsorted
type UnsortedGroup2 interface {
	D()
	C() // want "interface methods are not sorted"
}

// Interface with embedded interfaces
type WithEmbedded interface {
	SortedInterface
	UnsortedInterface
	
	E()
	F()
}