package analyzer_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"go.tomakado.io/sortir/internal/analyzer"
	"go.tomakado.io/sortir/internal/config"
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

		cfg := config.New()
		cfg.VariadicArgs.Enabled = true
		a := analyzer.New().WithConfig(cfg)

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

		cfg := config.New()
		cfg.GlobalPrefix = "Pref"
		a := analyzer.New().WithConfig(cfg)

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/global")
	})

	t.Run("constants", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		cfg.GlobalPrefix = ""
		cfg.Constants.Prefix = "Pref"
		a := analyzer.New().WithConfig(cfg)

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/constants")
	})

	t.Run("interface methods", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		cfg.InterfaceMethods.Prefix = "Pref"
		a := analyzer.New().WithConfig(cfg)

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/interfaces")
	})

	t.Run("structs", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		cfg.StructFields.Prefix = "Pref"
		a := analyzer.New().WithConfig(cfg)

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/structs")
	})

	t.Run("maps", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		cfg.MapKeys.Prefix = "Pref"
		a := analyzer.New().WithConfig(cfg)

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/maps")
	})

	t.Run("variadic", func(t *testing.T) {
		t.Parallel()

		cfg := config.New()
		cfg.VariadicArgs.Enabled = true
		cfg.VariadicArgs.Prefix = "Pref"
		a := analyzer.New().WithConfig(cfg)

		analysistest.Run(t, testdata, a.Analyzer(), "filterprefix/variadic")
	})
}

type mockNode struct{}

func (m *mockNode) Pos() token.Pos { return token.NoPos }
func (m *mockNode) End() token.Pos { return token.NoPos }

type AnalyzerTestSuite struct {
	suite.Suite
}

func (s *AnalyzerTestSuite) TestCheckElementsSorted_AllSortedSingleGroup() {
	s.T().Parallel()

	groups := [][]analyzer.Metadata{
		{
			{Value: "a", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "b", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "c", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	a := analyzer.New()

	result := a.CheckElementsSorted(pass, groups, "", "test message")
	s.Require().True(result)
	s.Require().Empty(reported)
}

func (s *AnalyzerTestSuite) TestCheckElementsSorted_NotSortedSingleGroup() {
	s.T().Parallel()

	groups := [][]analyzer.Metadata{
		{
			{Value: "b", Position: token.Pos(1), Line: 1, Node: &mockNode{}},
			{Value: "a", Position: token.Pos(2), Line: 2, Node: &mockNode{}},
			{Value: "c", Position: token.Pos(3), Line: 3, Node: &mockNode{}},
		},
	}

	reported := []analysis.Diagnostic{}
	pass := &analysis.Pass{
		Report: func(d analysis.Diagnostic) {
			reported = append(reported, d)
		},
	}

	a := analyzer.New()

	result := a.CheckElementsSorted(pass, groups, "", "test message")
	s.Require().False(result)
	s.Require().Len(reported, 1)
	s.Require().Equal(token.Pos(2), reported[0].Pos)
	s.Require().Equal("test message", reported[0].Message)
}

type AnalyzerMethodsTestSuite struct {
	suite.Suite
	fset     *token.FileSet
	analyzer *analyzer.Analyzer
}

func (s *AnalyzerMethodsTestSuite) SetupTest() {
	s.fset = token.NewFileSet()
}

var diagnosticsAnalyzer = &analysis.Analyzer{Name: "diagnostics"}

func (s *AnalyzerMethodsTestSuite) createPass(src string) *analysis.Pass {
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

	pass.Report = func(d analysis.Diagnostic) {
		diagnostics = append(diagnostics, d)
	}

	pass.ResultOf = make(map[*analysis.Analyzer]any)
	pass.ResultOf[diagnosticsAnalyzer] = &diagnostics

	return pass
}

func (s *AnalyzerMethodsTestSuite) getDiagnostics(pass *analysis.Pass) []analysis.Diagnostic {
	if result, ok := pass.ResultOf[diagnosticsAnalyzer].(*[]analysis.Diagnostic); ok {
		return *result
	}
	return nil
}

type testParams struct {
	cfg          *config.SortConfig
	src          string
	errorMessage string
}

func (s *AnalyzerMethodsTestSuite) testStructTypeSorting(params testParams, shouldPass bool) {
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(params.cfg)
	pass := s.createPass(params.src)

	var structType *ast.StructType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			structType = st
			return false
		}
		return true
	})

	result := s.analyzer.CheckStructType(pass, structType)
	s.Equal(shouldPass, result)

	diagnostics := s.getDiagnostics(pass)
	if shouldPass {
		s.Empty(diagnostics)
	} else {
		s.Len(diagnostics, 1)
		s.Contains(diagnostics[0].Message, params.errorMessage)
	}
}

