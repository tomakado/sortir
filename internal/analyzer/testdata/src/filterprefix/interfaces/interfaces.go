package interfaces

// Interface methods with prefix defined in config, should be checked
type MyInterface1 interface {
	PrefB()
	PrefA() // want "interface methods are not sorted"
}

// Interface methods without prefix defined in config, shouldn't be checked
type MyInterface2 interface {
	E()
	D()
}
