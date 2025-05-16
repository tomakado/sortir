package maps

var map1 = map[string]int{
	"PrefB": 1,
	"PrefA": 2, // want "map keys are not sorted"
}

var map2 = map[string]int{
	"E": 3,
	"D": 4,
}
