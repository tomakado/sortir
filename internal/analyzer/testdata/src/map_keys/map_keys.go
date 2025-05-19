package map_keys

// Sorted string keys
var sortedStringMap = map[string]int{
	"a": 1,
	"b": 2,
	"c": 3,
}

// Unsorted string keys
var unsortedStringMap = map[string]int{
	"c": 3,
	"a": 1, // want "composite literal elements are not sorted"
	"b": 2,
}

// Sorted int keys
var sortedIntMap = map[int]string{
	1: "a",
	2: "b",
	3: "c",
}

// Unsorted int keys
var unsortedIntMap = map[int]string{
	3: "c",
	1: "a", // want "composite literal elements are not sorted"
	2: "b",
}

// Map with groups separated by empty lines
var mapWithGroups = map[string]int{
	"a": 1,
	"b": 2,

	"c": 3,
	"d": 4,
}

// Unsorted map with groups
var unsortedMapWithGroups = map[string]int{
	"b": 2,
	"a": 1, // want "composite literal elements are not sorted"

	// Empty line resets sorting with default config
	"d": 4,
	"c": 3, // want "composite literal elements are not sorted"
}

// Non-sortable key type (not checked)
var mapWithNonSortableKeys = map[complex64]int{
	1 + 2i: 1,
	0 + 1i: 2,
}
