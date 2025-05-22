// Package analyzer implements the core code analysis for sorting Go code elements.
package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"go.tomakado.io/sortir/internal/config"
	"go.tomakado.io/sortir/internal/log"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out mocks/logger.go -pkg mocks . Logger

type Logger interface {
	Verbose(string, ...any)
}

type Analyzer struct {
	cfg *config.SortConfig

	analyzer    *analysis.Analyzer
	diagnostics []Diagnostic
	logger      Logger
}

func New() *Analyzer {
	analyzer := &analysis.Analyzer{
		Doc:              "Checks and fixes sorting of Go code elements",
		Name:             "sortir",
		Requires:         []*analysis.Analyzer{inspect.Analyzer},
		RunDespiteErrors: false,
		URL:              "go.tomakado.io/sortir",
	}

	a := &Analyzer{analyzer: analyzer}
	a.initCfg()
	a.logger = log.NewLogger(a.cfg.LogLevel())

	analyzer.Run = a.run

	return a
}

func (a *Analyzer) WithConfig(cfg *config.SortConfig) *Analyzer {
	a.cfg = cfg
	return a
}

func (a *Analyzer) Analyzer() *analysis.Analyzer {
	return a.analyzer
}

func (a *Analyzer) run(pass *analysis.Pass) (any, error) {
	a.logger.Verbose("Starting analysis", log.FieldPackage, pass.Pkg.Path())
	a.logger.Verbose("Config", "config", a.cfg)

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
		a.CheckNode(pass, n)
	})

	a.logger.Verbose("Analysis complete", log.FieldPackage, pass.Pkg.Path())
	return nil, nil
}

func (a *Analyzer) initCfg() {
	a.cfg = config.New()

	a.analyzer.Flags.StringVar(
		&a.cfg.GlobalPrefix,
		config.FlagFilterPrefix,
		config.Default[string](config.FlagFilterPrefix),
		"only check sorting for symbols starting with specified prefix (global)",
	)

	a.analyzer.Flags.BoolVar(
		&a.cfg.IgnoreGroups,
		config.FlagIgnoreGroups,
		config.Default[bool](config.FlagIgnoreGroups),
		"ignore sorting checks for specific groups",
	)

	a.analyzer.Flags.BoolVar(
		&a.cfg.Verbose,
		config.FlagVerbose,
		config.Default[bool](config.FlagVerbose),
		"enable verbose logging",
	)

	a.analyzer.Flags.BoolVar(
		&a.cfg.Constants.Enabled,
		config.FlagConstants,
		config.Default[bool](config.FlagConstants),
		"enable constant sorting checks",
	)

	a.analyzer.Flags.StringVar(
		&a.cfg.Constants.Prefix,
		config.FlagConstantsPrefix,
		config.Default[string](config.FlagConstantsPrefix),
		"only check sorting for constants starting with specified prefix",
	)

	a.analyzer.Flags.BoolVar(
		&a.cfg.Variables.Enabled,
		config.FlagVariables,
		config.Default[bool](config.FlagVariables),
		"enable variable sorting checks",
	)

	a.analyzer.Flags.StringVar(
		&a.cfg.Variables.Prefix,
		config.FlagVariablesPrefix,
		config.Default[string](config.FlagVariablesPrefix),
		"only check sorting for variables starting with specified prefix",
	)

	a.analyzer.Flags.BoolVar(
		&a.cfg.StructFields.Enabled,
		config.FlagStructFields,
		config.Default[bool](config.FlagStructFields),
		"enable struct field sorting checks",
	)

	a.analyzer.Flags.StringVar(
		&a.cfg.StructFields.Prefix,
		config.FlagStructFieldsPrefix,
		config.Default[string](config.FlagStructFieldsPrefix),
		"only check sorting for struct fields starting with specified prefix",
	)

	a.analyzer.Flags.BoolVar(
		&a.cfg.InterfaceMethods.Enabled,
		config.FlagInterfaceMethods,
		config.Default[bool](config.FlagInterfaceMethods),
		"enable interface method sorting checks",
	)

	a.analyzer.Flags.StringVar(
		&a.cfg.InterfaceMethods.Prefix,
		config.FlagInterfaceMethodsPrefix,
		config.Default[string](config.FlagInterfaceMethodsPrefix),
		"only check sorting for interface methods starting with specified prefix",
	)

	a.analyzer.Flags.BoolVar(
		&a.cfg.VariadicArgs.Enabled,
		config.FlagVariadicArgs,
		config.Default[bool](config.FlagVariadicArgs),
		"enable variadic argument sorting checks",
	)

	a.analyzer.Flags.StringVar(
		&a.cfg.VariadicArgs.Prefix,
		config.FlagVariadicArgsPrefix,
		config.Default[string](config.FlagVariadicArgsPrefix),
		"only check sorting for variadic arguments starting with specified prefix",
	)

	a.analyzer.Flags.BoolVar(
		&a.cfg.MapKeys.Enabled,
		config.FlagMapKeys,
		config.Default[bool](config.FlagMapKeys),
		"enable map value sorting checks",
	)

	a.analyzer.Flags.StringVar(
		&a.cfg.MapKeys.Prefix,
		config.FlagMapKeysPrefix,
		config.Default[string](config.FlagMapKeysPrefix),
		"only check sorting for map values starting with specified prefix",
	)
}

