package analyzer_test

import (
	"testing"

	"go.tomakado.io/sortir/internal/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	t.Run("default config", func(t *testing.T) {
		t.Parallel()

		a := analyzer.New()
		analysistest.Run(t, testdata, a.Analyzer(),
			"basic",
			"constants",
			"variables",
			"struct_fields",
			"interfaces",
			"variadic/disabled",
			"map_keys",
		)
	})

	t.Run("variadic args enabled", func(t *testing.T) {
		t.Parallel()

		a := analyzer.New()
		cfg := *a.Checker().Config
		cfg.VariadicArgs.Enabled = true
		a.Checker().Config = &cfg

		analysistest.Run(t, testdata, a.Analyzer(),
			"variadic/enabled",
		)
	})
}

func TestAnalyzerWithPrefix(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	t.Run("global", func(t *testing.T) {
		t.Parallel()

		a := analyzer.New()
		cfg := *a.Checker().Config
		cfg.GlobalPrefix = "Pref"
		a.Checker().Config = &cfg

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/global")
	})

	t.Run("constants", func(t *testing.T) {
		t.Parallel()

		a := analyzer.New()
		cfg := *a.Checker().Config
		cfg.GlobalPrefix = ""
		cfg.Constants.Prefix = "Pref"
		a.Checker().Config = &cfg

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/constants")
	})

	t.Run("interface methods", func(t *testing.T) {
		t.Parallel()

		a := analyzer.New()
		cfg := *a.Checker().Config
		cfg.GlobalPrefix = ""
		cfg.InterfaceMethods.Prefix = "Pref"
		a.Checker().Config = &cfg

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/interfaces")
	})

	t.Run("structs", func(t *testing.T) {
		t.Parallel()

		a := analyzer.New()
		cfg := *a.Checker().Config
		cfg.GlobalPrefix = ""
		cfg.StructFields.Prefix = "Pref"
		a.Checker().Config = &cfg

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/structs")
	})

	t.Run("maps", func(t *testing.T) {
		t.Parallel()

		a := analyzer.New()
		cfg := *a.Checker().Config
		cfg.GlobalPrefix = ""
		cfg.MapKeys.Prefix = "Pref"
		a.Checker().Config = &cfg

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/maps")
	})

	t.Run("variadic", func(t *testing.T) {
		t.Parallel()

		a := analyzer.New()
		cfg := *a.Checker().Config
		cfg.GlobalPrefix = ""
		cfg.VariadicArgs.Enabled = true
		cfg.VariadicArgs.Prefix = "Pref"
		a.Checker().Config = &cfg

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/variadic")
	})
}
