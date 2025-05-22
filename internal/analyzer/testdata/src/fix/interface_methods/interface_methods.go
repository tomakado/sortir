package interface_methods

type UnsortedInterface interface {
	ZMethod()
	AMethod() // want "interface methods are not sorted"
	BMethod()
}

type AnotherInterface interface {
	Write([]byte) (int, error)
	Close() error // want "interface methods are not sorted"
	Read([]byte) (int, error)
}