func (s *AnalyzerMethodsTestSuite) testInterfaceTypeSorting(params testParams, shouldPass bool) {
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(params.cfg)
	pass := s.createPass(params.src)

	var interfaceType *ast.InterfaceType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if it, ok := n.(*ast.InterfaceType); ok {
			interfaceType = it
			return false
		}
		return true
	})

	result := s.analyzer.CheckInterfaceType(pass, interfaceType)
	s.Equal(shouldPass, result)

	diagnostics := s.getDiagnostics(pass)
	if shouldPass {
		s.Empty(diagnostics)
	} else {
		s.Len(diagnostics, 1)
		s.Contains(diagnostics[0].Message, params.errorMessage)
	}
}

func (s *AnalyzerMethodsTestSuite) testCompositeLitSorting(params testParams, shouldPass bool) {
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(params.cfg)
	pass := s.createPass(params.src)

	var compositeLit *ast.CompositeLit
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			compositeLit = cl
			return false
		}
		return true
	})

	result := s.analyzer.CheckCompositeLit(pass, compositeLit)
	s.Equal(shouldPass, result)

	diagnostics := s.getDiagnostics(pass)
	if shouldPass {
		s.Empty(diagnostics)
	} else {
		s.Len(diagnostics, 1)
		s.Contains(diagnostics[0].Message, params.errorMessage)
	}
}

func (s *AnalyzerMethodsTestSuite) testGenDeclSorting(cfg *config.SortConfig, src string, expectedToken token.Token, shouldPass bool, expectedError string) {
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)
	pass := s.createPass(src)

	var genDecl *ast.GenDecl
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if g, ok := n.(*ast.GenDecl); ok && g.Tok == expectedToken {
			genDecl = g
			return false
		}
		return true
	})

	result := s.analyzer.CheckGenDecl(pass, genDecl)
	s.Equal(shouldPass, result)

	diagnostics := s.getDiagnostics(pass)
	if shouldPass {
		s.Empty(diagnostics)
	} else {
		s.Len(diagnostics, 1)
		s.Contains(diagnostics[0].Message, expectedError)
	}
}

func TestAnalyzerTestSuite(t *testing.T) {
	suite.Run(t, new(AnalyzerTestSuite))
}

