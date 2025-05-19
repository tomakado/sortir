package checker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/analysis"

	"go.tomakado.io/sortir/internal/config"
)

type CheckerMethodsTestSuite struct {
	suite.Suite
	fset    *token.FileSet
	checker *Checker
}

func TestCheckerMethods(t *testing.T) {
	suite.Run(t, new(CheckerMethodsTestSuite))
}

func (s *CheckerMethodsTestSuite) SetupTest() {
	s.fset = token.NewFileSet()
}

var diagnosticsAnalyzer = &analysis.Analyzer{Name: "diagnostics"}

func (s *CheckerMethodsTestSuite) createPass(src string) *analysis.Pass {
	file, err := parser.ParseFile(s.fset, "test.go", src, parser.ParseComments)
	s.Require().NoError(err)

	var diagnostics []analysis.Diagnostic
	conf := &types.Config{}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	pkg, err := conf.Check("test", s.fset, []*ast.File{file}, info)
	s.Require().NoError(err)

	pass := &analysis.Pass{
		Fset:      s.fset,
		Files:     []*ast.File{file},
		Pkg:       pkg,
		TypesInfo: info,
		Report: func(d analysis.Diagnostic) {
			diagnostics = append(diagnostics, d)
		},
		ResultOf: make(map[*analysis.Analyzer]any),
	}

	// No need for typeInfoCollector in these tests

	pass.Report = func(d analysis.Diagnostic) {
		diagnostics = append(diagnostics, d)
	}

	pass.ResultOf = make(map[*analysis.Analyzer]any)
	pass.ResultOf[diagnosticsAnalyzer] = &diagnostics

	return pass
}

func (s *CheckerMethodsTestSuite) getDiagnostics(pass *analysis.Pass) []analysis.Diagnostic {
	if result, ok := pass.ResultOf[diagnosticsAnalyzer].(*[]analysis.Diagnostic); ok {
		return *result
	}
	return nil
}

func (s *CheckerMethodsTestSuite) testGenDeclSorting(cfg *config.SortConfig, src string, expectedToken token.Token, shouldPass bool, expectedError string) {
	s.checker = NewChecker(cfg)
	pass := s.createPass(src)

	var genDecl *ast.GenDecl
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if g, ok := n.(*ast.GenDecl); ok && g.Tok == expectedToken {
			genDecl = g
			return false
		}
		return true
	})

	result := s.checker.checkGenDeclIfEnabled(pass, genDecl)
	s.Equal(shouldPass, result)

	diagnostics := s.getDiagnostics(pass)
	if shouldPass {
		s.Empty(diagnostics)
	} else {
		s.Len(diagnostics, 1)
		s.Contains(diagnostics[0].Message, expectedError)
	}
}

