package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/tools/go/analysis"
)

type ExtractTestSuite struct {
	suite.Suite
	fset *token.FileSet
	pass *analysis.Pass
}

func (s *ExtractTestSuite) SetupTest() {
	s.fset = token.NewFileSet()
	file := s.fset.AddFile("test.go", -1, 1000)
	file.SetLines([]int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100})

	s.pass = &analysis.Pass{
		Fset: s.fset,
		TypesInfo: &types.Info{
			Uses: make(map[*ast.Ident]types.Object),
			Defs: make(map[*ast.Ident]types.Object),
		},
	}
}

func (s *ExtractTestSuite) TestExtractMapKey() {
	tests := []struct {
		name     string
		key      ast.Expr
		expected string
	}{
		{
			name:     "string literal key",
			key:      &ast.BasicLit{Kind: token.STRING, Value: `"test"`},
			expected: "test",
		},
		{
			name:     "string literal with quotes",
			key:      &ast.BasicLit{Kind: token.STRING, Value: `"hello world"`},
			expected: "hello world",
		},
		{
			name:     "identifier key",
			key:      &ast.Ident{Name: "myKey"},
			expected: "myKey",
		},
		{
			name:     "int literal key",
			key:      &ast.BasicLit{Kind: token.INT, Value: "42"},
			expected: "42",
		},
		{
			name:     "malformed string literal",
			key:      &ast.BasicLit{Kind: token.STRING, Value: "unquoted"},
			expected: "unquoted",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			keyPos := token.Pos(15)

			switch k := tt.key.(type) {
			case *ast.BasicLit:
				k.ValuePos = keyPos
			case *ast.Ident:
				k.NamePos = keyPos
			}

			node := &ast.KeyValueExpr{
				Key:   tt.key,
				Value: &ast.BasicLit{},
			}

			value, pos, line := extractMapKey(s.pass, node)
			s.Assert().Equal(tt.expected, value)
			s.Assert().Equal(keyPos, pos)
			s.Assert().Equal(2, line)
		})
	}
}

func (s *ExtractTestSuite) TestExtractStructField() {
	tests := []struct {
		name     string
		field    *ast.Field
		expected string
	}{
		{
			name: "named field",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "MyField"}},
				Type:  &ast.Ident{Name: "string"},
			},
			expected: "MyField",
		},
		{
			name: "embedded field",
			field: &ast.Field{
				Names: nil,
				Type:  &ast.Ident{Name: "EmbeddedType"},
			},
			expected: "EmbeddedType",
		},
		{
			name: "embedded field with selector",
			field: &ast.Field{
				Names: nil,
				Type: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "pkg"},
					Sel: &ast.Ident{Name: "Type"},
				},
			},
			expected: "pkg.Type",
		},
		{
			name: "multiple names (takes first)",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "First"}, {Name: "Second"}},
				Type:  &ast.Ident{Name: "int"},
			},
			expected: "First",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			setFieldPosition(tt.field, token.Pos(25))
			value, pos, line := extractStructField(s.pass, tt.field)
			s.Require().Equal(tt.expected, value)

			if pos == 0 {
				s.Require().Equal(token.Pos(0), pos)
				s.Require().Equal(0, line)
			} else {
				s.Require().Equal(token.Pos(25), pos)
				s.Require().Equal(3, line)
			}
		})
	}
}

func (s *ExtractTestSuite) TestExtractGenDecl() {
	tests := []struct {
		name     string
		spec     *ast.ValueSpec
		expected string
	}{
		{
			name: "const declaration",
			spec: &ast.ValueSpec{
				Names: []*ast.Ident{{Name: "MyConst", NamePos: token.Pos(35)}},
				Type:  &ast.Ident{Name: "int"},
			},
			expected: "MyConst",
		},
		{
			name: "variable declaration",
			spec: &ast.ValueSpec{
				Names: []*ast.Ident{{Name: "myVar", NamePos: token.Pos(35)}},
				Type:  &ast.Ident{Name: "string"},
			},
			expected: "myVar",
		},
		{
			name: "multiple names (takes first)",
			spec: &ast.ValueSpec{
				Names: []*ast.Ident{
					{Name: "first", NamePos: token.Pos(35)},
					{Name: "second"},
				},
			},
			expected: "first",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			value, pos, line := extractGenDecl(s.pass, tt.spec)
			s.Assert().Equal(tt.expected, value)
			s.Assert().Equal(token.Pos(35), pos)
			s.Assert().Equal(4, line)
		})
	}
}