func (a *Analyzer) CheckNode(pass *analysis.Pass, node ast.Node) bool {
	switch n := node.(type) {
	case *ast.GenDecl:
		a.logger.Verbose("Checking GenDecl node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return a.checkGenDecl(pass, n)
	case *ast.StructType:
		a.logger.Verbose("Checking StructType node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return a.checkStructType(pass, n)
	case *ast.InterfaceType:
		a.logger.Verbose("Checking InterfaceType node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return a.checkInterfaceType(pass, n)
	case *ast.CallExpr:
		a.logger.Verbose("Checking CallExpr node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return a.checkCallExpr(pass, n)
	case *ast.CompositeLit:
		a.logger.Verbose("Checking CompositeLit node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return a.checkCompositeLit(pass, n)
	}
	return true
}

func (a *Analyzer) CheckGenDecl(pass *analysis.Pass, node *ast.GenDecl) bool {
	return a.checkGenDecl(pass, node)
}

func (a *Analyzer) CheckStructType(pass *analysis.Pass, node *ast.StructType) bool {
	return a.checkStructType(pass, node)
}

func (a *Analyzer) CheckInterfaceType(pass *analysis.Pass, node *ast.InterfaceType) bool {
	return a.checkInterfaceType(pass, node)
}

func (a *Analyzer) CheckCallExpr(pass *analysis.Pass, node *ast.CallExpr) bool {
	return a.checkCallExpr(pass, node)
}

func (a *Analyzer) CheckCompositeLit(pass *analysis.Pass, node *ast.CompositeLit) bool {
	return a.checkCompositeLit(pass, node)
}

func (a *Analyzer) checkGenDecl(pass *analysis.Pass, node *ast.GenDecl) bool {
	var prefix string
	switch node.Tok {
	case token.CONST:
		prefix = a.cfg.Constants.Prefix
		a.logger.Verbose("Processing constants", log.FieldEnabled, a.cfg.Constants.Enabled, log.FieldPrefix, prefix)
	case token.VAR:
		prefix = a.cfg.Variables.Prefix
		a.logger.Verbose("Processing variables", log.FieldEnabled, a.cfg.Variables.Enabled, log.FieldPrefix, prefix)
	}

	isConst := node.Tok == token.CONST
	isVar := node.Tok == token.VAR

	shouldCheck := isConst && a.cfg.Constants.Enabled || isVar && a.cfg.Variables.Enabled
	if !shouldCheck {
		a.logger.Verbose("Skipping checks", log.FieldNodeType, node.Tok.String())
		return true
	}

	valueSpecs := make([]*ast.ValueSpec, 0, len(node.Specs))
	for _, spec := range node.Specs {
		valueSpecs = append(valueSpecs, spec.(*ast.ValueSpec))
	}

	a.logger.Verbose("Extracting metadata", log.FieldSpecsCount, len(valueSpecs), log.FieldIgnoreGroups, a.cfg.IgnoreGroups)
	metadata := extractMetadata(pass, valueSpecs, extractGenDecl, a.cfg.IgnoreGroups)
	a.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, prefix, log.FieldGlobalPrefix, a.cfg.GlobalPrefix)
	return a.checkElementsSorted(
		pass,
		metadata,
		prefix,
		"variable/constant declarations are not sorted",
	)
}

type checkParams struct {
	countField   string
	enabled      bool
	errorMessage string
	extractFunc  func(*analysis.Pass, *ast.Field) (string, token.Pos, int)
	fieldList    []*ast.Field
	itemName     string
	prefix       string
	skipMessage  string
}

func (a *Analyzer) checkFieldList(pass *analysis.Pass, params checkParams) bool {
	a.logger.Verbose("Processing "+params.itemName, log.FieldEnabled, params.enabled, log.FieldPrefix, params.prefix)
	if !params.enabled {
		a.logger.Verbose("Skipping " + params.skipMessage)
		return true
	}

	a.logger.Verbose("Extracting metadata", params.countField, len(params.fieldList), log.FieldIgnoreGroups, a.cfg.IgnoreGroups)
	metadata := extractMetadata(pass, params.fieldList, params.extractFunc, a.cfg.IgnoreGroups)
	a.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, params.prefix, log.FieldGlobalPrefix, a.cfg.GlobalPrefix)
	return a.checkElementsSorted(
		pass,
		metadata,
		params.prefix,
		params.errorMessage,
	)
}

func (a *Analyzer) checkStructType(pass *analysis.Pass, node *ast.StructType) bool {
	return a.checkFieldList(pass, checkParams{
		countField:   log.FieldFieldsCount,
		enabled:      a.cfg.StructFields.Enabled,
		errorMessage: "struct fields are not sorted",
		extractFunc:  extractStructField,
		fieldList:    node.Fields.List,
		itemName:     "struct fields",
		prefix:       a.cfg.StructFields.Prefix,
		skipMessage:  "struct field checks",
	})
}

func (a *Analyzer) checkInterfaceType(pass *analysis.Pass, node *ast.InterfaceType) bool {
	return a.checkFieldList(pass, checkParams{
		countField:   log.FieldMethodsCount,
		enabled:      a.cfg.InterfaceMethods.Enabled,
		errorMessage: "interface methods are not sorted",
		extractFunc:  extractInterfaceMethod,
		fieldList:    node.Methods.List,
		itemName:     "interface methods",
		prefix:       a.cfg.InterfaceMethods.Prefix,
		skipMessage:  "interface method checks",
	})
}

func (a *Analyzer) checkCallExpr(pass *analysis.Pass, node *ast.CallExpr) bool {
	a.logger.Verbose("Processing variadic arguments", log.FieldEnabled, a.cfg.VariadicArgs.Enabled, log.FieldPrefix, a.cfg.VariadicArgs.Prefix)
	if !a.cfg.VariadicArgs.Enabled {
		a.logger.Verbose("Skipping variadic argument checks")
		return true
	}

	a.logger.Verbose("Extracting metadata", log.FieldArgsCount, len(node.Args), log.FieldIgnoreGroups, a.cfg.IgnoreGroups)
	metadata := extractVariadicArgMetadata(pass, node, a.cfg.IgnoreGroups)
	a.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, a.cfg.VariadicArgs.Prefix, log.FieldGlobalPrefix, a.cfg.GlobalPrefix)
	return a.checkElementsSorted(
		pass,
		metadata,
		a.cfg.VariadicArgs.Prefix,
		"variadic arguments are not sorted",
	)
}

func (a *Analyzer) checkCompositeLit(pass *analysis.Pass, node *ast.CompositeLit) bool {
	a.logger.Verbose("Processing map keys", "composite_lit_type", fmt.Sprintf("%#v", node.Type), log.FieldEnabled, a.cfg.MapKeys.Enabled, log.FieldPrefix, a.cfg.MapKeys.Prefix)
	if !a.cfg.MapKeys.Enabled {
		a.logger.Verbose("Skipping map key checks")
		return true
	}

	keyValueExprs := make([]*ast.KeyValueExpr, 0, len(node.Elts))
	for _, elt := range node.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			keyValueExprs = append(keyValueExprs, kv)
		}
	}

	a.logger.Verbose("Extracting metadata", log.FieldKeyValueCount, len(keyValueExprs), log.FieldIgnoreGroups, a.cfg.IgnoreGroups)
	metadata := extractMetadata(pass, keyValueExprs, extractMapKey, a.cfg.IgnoreGroups)
	a.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, a.cfg.MapKeys.Prefix, log.FieldGlobalPrefix, a.cfg.GlobalPrefix)
	return a.checkElementsSorted(
		pass,
		metadata,
		a.cfg.MapKeys.Prefix,
		"composite literal elements are not sorted",
	)
}

