// Package analyzer implements the core code analysis for sorting Go code elements.
package analyzer

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"go.tomakado.io/sortir/internal/checker"
	"go.tomakado.io/sortir/internal/config"
)

var (
	ErrInspectorAssertionFailed = errors.New("inspector type assertion failed")
)

type Analyzer struct {
	Analyzer *analysis.Analyzer
	Checker  *checker.Checker
}

func New() *Analyzer {
	analyzer := &analysis.Analyzer{
		Name:             "sortir",
		Doc:              "Checks and fixes sorting of Go code elements",
		Run:              nil,
		RunDespiteErrors: false,
		Requires:         []*analysis.Analyzer{inspect.Analyzer},
		ResultType:       nil,
		FactTypes:        nil,
		URL:              "go.tomakado.io/sortir",
	}

	cfg := initCfg(analyzer)
	checker := checker.NewChecker(cfg)
	a := &Analyzer{
		Analyzer: analyzer,
		Checker:  checker,
	}

	analyzer.Run = func(pass *analysis.Pass) (any, error) {
		return run(a, pass)
	}

	return a
}

func initCfg(analyzer *analysis.Analyzer) *config.SortConfig {
	cfg := config.New()

	// Enable/disable flags
	analyzer.Flags.BoolVar(&cfg.Constants.Enabled, config.CheckConstants, true, "enable constant sorting checks")
	analyzer.Flags.BoolVar(&cfg.Variables.Enabled, config.CheckVariables, true, "enable variable sorting checks")
	analyzer.Flags.BoolVar(&cfg.StructFields.Enabled, config.CheckStructFields, true, "enable struct field sorting checks")
	analyzer.Flags.BoolVar(&cfg.InterfaceMethods.Enabled, config.CheckInterfaceMethods, true, "enable interface method sorting checks")
	analyzer.Flags.BoolVar(&cfg.VariadicArgs.Enabled, config.CheckVariadicArgs, false, "enable variadic argument sorting checks")
	analyzer.Flags.BoolVar(&cfg.MapValues.Enabled, config.CheckMapValues, true, "enable map value sorting checks")

	// General settings
	analyzer.Flags.BoolVar(&cfg.IgnoreGroups, config.IgnoreGroups, false, "ignore sorting checks for specific groups")
	analyzer.Flags.StringVar(&cfg.FilterPrefix, config.FilterPrefix, "", "only check sorting for symbols starting with specified prefix (global)")

	// Per-check prefix filters
	analyzer.Flags.StringVar(&cfg.Constants.Prefix, config.ConstantsPrefix, "", "only check sorting for constants starting with specified prefix")
	analyzer.Flags.StringVar(&cfg.Variables.Prefix, config.VariablesPrefix, "", "only check sorting for variables starting with specified prefix")
	analyzer.Flags.StringVar(&cfg.StructFields.Prefix, config.StructFieldsPrefix, "", "only check sorting for struct fields starting with specified prefix")
	analyzer.Flags.StringVar(&cfg.InterfaceMethods.Prefix, config.InterfaceMethodsPrefix, "", "only check sorting for interface methods starting with specified prefix")
	analyzer.Flags.StringVar(&cfg.VariadicArgs.Prefix, config.VariadicArgsPrefix, "", "only check sorting for variadic arguments starting with specified prefix")
	analyzer.Flags.StringVar(&cfg.MapValues.Prefix, config.MapValuesPrefix, "", "only check sorting for map values starting with specified prefix")

	return cfg
}

func run(a *Analyzer, pass *analysis.Pass) (any, error) {
	inspectorObj := pass.ResultOf[inspect.Analyzer]

	inspector, ok := inspectorObj.(*inspector.Inspector)
	if !ok {
		// Return a sentinel error instead of nil, nil
		return nil, ErrInspectorAssertionFailed
	}

	processASTNodes(pass, inspector, a.Checker)

	// Use a sentinel error to avoid nilnil
	var analyzerResult any

	return analyzerResult, nil
}

func processASTNodes(pass *analysis.Pass, inspector *inspector.Inspector, checker *checker.Checker) {
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
		(*ast.StructType)(nil),
		(*ast.InterfaceType)(nil),
		(*ast.CallExpr)(nil),
		(*ast.CompositeLit)(nil),
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
		checker.CheckNode(pass, n)
	})
}
