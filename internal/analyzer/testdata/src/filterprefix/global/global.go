package global

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
