package analyzer

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"sort"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

type extractFunc[T ast.Node] func(pass *analysis.Pass, node T) (string, token.Pos, int)

type Metadata struct {
	Line     int
	Node     ast.Node
	Position token.Pos
	Value    string
}

func extractVariadicArgMetadata(
	pass *analysis.Pass,
	callExpr *ast.CallExpr,
	groupByEmptyLine bool,
) [][]Metadata {

	variadicArgs, ok := extractVariadicArgs(pass, callExpr)
	if !ok {
		return [][]Metadata{}
	}

	return extractMetadata(pass, variadicArgs, extractVariadicArg, groupByEmptyLine)
}

func extractVariadicArgs(pass *analysis.Pass, callExpr *ast.CallExpr) ([]ast.Expr, bool) {
	// To properly extract variadic arguments, we need to:
	// 1. Find the function's type signature
	// 2. Determine which arguments correspond to the variadic parameter

	variadicIndex := findVariadicIndex(pass, callExpr)
	if variadicIndex < 0 || variadicIndex >= len(callExpr.Args) {
		return nil, false
	}

	args := callExpr.Args[variadicIndex:]
	if len(args) == 1 {
		if _, ok := args[0].(*ast.Ellipsis); ok {
			// This is a case like fmt.Println(args...) where args is a slice
			// We can't check sorting for this case since it's a single slice argument
			return nil, false
		}
	}

	return args, true
}

func findVariadicIndex(pass *analysis.Pass, callExpr *ast.CallExpr) int {
	if ident, ok := callExpr.Fun.(*ast.Ident); ok {
		// Direct function call like "Println()"
		obj := pass.TypesInfo.ObjectOf(ident)
		if obj != nil {
			if funcType, ok := obj.Type().(*types.Signature); ok {
				if funcType.Variadic() {
					return funcType.Params().Len() - 1
				}
			}
		}
	}

	if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		// Method call or package function like "fmt.Println()"
		obj := pass.TypesInfo.ObjectOf(selExpr.Sel)
		if obj != nil {
			if funcType, ok := obj.Type().(*types.Signature); ok {
				if funcType.Variadic() {
					return funcType.Params().Len() - 1
				}
			}
		}
	}

	return -1
}

func extractMetadata[T ast.Node](
	pass *analysis.Pass,
	nodes []T,
	extract extractFunc[T],
	ignoreGroups bool,
) [][]Metadata {
	if len(nodes) == 0 {
		return [][]Metadata{}
	}

	var allData []Metadata
	for _, node := range nodes {
		value, pos, line := extract(pass, node)
		allData = append(allData, Metadata{
			Line:     line,
			Node:     node,
			Position: pos,
			Value:    value,
		})
	}

	// If not grouping by empty lines, return all elements in a single group
	if ignoreGroups {
		return [][]Metadata{allData}
	}

	sort.Slice(allData, func(i, j int) bool {
		return allData[i].Line < allData[j].Line
	})

	var result [][]Metadata
	if len(allData) == 0 {
		return result
	}

	currentGroup := []Metadata{allData[0]}

	for i := 1; i < len(allData); i++ {
		// Check if there's at least one empty line between elements
		lineDiff := allData[i].Line - allData[i-1].Line

		if lineDiff > 1 {
			// Empty line detected
			result = append(result, currentGroup)
			currentGroup = []Metadata{allData[i]}
		} else {
			currentGroup = append(currentGroup, allData[i])
		}
	}

	// Add the last group
	if len(currentGroup) > 0 {
		result = append(result, currentGroup)
	}

	return result
}

func extractMapKey(pass *analysis.Pass, node *ast.KeyValueExpr) (string, token.Pos, int) {
	value := getKeyString(node.Key)
	pos := node.Key.Pos()
	line := pass.Fset.File(pos).Line(pos)
	return value, pos, line
}

func extractStructField(pass *analysis.Pass, node *ast.Field) (string, token.Pos, int) {
	var value string
	if len(node.Names) > 0 {
		value = node.Names[0].Name
	} else {
		value = getTypeString(node.Type)
	}
	pos := node.Pos()
	line := pass.Fset.File(pos).Line(pos)
	return value, pos, line
}

func extractGenDecl(pass *analysis.Pass, node *ast.ValueSpec) (string, token.Pos, int) {
	value := node.Names[0].Name
	pos := node.Names[0].Pos()
	line := pass.Fset.File(pos).Line(pos)
	return value, pos, line
}

func extractInterfaceMethod(pass *analysis.Pass, node *ast.Field) (string, token.Pos, int) {
	var value string
	var pos token.Pos
	if len(node.Names) > 0 {
		value = node.Names[0].Name
		pos = node.Names[0].Pos()
	} else {
		value = getTypeString(node.Type)
		pos = node.Type.Pos()
	}
	// pos := node.Pos()
	line := pass.Fset.File(pos).Line(pos)
	return value, pos, line
}

func extractVariadicArg(pass *analysis.Pass, node ast.Expr) (string, token.Pos, int) {
	var value string
	switch arg := node.(type) {
	case *ast.Ident:
		value = arg.Name
	case *ast.BasicLit:
		value = arg.Value
	default:
		// Try to get a string representation for more complex expressions
		var buf bytes.Buffer
		_ = printer.Fprint(&buf, token.NewFileSet(), node)
		value = buf.String()
	}

	pos := node.Pos()
	line := pass.Fset.File(pos).Line(pos)
	return value, pos, line
}

// getTypeString returns a string representation of a type expression.
func getTypeString(expr ast.Expr) string {
	switch typeExpr := expr.(type) {
	case *ast.Ident:
		return typeExpr.Name
	case *ast.SelectorExpr:
		if x, ok := typeExpr.X.(*ast.Ident); ok {
			return x.Name + "." + typeExpr.Sel.Name
		}
	}

	return ""
}

// getKeyString extracts a string representation of a map key.
func getKeyString(expr ast.Expr) string {
	switch exprVal := expr.(type) {
	case *ast.BasicLit:
		unqouted, err := strconv.Unquote(exprVal.Value)
		if err != nil {
			// the string is not “unquotable”, so we need to just return it as is
			return exprVal.Value
		}
		return unqouted
	case *ast.Ident:
		return exprVal.Name
	}

	return ""
}