func (s *CheckerMethodsTestSuite) TestCheckGenDeclIfEnabled_ConstantsEnabled() {
	cfg := &config.SortConfig{
		Constants: &config.CheckConfig{
			Enabled: true,
			Prefix:  "C",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}

	src := `
package test

const (
	CB = 1
	CA = 2
)
`
	s.testGenDeclSorting(cfg, src, token.CONST, false, "variable/constant declarations are not sorted")
}

func (s *CheckerMethodsTestSuite) TestCheckGenDeclIfEnabled_ConstantsDisabled() {
	cfg := &config.SortConfig{
		Constants: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

const (
	CB = 1
	CA = 2
)
`
	pass := s.createPass(src)

	var genDecl *ast.GenDecl
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if g, ok := n.(*ast.GenDecl); ok && g.Tok == token.CONST {
			genDecl = g
			return false
		}
		return true
	})

	result := s.checker.checkGenDeclIfEnabled(pass, genDecl)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckGenDeclIfEnabled_VariablesEnabled() {
	cfg := &config.SortConfig{
		Variables: &config.CheckConfig{
			Enabled: true,
			Prefix:  "v",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}

	src := `
package test

var (
	vB = 1
	vA = 2
)
`
	s.testGenDeclSorting(cfg, src, token.VAR, false, "variable/constant declarations are not sorted")
}

func (s *CheckerMethodsTestSuite) TestCheckGenDeclIfEnabled_SortedVariables() {
	cfg := &config.SortConfig{
		Variables: &config.CheckConfig{
			Enabled: true,
			Prefix:  "v",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

var (
	vA = 1
	vB = 2
)
`
	pass := s.createPass(src)

	var genDecl *ast.GenDecl
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if g, ok := n.(*ast.GenDecl); ok && g.Tok == token.VAR {
			genDecl = g
			return false
		}
		return true
	})

	result := s.checker.checkGenDeclIfEnabled(pass, genDecl)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckStructTypeIfEnabled_Enabled() {
	cfg := &config.SortConfig{
		StructFields: &config.CheckConfig{
			Enabled: true,
			Prefix:  "f",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

type MyStruct struct {
	fB int
	fA string
}
`
	pass := s.createPass(src)

	var structType *ast.StructType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			structType = st
			return false
		}
		return true
	})

	result := s.checker.checkStructTypeIfEnabled(pass, structType)
	s.False(result)

	diagnostics := s.getDiagnostics(pass)
	s.Len(diagnostics, 1)
	s.Contains(diagnostics[0].Message, "struct fields are not sorted")
}

func (s *CheckerMethodsTestSuite) TestCheckStructTypeIfEnabled_Disabled() {
	cfg := &config.SortConfig{
		StructFields: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

type MyStruct struct {
	fB int
	fA string
}
`
	pass := s.createPass(src)

	var structType *ast.StructType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			structType = st
			return false
		}
		return true
	})

	result := s.checker.checkStructTypeIfEnabled(pass, structType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckStructTypeIfEnabled_Sorted() {
	cfg := &config.SortConfig{
		StructFields: &config.CheckConfig{
			Enabled: true,
			Prefix:  "f",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

type MyStruct struct {
	fA string
	fB int
}
`
	pass := s.createPass(src)

	var structType *ast.StructType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			structType = st
			return false
		}
		return true
	})

	result := s.checker.checkStructTypeIfEnabled(pass, structType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckInterfaceTypeIfEnabled_Enabled() {
	cfg := &config.SortConfig{
		InterfaceMethods: &config.CheckConfig{
			Enabled: true,
			Prefix:  "M",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

type MyInterface interface {
	MB() int
	MA() string
}
`
	pass := s.createPass(src)

	var interfaceType *ast.InterfaceType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if it, ok := n.(*ast.InterfaceType); ok {
			interfaceType = it
			return false
		}
		return true
	})

	result := s.checker.checkInterfaceTypeIfEnabled(pass, interfaceType)
	s.False(result)

	diagnostics := s.getDiagnostics(pass)
	s.Len(diagnostics, 1)
	s.Contains(diagnostics[0].Message, "interface methods are not sorted")
}

func (s *CheckerMethodsTestSuite) TestCheckInterfaceTypeIfEnabled_Disabled() {
	cfg := &config.SortConfig{
		InterfaceMethods: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

type MyInterface interface {
	MB() int
	MA() string
}
`
	pass := s.createPass(src)

	var interfaceType *ast.InterfaceType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if it, ok := n.(*ast.InterfaceType); ok {
			interfaceType = it
			return false
		}
		return true
	})

	result := s.checker.checkInterfaceTypeIfEnabled(pass, interfaceType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckCallExprIfEnabled_Enabled() {
	cfg := &config.SortConfig{
		VariadicArgs: &config.CheckConfig{
			Enabled: true,
			Prefix:  "",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

func myVariadicFunc(format string, args ...interface{}) {}

func test() {
	myVariadicFunc("test", "b", "a")
}
`
	pass := s.createPass(src)

	var callExpr *ast.CallExpr
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			if ident, ok := ce.Fun.(*ast.Ident); ok && ident.Name == "myVariadicFunc" {
				callExpr = ce
				return false
			}
		}
		return true
	})

	result := s.checker.checkCallExprIfEnabled(pass, callExpr)
	s.False(result)

	diagnostics := s.getDiagnostics(pass)
	s.Len(diagnostics, 1)
	s.Contains(diagnostics[0].Message, "variadic arguments are not sorted")
}

func (s *CheckerMethodsTestSuite) TestCheckCallExprIfEnabled_Disabled() {
	cfg := &config.SortConfig{
		VariadicArgs: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

func myVariadicFunc(format string, args ...interface{}) {}

func test() {
	myVariadicFunc("test", "b", "a")
}
`
	pass := s.createPass(src)

	var callExpr *ast.CallExpr
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			if ident, ok := ce.Fun.(*ast.Ident); ok && ident.Name == "myVariadicFunc" {
				callExpr = ce
				return false
			}
		}
		return true
	})

	result := s.checker.checkCallExprIfEnabled(pass, callExpr)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckCompositeLitIfEnabled_Enabled() {
	cfg := &config.SortConfig{
		MapKeys: &config.CheckConfig{
			Enabled: true,
			Prefix:  "",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

var m = map[string]int{
	"b": 2,
	"a": 1,
}
`
	pass := s.createPass(src)

	var compositeLit *ast.CompositeLit
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			compositeLit = cl
			return false
		}
		return true
	})

	result := s.checker.checkCompositeLitIfEnabled(pass, compositeLit)
	s.False(result)

	diagnostics := s.getDiagnostics(pass)
	s.Len(diagnostics, 1)
	s.Contains(diagnostics[0].Message, "map keys are not sorted")
}

