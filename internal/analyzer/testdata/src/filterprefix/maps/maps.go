package maps

var map1 = map[string]int{
	"PrefB": 1,
	"PrefA": 2, // want "composite literal elements are not sorted"
}

var map2 = map[string]int{
	"E": 3,
	"D": 4,
}
