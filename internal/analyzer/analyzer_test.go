package analyzer_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
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

var diagnosticsAnalyzer = &analysis.Analyzer{Name: "diagnostics"}

func createPass(t *testing.T, src string) *analysis.Pass {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	var diagnostics []analysis.Diagnostic
	conf := &types.Config{}

	info := &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	pkg, err := conf.Check("test", fset, []*ast.File{file}, info)
	require.NoError(t, err)

	pass := &analysis.Pass{
		Files: []*ast.File{file},
		Fset:  fset,
		Pkg:   pkg,
		Report: func(d analysis.Diagnostic) {
			diagnostics = append(diagnostics, d)
		},
		ResultOf:  make(map[*analysis.Analyzer]any),
		TypesInfo: info,
	}

	pass.Report = func(d analysis.Diagnostic) {
		diagnostics = append(diagnostics, d)
	}

	pass.ResultOf = make(map[*analysis.Analyzer]any)
	pass.ResultOf[diagnosticsAnalyzer] = &diagnostics

	return pass
}

func getDiagnostics(pass *analysis.Pass) []analysis.Diagnostic {
	if result, ok := pass.ResultOf[diagnosticsAnalyzer].(*[]analysis.Diagnostic); ok {
		return *result
	}
	return nil
}

type testParams struct {
	cfg          *config.SortConfig
	errorMessage string
	src          string
}

func testStructTypeSorting(t *testing.T, params testParams, shouldPass bool) {
	a := analyzer.New().WithConfig(params.cfg)
	pass := createPass(t, params.src)

	var structType *ast.StructType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if st, ok := n.(*ast.StructType); ok {
			structType = st
			return false
		}
		return true
	})

	result := a.CheckStructType(pass, structType)
	require.Equal(t, shouldPass, result)

	diagnostics := getDiagnostics(pass)
	if shouldPass {
		require.Empty(t, diagnostics)
	} else {
		require.Len(t, diagnostics, 1)
		require.Contains(t, diagnostics[0].Message, params.errorMessage)
	}
}

func testInterfaceTypeSorting(t *testing.T, params testParams, shouldPass bool) {
	a := analyzer.New().WithConfig(params.cfg)
	pass := createPass(t, params.src)

	var interfaceType *ast.InterfaceType
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if it, ok := n.(*ast.InterfaceType); ok {
			interfaceType = it
			return false
		}
		return true
	})

	result := a.CheckInterfaceType(pass, interfaceType)
	require.Equal(t, shouldPass, result)

	diagnostics := getDiagnostics(pass)
	if shouldPass {
		require.Empty(t, diagnostics)
	} else {
		require.Len(t, diagnostics, 1)
		require.Contains(t, diagnostics[0].Message, params.errorMessage)
	}
}

func testCompositeLitSorting(t *testing.T, params testParams, shouldPass bool) {
	a := analyzer.New().WithConfig(params.cfg)
	pass := createPass(t, params.src)

	var compositeLit *ast.CompositeLit
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			compositeLit = cl
			return false
		}
		return true
	})

	result := a.CheckCompositeLit(pass, compositeLit)
	require.Equal(t, shouldPass, result)

	diagnostics := getDiagnostics(pass)
	if shouldPass {
		require.Empty(t, diagnostics)
	} else {
		require.Len(t, diagnostics, 1)
		require.Contains(t, diagnostics[0].Message, params.errorMessage)
	}
}

func testGenDeclSorting(t *testing.T, cfg *config.SortConfig, src string, expectedToken token.Token, shouldPass bool, expectedError string) {
	a := analyzer.New().WithConfig(cfg)
	pass := createPass(t, src)

	var genDecl *ast.GenDecl
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if g, ok := n.(*ast.GenDecl); ok && g.Tok == expectedToken {
			genDecl = g
			return false
		}
		return true
	})

	result := a.CheckGenDecl(pass, genDecl)
	require.Equal(t, shouldPass, result)

	diagnostics := getDiagnostics(pass)
	if shouldPass {
		require.Empty(t, diagnostics)
	} else {
		require.Len(t, diagnostics, 1)
		require.Contains(t, diagnostics[0].Message, expectedError)
	}
}