func (s *AnalyzerMethodsTestSuite) TestCheckGenDecl_ConstantsEnabled() {
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

func (s *AnalyzerMethodsTestSuite) TestCheckGenDecl_ConstantsDisabled() {
	cfg := &config.SortConfig{
		Constants: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckGenDecl(pass, genDecl)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckGenDecl_VariablesEnabled() {
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

func (s *AnalyzerMethodsTestSuite) TestCheckGenDecl_SortedVariables() {
	cfg := &config.SortConfig{
		Variables: &config.CheckConfig{
			Enabled: true,
			Prefix:  "v",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckGenDecl(pass, genDecl)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckStructType_Enabled() {
	s.testStructTypeSorting(testParams{
		cfg: &config.SortConfig{
			StructFields: &config.CheckConfig{
				Enabled: true,
				Prefix:  "f",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		},
		src: `
package test

type MyStruct struct {
	fB int
	fA string
}
`,
		errorMessage: "struct fields are not sorted",
	}, false)
}

func (s *AnalyzerMethodsTestSuite) TestCheckStructType_Disabled() {
	cfg := &config.SortConfig{
		StructFields: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckStructType(pass, structType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckStructType_Sorted() {
	cfg := &config.SortConfig{
		StructFields: &config.CheckConfig{
			Enabled: true,
			Prefix:  "f",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckStructType(pass, structType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckInterfaceType_Enabled() {
	s.testInterfaceTypeSorting(testParams{
		cfg: &config.SortConfig{
			InterfaceMethods: &config.CheckConfig{
				Enabled: true,
				Prefix:  "M",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		},
		src: `
package test

type MyInterface interface {
	MB() int
	MA() string
}
`,
		errorMessage: "interface methods are not sorted",
	}, false)
}

func (s *AnalyzerMethodsTestSuite) TestCheckInterfaceType_Disabled() {
	cfg := &config.SortConfig{
		InterfaceMethods: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckInterfaceType(pass, interfaceType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckCallExpr_Enabled() {
	cfg := &config.SortConfig{
		VariadicArgs: &config.CheckConfig{
			Enabled: true,
			Prefix:  "",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckCallExpr(pass, callExpr)
	s.False(result)

	diagnostics := s.getDiagnostics(pass)
	s.Len(diagnostics, 1)
	s.Contains(diagnostics[0].Message, "variadic arguments are not sorted")
}

func (s *AnalyzerMethodsTestSuite) TestCheckCallExpr_Disabled() {
	cfg := &config.SortConfig{
		VariadicArgs: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckCallExpr(pass, callExpr)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckCompositeLit_Enabled() {
	s.testCompositeLitSorting(testParams{
		cfg: &config.SortConfig{
			MapKeys: &config.CheckConfig{
				Enabled: true,
				Prefix:  "",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		},
		src: `
package test

var m = map[string]int{
	"b": 2,
	"a": 1,
}
`,
		errorMessage: "composite literal elements are not sorted",
	}, false)
}

func (s *AnalyzerMethodsTestSuite) TestCheckCompositeLit_Disabled() {
	cfg := &config.SortConfig{
		MapKeys: &config.CheckConfig{
			Enabled: false,
		},
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckCompositeLit(pass, compositeLit)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckGenDecl_GlobalPrefix() {
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

func (s *AnalyzerMethodsTestSuite) TestCheckGenDecl_IgnoreGroups() {
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

func (s *AnalyzerMethodsTestSuite) TestCheckStructType_EmptyFieldList() {
	cfg := &config.SortConfig{
		StructFields: &config.CheckConfig{
			Enabled: true,
			Prefix:  "f",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckStructType(pass, structType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckInterfaceType_EmptyMethodList() {
	cfg := &config.SortConfig{
		InterfaceMethods: &config.CheckConfig{
			Enabled: true,
			Prefix:  "M",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckInterfaceType(pass, interfaceType)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckCallExpr_NonVariadic() {
	cfg := &config.SortConfig{
		VariadicArgs: &config.CheckConfig{
			Enabled: true,
			Prefix:  "",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckCallExpr(pass, callExpr)
	s.True(result)
	s.Empty(s.getDiagnostics(pass))
}

func (s *AnalyzerMethodsTestSuite) TestCheckCompositeLit_StructLiteral() {
	cfg := &config.SortConfig{
		MapKeys: &config.CheckConfig{
			Enabled: true,
			Prefix:  "",
		},
		GlobalPrefix: "",
		IgnoreGroups: false,
	}
	s.analyzer = analyzer.New()
	s.analyzer = s.analyzer.WithConfig(cfg)

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

	result := s.analyzer.CheckCompositeLit(pass, structLiteral)
	s.False(result, "struct field literals with unsorted keys should return false")
	s.Len(s.getDiagnostics(pass), 1, "should have one diagnostic for unsorted struct fields")

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

	result = s.analyzer.CheckCompositeLit(pass2, mapLiteral)
	s.False(result, "map literal with unsorted keys should return false")
	s.Len(s.getDiagnostics(pass2), 1, "should have one diagnostic for unsorted map keys")
}

func TestAnalyzerMethods(t *testing.T) {
	suite.Run(t, new(AnalyzerMethodsTestSuite))
}
