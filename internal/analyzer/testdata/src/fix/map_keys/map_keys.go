package map_keys

var m1 = map[string]int{
	"zebra": 3,
	"apple": 1, // want "composite literal elements are not sorted"
	"banana": 2,
}

func foo() {
	m2 := map[string]string{
		"z": "Z",
		"a": "A", // want "composite literal elements are not sorted"
		"b": "B",
	}
	_ = m2
}