func TestCheckElementsSorted(t *testing.T) {
	t.Parallel()

	t.Run("all sorted single group", func(t *testing.T) {
		t.Parallel()

		groups := [][]analyzer.Metadata{
			{
				{Line: 1,
					Node:     &mockNode{},
					Position: token.Pos(1),
					Value:    "a"},
				{Line: 2,
					Node:     &mockNode{},
					Position: token.Pos(2),
					Value:    "b"},
				{Line: 3,
					Node:     &mockNode{},
					Position: token.Pos(3),
					Value:    "c"},
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
		require.True(t, result)
		require.Empty(t, reported)
	})

	t.Run("not sorted single group", func(t *testing.T) {
		t.Parallel()

		groups := [][]analyzer.Metadata{
			{
				{Line: 1,
					Node:     &mockNode{},
					Position: token.Pos(1),
					Value:    "b"},
				{Line: 2,
					Node:     &mockNode{},
					Position: token.Pos(2),
					Value:    "a"},
				{Line: 3,
					Node:     &mockNode{},
					Position: token.Pos(3),
					Value:    "c"},
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
		require.False(t, result)
		require.Len(t, reported, 1)
		require.Equal(t, token.Pos(2), reported[0].Pos)
		require.Equal(t, "test message", reported[0].Message)
	})
}

func TestCheckGenDecl(t *testing.T) {
	t.Parallel()

	t.Run("constants enabled", func(t *testing.T) {
		t.Parallel()

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
		testGenDeclSorting(t, cfg, src, token.CONST, false, "variable/constant declarations are not sorted")
	})

	t.Run("constants disabled", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			Constants: &config.CheckConfig{
				Enabled: false,
			},
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

const (
	CB = 1
	CA = 2
)
`
		pass := createPass(t, src)

		var genDecl *ast.GenDecl
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if g, ok := n.(*ast.GenDecl); ok && g.Tok == token.CONST {
				genDecl = g
				return false
			}
			return true
		})

		result := a.CheckGenDecl(pass, genDecl)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("variables enabled", func(t *testing.T) {
		t.Parallel()

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
		testGenDeclSorting(t, cfg, src, token.VAR, false, "variable/constant declarations are not sorted")
	})

	t.Run("sorted variables", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			Variables: &config.CheckConfig{
				Enabled: true,
				Prefix:  "v",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

var (
	vA = 1
	vB = 2
)
`
		pass := createPass(t, src)

		var genDecl *ast.GenDecl
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if g, ok := n.(*ast.GenDecl); ok && g.Tok == token.VAR {
				genDecl = g
				return false
			}
			return true
		})

		result := a.CheckGenDecl(pass, genDecl)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("global prefix", func(t *testing.T) {
		t.Parallel()

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
		testGenDeclSorting(t, cfg, src, token.CONST, false, "variable/constant declarations are not sorted")
	})

	t.Run("ignore groups", func(t *testing.T) {
		t.Parallel()

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
		testGenDeclSorting(t, cfg, src, token.CONST, false, "variable/constant declarations are not sorted")
	})

	t.Run("respect groups", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			Constants: &config.CheckConfig{
				Enabled: true,
				Prefix:  "",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		}

		src := `
package test

const (
	A = 1
	B = 2

	Y = 3
	X = 4
)
`
		testGenDeclSorting(t, cfg, src, token.CONST, false, "variable/constant declarations are not sorted")
	})

	t.Run("ignore groups variables", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			Variables: &config.CheckConfig{
				Enabled: true,
				Prefix:  "",
			},
			GlobalPrefix: "",
			IgnoreGroups: true,
		}

		src := `
package test

var (
	a = 1

	c = 3
	b = 2
)
`
		testGenDeclSorting(t, cfg, src, token.VAR, false, "variable/constant declarations are not sorted")
	})
}

