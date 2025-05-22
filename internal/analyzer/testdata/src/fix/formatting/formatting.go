package formatting

var (
	zebra    int
	apple  int // want "variable/constant declarations are not sorted"
	banana        int
)

type MyStruct struct {
	Zebra    string
	Apple  int // want "struct fields are not sorted"
	Banana        bool `json:"banana"`
}