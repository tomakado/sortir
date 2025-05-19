// Package analyzer implements the core code analysis for sorting Go code elements.
package analyzer

import (
	"go/ast"
	"sync"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"go.tomakado.io/sortir/internal/analyzer/checker"
	"go.tomakado.io/sortir/internal/config"
	"go.tomakado.io/sortir/internal/log"
)

type Analyzer struct {
	cfg      *config.SortConfig
	initOnce sync.Once

	analyzer *analysis.Analyzer
	checker  *checker.Checker
	logger   *log.Logger
}

func New() *Analyzer {
	analyzer := &analysis.Analyzer{
		Name:             "sortir",
		Doc:              "Checks and fixes sorting of Go code elements",
		RunDespiteErrors: false,
		Requires:         []*analysis.Analyzer{inspect.Analyzer},
		URL:              "go.tomakado.io/sortir",
	}

	a := &Analyzer{
		cfg:      initCfg(analyzer),
		analyzer: analyzer,
	}
	analyzer.Run = a.run

	return a
}

func (a *Analyzer) Analyzer() *analysis.Analyzer {
	return a.analyzer
}

func (a *Analyzer) Checker() *checker.Checker {
	a.initState()
	return a.checker
}

func (a *Analyzer) run(pass *analysis.Pass) (any, error) {
	a.initState()

	a.logger.Verbose("Starting analysis", log.FieldPackage, pass.Pkg.Path())

	inspectorObj := pass.ResultOf[inspect.Analyzer]

	inspector, ok := inspectorObj.(*inspector.Inspector)
	if !ok {
		panic("inspectorObj is not *inspector.Inspector")
	}

	a.logger.Verbose("Processing AST nodes")
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
		(*ast.StructType)(nil),
		(*ast.InterfaceType)(nil),
		(*ast.CallExpr)(nil),
		(*ast.CompositeLit)(nil),
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
		a.checker.CheckNode(pass, n)
	})

	a.logger.Verbose("Analysis complete", log.FieldPackage, pass.Pkg.Path())
	return nil, nil
}

func (a *Analyzer) initState() {
	a.initOnce.Do(func() {
		a.logger = log.New(a.cfg.LogLevel())
		a.checker = checker.New(a.cfg).WithLogger(a.logger)
	})
}

func initCfg(analyzer *analysis.Analyzer) *config.SortConfig {
	cfg := config.New()

	analyzer.Flags.StringVar(&cfg.GlobalPrefix, config.FlagFilterPrefix, "", "only check sorting for symbols starting with specified prefix (global)")
	analyzer.Flags.BoolVar(&cfg.IgnoreGroups, config.FlagIgnoreGroups, false, "ignore sorting checks for specific groups")
	analyzer.Flags.BoolVar(&cfg.Verbose, config.FlagVerbose, false, "enable verbose logging")

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