func TestCheckStructType(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		testStructTypeSorting(t, testParams{
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
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			StructFields: &config.CheckConfig{
				Enabled: false,
			},
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

type MyStruct struct {
	fB int
	fA string
}
`
		pass := createPass(t, src)

		var structType *ast.StructType
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if st, ok := n.(*ast.StructType); ok {
				structType = st
				return false
			}
			return true
		})

		result := a.CheckStructType(pass, structType)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("sorted", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			StructFields: &config.CheckConfig{
				Enabled: true,
				Prefix:  "f",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

type MyStruct struct {
	fA string
	fB int
}
`
		pass := createPass(t, src)

		var structType *ast.StructType
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if st, ok := n.(*ast.StructType); ok {
				structType = st
				return false
			}
			return true
		})

		result := a.CheckStructType(pass, structType)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("empty field list", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			StructFields: &config.CheckConfig{
				Enabled: true,
				Prefix:  "f",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

type MyStruct struct {
}
`
		pass := createPass(t, src)

		var structType *ast.StructType
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if st, ok := n.(*ast.StructType); ok {
				structType = st
				return false
			}
			return true
		})

		result := a.CheckStructType(pass, structType)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("ignore groups", func(t *testing.T) {
		t.Parallel()

		testStructTypeSorting(t, testParams{
			cfg: &config.SortConfig{
				StructFields: &config.CheckConfig{
					Enabled: true,
					Prefix:  "",
				},
				GlobalPrefix: "",
				IgnoreGroups: true,
			},
			src: `
package test

type MyStruct struct {
	A string

	C int
	B bool
}
`,
			errorMessage: "struct fields are not sorted",
		}, false)
	})

	t.Run("respect groups", func(t *testing.T) {
		t.Parallel()

		testStructTypeSorting(t, testParams{
			cfg: &config.SortConfig{
				StructFields: &config.CheckConfig{
					Enabled: true,
					Prefix:  "",
				},
				GlobalPrefix: "",
				IgnoreGroups: false,
			},
			src: `
package test

type MyStruct struct {
	A string
	B bool

	Y int
	X string
}
`,
			errorMessage: "struct fields are not sorted",
		}, false)
	})
}

func TestCheckInterfaceType(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		testInterfaceTypeSorting(t, testParams{
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
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			InterfaceMethods: &config.CheckConfig{
				Enabled: false,
			},
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

type MyInterface interface {
	MB() int
	MA() string
}
`
		pass := createPass(t, src)

		var interfaceType *ast.InterfaceType
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if it, ok := n.(*ast.InterfaceType); ok {
				interfaceType = it
				return false
			}
			return true
		})

		result := a.CheckInterfaceType(pass, interfaceType)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("empty method list", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			InterfaceMethods: &config.CheckConfig{
				Enabled: true,
				Prefix:  "M",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

type MyInterface interface {
}
`
		pass := createPass(t, src)

		var interfaceType *ast.InterfaceType
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if it, ok := n.(*ast.InterfaceType); ok {
				interfaceType = it
				return false
			}
			return true
		})

		result := a.CheckInterfaceType(pass, interfaceType)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("ignore groups", func(t *testing.T) {
		t.Parallel()

		testInterfaceTypeSorting(t, testParams{
			cfg: &config.SortConfig{
				InterfaceMethods: &config.CheckConfig{
					Enabled: true,
					Prefix:  "",
				},
				GlobalPrefix: "",
				IgnoreGroups: true,
			},
			src: `
package test

type MyInterface interface {
	A() string

	C() int
	B() bool
}
`,
			errorMessage: "interface methods are not sorted",
		}, false)
	})

	t.Run("respect groups", func(t *testing.T) {
		t.Parallel()

		testInterfaceTypeSorting(t, testParams{
			cfg: &config.SortConfig{
				InterfaceMethods: &config.CheckConfig{
					Enabled: true,
					Prefix:  "",
				},
				GlobalPrefix: "",
				IgnoreGroups: false,
			},
			src: `
package test

type MyInterface interface {
	A() string
	B() bool

	Y() int
	X() string
}
`,
			errorMessage: "interface methods are not sorted",
		}, false)
	})
}

func TestCheckCallExpr(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			VariadicArgs: &config.CheckConfig{
				Enabled: true,
				Prefix:  "",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

func myVariadicFunc(format string, args ...interface{}) {}

func test() {
	myVariadicFunc("test", "b", "a")
}
`
		pass := createPass(t, src)

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

		result := a.CheckCallExpr(pass, callExpr)
		require.False(t, result)

		diagnostics := getDiagnostics(pass)
		require.Len(t, diagnostics, 1)
		require.Contains(t, diagnostics[0].Message, "variadic arguments are not sorted")
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			VariadicArgs: &config.CheckConfig{
				Enabled: false,
			},
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

func myVariadicFunc(format string, args ...interface{}) {}

func test() {
	myVariadicFunc("test", "b", "a")
}
`
		pass := createPass(t, src)

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

		result := a.CheckCallExpr(pass, callExpr)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("non variadic", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			VariadicArgs: &config.CheckConfig{
				Enabled: true,
				Prefix:  "",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

func nonVariadic(a, b string) {}

func test() {
	nonVariadic("b", "a")
}
`
		pass := createPass(t, src)

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

		result := a.CheckCallExpr(pass, callExpr)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})
}

