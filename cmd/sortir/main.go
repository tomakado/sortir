// Sortir is a Go linter and formatter that checks and fixes sorting of various Go code elements.
// It can analyze constant groups, variable groups, struct fields, interface methods,
// variadic arguments, and map values to ensure they are sorted consistently.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"go.tomakado.io/sortir/internal/analyzer"
)

func main() {
	analyzer := analyzer.New()
	singlechecker.Main(analyzer.Analyzer())
}
