package analyzer

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"go.tomakado.io/sortir/internal/config"
)

var analyzerConfig *config.SortConfig

func New(cfg *config.SortConfig) *analysis.Analyzer {
	analyzerConfig = cfg
	
	return &analysis.Analyzer{
		Name: "sortir",
		Doc:  "Checks and fixes sorting of Go code elements",
		Run:  run,
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	cfg := analyzerConfig
	
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),      // For const and var declarations
		(*ast.StructType)(nil),   // For struct fields
		(*ast.InterfaceType)(nil), // For interface methods
		(*ast.CallExpr)(nil),     // For variadic arguments
		(*ast.CompositeLit)(nil), // For map literals
	}

	inspector.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {
		case *ast.GenDecl:
			if (node.Tok == token.CONST && cfg.EnabledChecks.Constants) ||
			   (node.Tok == token.VAR && cfg.EnabledChecks.Variables) {
				inspectGenDecl(pass, node, cfg)
			}
		case *ast.StructType:
			if cfg.EnabledChecks.StructFields {
				inspectStructFields(pass, node, cfg)
			}
		case *ast.InterfaceType:
			if cfg.EnabledChecks.InterfaceMethods {
				inspectInterfaceMethods(pass, node, cfg)
			}
		case *ast.CallExpr:
			if cfg.EnabledChecks.VariadicArgs {
				inspectVariadicArgs(pass, node, cfg)
			}
		case *ast.CompositeLit:
			if cfg.EnabledChecks.MapValues {
				inspectMapLiteral(pass, node, cfg)
			}
		}
	})

	return nil, nil
}

func inspectGenDecl(pass *analysis.Pass, node *ast.GenDecl, cfg *config.SortConfig) {
	// Will implement checking of constant and variable declarations
}

func inspectStructFields(pass *analysis.Pass, node *ast.StructType, cfg *config.SortConfig) {
	// Will implement checking of struct fields
}

func inspectInterfaceMethods(pass *analysis.Pass, node *ast.InterfaceType, cfg *config.SortConfig) {
	// Will implement checking of interface methods
}

func inspectVariadicArgs(pass *analysis.Pass, node *ast.CallExpr, cfg *config.SortConfig) {
	// Will implement checking of variadic arguments
}

func inspectMapLiteral(pass *analysis.Pass, node *ast.CompositeLit, cfg *config.SortConfig) {
	// Will implement checking of map literals
}