func (s *CheckerMethodsTestSuite) TestCheckCompositeLitIfEnabled_Disabled() {
	cfg := &config.SortConfig{
		MapKeys: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

var m = map[string]int{
	"b": 2,
	"a": 1,
}
`
	pass := s.createPass(src)

	var compositeLit *ast.CompositeLit
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			compositeLit = cl
			return false
		}
		return true
	})

	result := s.checker.checkCompositeLitIfEnabled(pass, compositeLit)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckGenDeclIfEnabled_GlobalPrefix() {
	cfg := &config.SortConfig{
		Constants: &config.CheckConfig{
			Enabled: true,
			Prefix:  "",
		},
		GlobalPrefix: "g",
		IgnoreGroups: false,
	}

	src := `
package test

const (
	gB = 1
	gA = 2
	NotPrefixed = 3
)
`
	s.testGenDeclSorting(cfg, src, token.CONST, false, "variable/constant declarations are not sorted")
}

func (s *CheckerMethodsTestSuite) TestCheckGenDeclIfEnabled_IgnoreGroups() {
	cfg := &config.SortConfig{
		Constants: &config.CheckConfig{
			Enabled: true,
			Prefix:  "C",
		},
		GlobalPrefix: "",
		IgnoreGroups: true,
	}

	src := `
package test

const (
	CA = 1

	CC = 3
	CB = 2
)
`
	s.testGenDeclSorting(cfg, src, token.CONST, false, "variable/constant declarations are not sorted")
}

func (s *CheckerMethodsTestSuite) TestCheckStructTypeIfEnabled_EmptyFieldList() {
	cfg := &config.SortConfig{
		StructFields: &config.CheckConfig{
			Enabled: true,
			Prefix:  "f",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

type MyStruct struct {
}
`
	pass := s.createPass(src)

	var structType *ast.StructType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			structType = st
			return false
		}
		return true
	})

	result := s.checker.checkStructTypeIfEnabled(pass, structType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckInterfaceTypeIfEnabled_EmptyMethodList() {
	cfg := &config.SortConfig{
		InterfaceMethods: &config.CheckConfig{
			Enabled: true,
			Prefix:  "M",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

type MyInterface interface {
}
`
	pass := s.createPass(src)

	var interfaceType *ast.InterfaceType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if it, ok := n.(*ast.InterfaceType); ok {
			interfaceType = it
			return false
		}
		return true
	})

	result := s.checker.checkInterfaceTypeIfEnabled(pass, interfaceType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckCallExprIfEnabled_NonVariadic() {
	cfg := &config.SortConfig{
		VariadicArgs: &config.CheckConfig{
			Enabled: true,
			Prefix:  "",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	src := `
package test

func nonVariadic(a, b string) {}

func test() {
	nonVariadic("b", "a")
}
`
	pass := s.createPass(src)

	var callExpr *ast.CallExpr
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			if ident, ok := ce.Fun.(*ast.Ident); ok && ident.Name == "nonVariadic" {
				callExpr = ce
				return false
			}
		}
		return true
	})

	result := s.checker.checkCallExprIfEnabled(pass, callExpr)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *CheckerMethodsTestSuite) TestCheckCompositeLitIfEnabled_StructLiteral() {
	cfg := &config.SortConfig{
		MapKeys: &config.CheckConfig{
			Enabled: true,
			Prefix:  "",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.checker = NewChecker(cfg)

	// Test with struct field literals - current implementation checks KeyValueExpr regardless of type
	src := `
package test

type MyStruct struct {
	A int
	B string
}

var s = MyStruct{
	B: "hello",
	A: 42,
}
`
	pass := s.createPass(src)

	var structLiteral *ast.CompositeLit
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			structLiteral = cl
			return false
		}
		return true
	})

	result := s.checker.checkCompositeLitIfEnabled(pass, structLiteral)
	s.False(result, "struct field literals with unsorted keys should return false")
	s.Len(s.getDiagnostics(pass), 1, "should have one diagnostic for unsorted struct fields")

	// Test with map literals separately
	src2 := `
package test

var m = map[string]int{
	"b": 2,
	"a": 1,
}
`
	pass2 := s.createPass(src2)

	var mapLiteral *ast.CompositeLit
	ast.Inspect(pass2.Files[0], func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			mapLiteral = cl
			return false
		}
		return true
	})

	result = s.checker.checkCompositeLitIfEnabled(pass2, mapLiteral)
	s.False(result, "map literal with unsorted keys should return false")
	s.Len(s.getDiagnostics(pass2), 1, "should have one diagnostic for unsorted map keys")
}