func (s *ExtractTestSuite) TestExtractInterfaceMethod() {
	tests := []struct {
		name     string
		field    *ast.Field
		expected string
		line     int
	}{
		{
			name: "named method",
			field: &ast.Field{
				Names: []*ast.Ident{{Name: "DoSomething", NamePos: token.Pos(45)}},
				Type: &ast.FuncType{
					Params: &ast.FieldList{},
				},
			},
			expected: "DoSomething",
			line:     5,
		},
		{
			name: "embedded interface",
			field: &ast.Field{
				Names: nil,
				Type:  &ast.Ident{Name: "Reader", NamePos: token.Pos(55)},
			},
			expected: "Reader",
			line:     6,
		},
		{
			name: "embedded interface with selector",
			field: &ast.Field{
				Names: nil,
				Type: &ast.SelectorExpr{
					X:   &ast.Ident{Name: "io"},
					Sel: &ast.Ident{Name: "Writer", NamePos: token.Pos(65)},
				},
			},
			expected: "io.Writer",
			line:     7,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			value, pos, line := extractInterfaceMethod(s.pass, tt.field)
			s.Assert().Equal(tt.expected, value)

			if pos == 0 {
				s.Assert().Equal(0, line)
			} else {
				s.Assert().Equal(tt.line, line)
			}

			if tt.field.Names != nil {
				s.Assert().Equal(tt.field.Names[0].Pos(), pos)
			} else {
				s.Assert().Equal(tt.field.Type.Pos(), pos)
			}
		})
	}
}

