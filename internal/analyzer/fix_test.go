package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzerFixes(t *testing.T) {
	testdata := analysistest.TestData()

	tests := []struct {
		analyzer func() *Analyzer
		dir      string
		name     string
	}{
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/constants",
			name: "constants",
		},
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/variables",
			name: "variables",
		},
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/struct_fields",
			name: "struct_fields",
		},
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/interface_methods",
			name: "interface_methods",
		},
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/map_keys",
			name: "map_keys",
		},
		{
			analyzer: func() *Analyzer {
				a := New()
				a.cfg.VariadicArgs.Enabled = true
				return a
			},
			dir:  "fix/variadic_args",
			name: "variadic_args",
		},
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/multiname",
			name: "multiname_declarations",
		},
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/formatting",
			name: "preserve_formatting",
		},
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/empty_lines",
			name: "empty_lines",
		},
		{
			analyzer: func() *Analyzer {
				return New()
			},
			dir:  "fix/edge_cases",
			name: "edge_cases",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			analyzer := test.analyzer()
			analysistest.RunWithSuggestedFixes(t, testdata, analyzer.Analyzer(), test.dir)
		})
	}
}
