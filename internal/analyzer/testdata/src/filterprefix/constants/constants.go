package constants

// Constants with prefix defined in config, should be checked
const (
	PrefB = "c"
	PrefA = "d" // want "variable/constant declarations are not sorted"
)

// Constants without prefix defined in config, shouldn't be checked
const (
	B = 1
	A = 2
)

// The prefix is defined for constants only
var (
	PrefE = "PrefE"
	PrefD = "PrefD" // want "variable/constant declarations are not sorted"
)

var (
	E = 3
	D = 4 // want "variable/constant declarations are not sorted"
)
