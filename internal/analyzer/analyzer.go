// Package analyzer implements the core code analysis for sorting Go code elements.
package analyzer

import (
	"errors"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"go.tomakado.io/sortir/internal/analyzer/checker"
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

	analyzer.Flags.StringVar(&cfg.GlobalPrefix, config.FlagFilterPrefix, "", "only check sorting for symbols starting with specified prefix (global)")
	analyzer.Flags.BoolVar(&cfg.IgnoreGroups, config.FlagIgnoreGroups, false, "ignore sorting checks for specific groups")

	analyzer.Flags.BoolVar(&cfg.Constants.Enabled, config.FlagConstants, true, "enable constant sorting checks")
	analyzer.Flags.StringVar(&cfg.Constants.Prefix, config.FlagConstantsPrefix, "", "only check sorting for constants starting with specified prefix")

	analyzer.Flags.BoolVar(&cfg.Variables.Enabled, config.FlagVariables, true, "enable variable sorting checks")
	analyzer.Flags.StringVar(&cfg.Variables.Prefix, config.FlagVariablesPrefix, "", "only check sorting for variables starting with specified prefix")

	analyzer.Flags.BoolVar(&cfg.StructFields.Enabled, config.FlagStructFields, true, "enable struct field sorting checks")
	analyzer.Flags.StringVar(&cfg.StructFields.Prefix, config.FlagStructFieldsPrefix, "", "only check sorting for struct fields starting with specified prefix")

	analyzer.Flags.BoolVar(&cfg.InterfaceMethods.Enabled, config.FlagInterfaceMethods, true, "enable interface method sorting checks")
	analyzer.Flags.StringVar(&cfg.InterfaceMethods.Prefix, config.FlagInterfaceMethodsPrefix, "", "only check sorting for interface methods starting with specified prefix")

	analyzer.Flags.BoolVar(&cfg.VariadicArgs.Enabled, config.FlagVariadicArgs, false, "enable variadic argument sorting checks")
	analyzer.Flags.StringVar(&cfg.VariadicArgs.Prefix, config.FlagVariadicArgsPrefix, "", "only check sorting for variadic arguments starting with specified prefix")

	analyzer.Flags.BoolVar(&cfg.MapKeys.Enabled, config.FlagMapKeys, true, "enable map value sorting checks")
	analyzer.Flags.StringVar(&cfg.MapKeys.Prefix, config.FlagMapKeysPrefix, "", "only check sorting for map values starting with specified prefix")

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
