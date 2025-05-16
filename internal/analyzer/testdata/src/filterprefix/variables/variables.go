package constants

// Variables with prefix defined in config, should be checked
var (
	PrefE = "PrefE"
	PrefD = "PrefD" // want "variable/constant declarations are not sorted"
)

// Variables without prefix defined in config, shouldn't be checked
var (
	E = 3
	D = 4
)

// The prefix is defined for variables only
const (
	PrefB = "c"
	PrefA = "d" // want "variable/constant declarations are not sorted"
)

// Constants without prefix defined in config, shouldn't be checked
const (
	B = 1
	A = 2 // want "variable/constant declarations are not sorted"
)