func TestCheckCompositeLit(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()

		testCompositeLitSorting(t, testParams{
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
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			MapKeys: &config.CheckConfig{
				Enabled: false,
			},
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

		src := `
package test

var m = map[string]int{
	"b": 2,
	"a": 1,
}
`
		pass := createPass(t, src)

		var compositeLit *ast.CompositeLit
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if cl, ok := n.(*ast.CompositeLit); ok {
				compositeLit = cl
				return false
			}
			return true
		})

		result := a.CheckCompositeLit(pass, compositeLit)
		require.True(t, result)
		require.Empty(t, getDiagnostics(pass))
	})

	t.Run("struct literal", func(t *testing.T) {
		t.Parallel()

		cfg := &config.SortConfig{
			MapKeys: &config.CheckConfig{
				Enabled: true,
				Prefix:  "",
			},
			GlobalPrefix: "",
			IgnoreGroups: false,
		}
		a := analyzer.New().WithConfig(cfg)

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
		pass := createPass(t, src)

		var structLiteral *ast.CompositeLit
		ast.Inspect(pass.Files[0], func(n ast.Node) bool {
			if cl, ok := n.(*ast.CompositeLit); ok {
				structLiteral = cl
				return false
			}
			return true
		})

		result := a.CheckCompositeLit(pass, structLiteral)
		require.False(t, result, "struct field literals with unsorted keys should return false")
		require.Len(t, getDiagnostics(pass), 1, "should have one diagnostic for unsorted struct fields")

		src2 := `
package test

var m = map[string]int{
	"b": 2,
	"a": 1,
}
`
		pass2 := createPass(t, src2)

		var mapLiteral *ast.CompositeLit
		ast.Inspect(pass2.Files[0], func(n ast.Node) bool {
			if cl, ok := n.(*ast.CompositeLit); ok {
				mapLiteral = cl
				return false
			}
			return true
		})

		result = a.CheckCompositeLit(pass2, mapLiteral)
		require.False(t, result, "map literal with unsorted keys should return false")
		require.Len(t, getDiagnostics(pass2), 1, "should have one diagnostic for unsorted map keys")
	})
}