func (s *ExtractTestSuite) TestExtractVariadicArg() {
	tests := []struct {
		name     string
		arg      ast.Expr
		expected string
	}{
		{
			name:     "identifier argument",
			arg:      &ast.Ident{Name: "myVar", NamePos: token.Pos(15)},
			expected: "myVar",
		},
		{
			name:     "string literal",
			arg:      &ast.BasicLit{Kind: token.STRING, Value: `"hello"`, ValuePos: token.Pos(15)},
			expected: `"hello"`,
		},
		{
			name:     "int literal",
			arg:      &ast.BasicLit{Kind: token.INT, Value: "42", ValuePos: token.Pos(15)},
			expected: "42",
		},
		{
			name: "complex expression",
			arg: &ast.CallExpr{
				Fun: &ast.Ident{Name: "fmt.Sprintf", NamePos: token.Pos(15)},
				Args: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"%d"`},
					&ast.Ident{Name: "x"},
				},
			},
			expected: "fmt.Sprintf(\"%d\", x)",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			value, pos, line := extractVariadicArg(s.pass, tt.arg)
			s.Assert().Equal(tt.expected, value)

			if pos == 0 {
				s.Assert().Equal(0, pos)
				s.Assert().Equal(0, line)
			} else {
				s.Assert().Equal(token.Pos(15), pos)
				s.Assert().Equal(2, line)
			}
		})
	}
}

func (s *ExtractTestSuite) TestExtractVariadicArgMetadata() {
	code := `
	package test
	import "fmt"
	
	func foo() {
		fmt.Println("a", "b", "c")
		fmt.Println("single")
		fmt.Printf("%d %d", 1, 2)
	}
	
	func varadic(args ...string) {}
	
	func bar() {
		values := []string{"a", "b"}
		fmt.Println(values...)
		varadic("one", "two", "three")
	}
	`

	file := s.fset.AddFile("test.go", -1, len(code))
	var lineOffsets []int
	for i, ch := range code {
		if ch == '\n' {
			lineOffsets = append(lineOffsets, i)
		}
	}
	file.SetLines(lineOffsets)

	callExpr := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "fmt"},
			Sel: &ast.Ident{Name: "Println"},
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: `"a"`, ValuePos: token.Pos(100)},
			&ast.BasicLit{Kind: token.STRING, Value: `"b"`, ValuePos: token.Pos(110)},
			&ast.BasicLit{Kind: token.STRING, Value: `"c"`, ValuePos: token.Pos(120)},
		},
	}

	result := extractVariadicArgMetadata(s.pass, callExpr, false)
	s.Assert().Len(result, 0)

	result = extractVariadicArgMetadata(s.pass, callExpr, true)
	s.Assert().Len(result, 0)
}

func (s *ExtractTestSuite) TestExtractVariadicArgs() {
	ellipsisExpr := &ast.CallExpr{
		Fun: &ast.Ident{Name: "foo"},
		Args: []ast.Expr{
			&ast.Ellipsis{
				Elt: &ast.Ident{Name: "args"},
			},
		},
	}

	args, ok := extractVariadicArgs(s.pass, ellipsisExpr)
	s.Assert().False(ok)
	s.Assert().Nil(args)

	noVariadicExpr := &ast.CallExpr{
		Fun: &ast.Ident{Name: "foo"},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: `"test"`},
		},
	}

	args, ok = extractVariadicArgs(s.pass, noVariadicExpr)
	s.Assert().False(ok)
	s.Assert().Nil(args)
}

func (s *ExtractTestSuite) TestExtractMetadata() {
	s.Run("empty nodes", func() {
		var nodes []testNode

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Assert().Empty(result)

		result = extractMetadata(s.pass, nodes, testExtractFunc, true)
		s.Assert().Empty(result)
	})

	s.Run("single node", func() {
		nodes := []testNode{
			{value: "test1", line: 1},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Assert().Len(result, 1)
		s.Require().Len(result[0], 1)
		s.Require().Equal("test1", result[0][0].Value)
		s.Require().Equal(1, result[0][0].Line)
		s.Require().Equal(token.Pos(1), result[0][0].Position)

		result = extractMetadata(s.pass, nodes, testExtractFunc, true)
		s.Require().Len(result, 1)
		s.Require().Len(result[0], 1)
		s.Require().Equal("test1", result[0][0].Value)
	})

	s.Run("ignore groups", func() {
		nodes := []testNode{
			{value: "test1", line: 1},
			{value: "test2", line: 3},
			{value: "test3", line: 5},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, true)
		s.Require().Len(result, 1)
		s.Require().Len(result[0], 3)
		s.Require().Equal("test1", result[0][0].Value)
		s.Require().Equal("test2", result[0][1].Value)
		s.Require().Equal("test3", result[0][2].Value)
	})

	s.Run("group by empty line", func() {
		nodes := []testNode{
			{value: "test1", line: 1},
			{value: "test2", line: 2},
			{value: "test3", line: 4},
			{value: "test4", line: 5},
			{value: "test5", line: 7},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Require().Len(result, 3)

		s.Require().Len(result[0], 2)
		s.Assert().Equal("test1", result[0][0].Value)
		s.Assert().Equal("test2", result[0][1].Value)

		s.Require().Len(result[1], 2)
		s.Assert().Equal("test3", result[1][0].Value)
		s.Assert().Equal("test4", result[1][1].Value)

		s.Require().Len(result[2], 1)
		s.Assert().Equal("test5", result[2][0].Value)
	})

	s.Run("consecutive empty lines", func() {
		nodes := []testNode{
			{value: "test1", line: 1},
			{value: "test2", line: 4},
			{value: "test3", line: 7},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Require().Len(result, 3)

		s.Require().Len(result[0], 1)
		s.Assert().Equal("test1", result[0][0].Value)

		s.Require().Len(result[1], 1)
		s.Assert().Equal("test2", result[1][0].Value)

		s.Require().Len(result[2], 1)
		s.Assert().Equal("test3", result[2][0].Value)
	})

	s.Run("unsorted nodes", func() {
		nodes := []testNode{
			{value: "test3", line: 5},
			{value: "test1", line: 1},
			{value: "test2", line: 2},
			{value: "test4", line: 7},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Require().Len(result, 3)

		s.Require().Len(result[0], 2)
		s.Assert().Equal("test1", result[0][0].Value)
		s.Assert().Equal(1, result[0][0].Line)
		s.Assert().Equal("test2", result[0][1].Value)
		s.Assert().Equal(2, result[0][1].Line)

		s.Require().Len(result[1], 1)
		s.Assert().Equal("test3", result[1][0].Value)
		s.Assert().Equal(5, result[1][0].Line)

		s.Require().Len(result[2], 1)
		s.Assert().Equal("test4", result[2][0].Value)
		s.Assert().Equal(7, result[2][0].Line)
	})

	s.Run("single line gap", func() {
		nodes := []testNode{
			{value: "test1", line: 1},
			{value: "test2", line: 2},
			{value: "test3", line: 3},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Require().Len(result, 1)
		s.Require().Len(result[0], 3)
		s.Assert().Equal("test1", result[0][0].Value)
		s.Assert().Equal("test2", result[0][1].Value)
		s.Assert().Equal("test3", result[0][2].Value)
	})

	s.Run("node preservation", func() {
		type customNode struct {
			testNode
			extra string
		}

		customExtract := func(pass *analysis.Pass, node customNode) (string, token.Pos, int) {
			return node.value, node.Pos(), node.line
		}

		nodes := []customNode{
			{testNode: testNode{value: "test1", line: 1}, extra: "extra1"},
			{testNode: testNode{value: "test2", line: 3}, extra: "extra2"},
		}

		result := extractMetadata(s.pass, nodes, customExtract, false)
		s.Require().Len(result, 2)

		s.Assert().Equal("extra1", result[0][0].Node.(customNode).extra)
		s.Assert().Equal("extra2", result[1][0].Node.(customNode).extra)
	})

	s.Run("two line gap", func() {
		nodes := []testNode{
			{value: "test1", line: 1},
			{value: "test2", line: 4},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Require().Len(result, 2)

		s.Assert().Len(result[0], 1)
		s.Assert().Equal("test1", result[0][0].Value)

		s.Assert().Len(result[1], 1)
		s.Assert().Equal("test2", result[1][0].Value)
	})

	s.Run("mixed gaps with ignore groups", func() {
		nodes := []testNode{
			{value: "test1", line: 1},
			{value: "test2", line: 3},
			{value: "test3", line: 4},
			{value: "test4", line: 7},
			{value: "test5", line: 8},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, true)
		s.Require().Len(result, 1)
		s.Assert().Len(result[0], 5)

		for i, expected := range []string{"test1", "test2", "test3", "test4", "test5"} {
			s.Assert().Equal(expected, result[0][i].Value)
		}
	})

	s.Run("large gaps", func() {
		nodes := []testNode{
			{value: "test1", line: 1},
			{value: "test2", line: 10},
			{value: "test3", line: 20},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Require().Len(result, 3)

		for i, expected := range []string{"test1", "test2", "test3"} {
			s.Assert().Len(result[i], 1)
			s.Assert().Equal(expected, result[i][0].Value)
		}
	})

	s.Run("all same line", func() {
		nodes := []testNode{
			{value: "test1", line: 5},
			{value: "test2", line: 5},
			{value: "test3", line: 5},
		}

		result := extractMetadata(s.pass, nodes, testExtractFunc, false)
		s.Require().Len(result, 1)
		s.Assert().Len(result[0], 3)

		for i, expected := range []string{"test1", "test2", "test3"} {
			s.Assert().Equal(expected, result[0][i].Value)
			s.Assert().Equal(5, result[0][i].Line)
		}
	})
}

func setFieldPosition(field *ast.Field, pos token.Pos) {
	if field.Names != nil {
		field.Names[0].NamePos = pos
	} else if ident, ok := field.Type.(*ast.Ident); ok {
		ident.NamePos = pos
	} else if sel, ok := field.Type.(*ast.SelectorExpr); ok {
		sel.Sel.NamePos = pos
	}
}

type testNode struct {
	value string
	line  int
}

func (n testNode) Pos() token.Pos {
	return token.Pos(n.line)
}

func (n testNode) End() token.Pos {
	return token.Pos(n.line + 1)
}

func testExtractFunc(pass *analysis.Pass, node testNode) (string, token.Pos, int) {
	return node.value, node.Pos(), node.line
}

func TestExtractTestSuite(t *testing.T) {
	suite.Run(t, new(ExtractTestSuite))
}