func (a *Analyzer) report(pass *analysis.Pass, diagnostic Diagnostic) {
	a.logger.Verbose("Reporting diagnostic", log.FieldDiagnostic, diagnostic)
	a.diagnostics = append(a.diagnostics, diagnostic)

	pass.Report(diagnostic.AsGoAnalysisDiagnostic())
}

func (a *Analyzer) CheckElementsSorted(
	pass *analysis.Pass,
	groups [][]Metadata,
	prefix, msg string,
) bool {
	return a.checkElementsSorted(pass, groups, prefix, msg)
}

func (a *Analyzer) checkElementsSorted(
	pass *analysis.Pass,
	groups [][]Metadata,
	prefix, msg string,
) bool {

	allSorted := true
	for groupIdx, group := range groups {
		if len(group) <= 1 {
			a.logger.Verbose("Skipping group with less than 2 elements", log.FieldGroupIndex, groupIdx, log.FieldGroupSize, len(group))
			continue
		}

		a.logger.Verbose("Checking group sorting", log.FieldGroupIndex, groupIdx, log.FieldGroupSize, len(group))
		groupNeedsSorting := false
		var unsortedIndex int

		for i := 1; i < len(group); i++ {
			if !hasPrefixOrGlobal(group[i].Value, prefix, a.cfg.GlobalPrefix) {
				a.logger.Verbose("Skipping element - no matching prefix", log.FieldElement, group[i].Value, log.FieldPrefix, prefix, log.FieldGlobalPrefix, a.cfg.GlobalPrefix)
				continue
			}

			if group[i].Value < group[i-1].Value {
				allSorted = false
				groupNeedsSorting = true
				unsortedIndex = i
				if pass.Fset != nil {
					a.logger.Verbose("Found unsorted elements", log.FieldCurrent, group[i].Value, log.FieldPrevious, group[i-1].Value, log.FieldPosition, pass.Fset.Position(group[i].Position))
				} else {
					a.logger.Verbose("Found unsorted elements", log.FieldCurrent, group[i].Value, log.FieldPrevious, group[i-1].Value, log.FieldPosition, group[i].Position)
				}
				break
			}
		}

		if groupNeedsSorting {
			elementType := getElementType(msg)
			fix := a.generateFix(pass, group, elementType)

			a.report(pass, Diagnostic{
				From:       group[unsortedIndex].Position,
				Message:    msg,
				Suggestion: fix,
			})
		}
	}

	a.logger.Verbose("Sorting check complete", log.FieldDiagnosticsCount, len(a.diagnostics), log.FieldAllSorted, allSorted)
	return allSorted
}

func hasPrefixOrGlobal(name, prefix, globalPrefix string) bool {
	if prefix == "" {
		return hasPrefix(name, globalPrefix)
	}
	return hasPrefix(name, prefix)
}

func hasPrefix(name, prefix string) bool {
	if prefix == "" {
		return true
	}

	if name == "" {
		return false
	}

	return strings.HasPrefix(name, prefix)
}

func getElementType(msg string) string {
	switch msg {
	case "variable/constant declarations are not sorted":
		return "declarations"
	case "struct fields are not sorted":
		return "struct fields"
	case "interface methods are not sorted":
		return "interface methods"
	case "variadic arguments are not sorted":
		return "variadic arguments"
	case "composite literal elements are not sorted":
		return "map keys"
	default:
		return "elements"
	}
}
