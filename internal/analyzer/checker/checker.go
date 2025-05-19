// Package checker implements the checking logic for sorting of Go code elements.
package checker

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"go.tomakado.io/sortir/internal/config"
)

type Checker struct {
	Config *config.SortConfig
}

func NewChecker(cfg *config.SortConfig) *Checker {
	return &Checker{
		Config: cfg,
	}
}

func (c *Checker) CheckNode(pass *analysis.Pass, node ast.Node) bool {
	switch n := node.(type) {
	case *ast.GenDecl:
		return c.checkGenDecl(pass, n)
	case *ast.StructType:
		return c.checkStructType(pass, n)
	case *ast.InterfaceType:
		return c.checkInterfaceType(pass, n)
	case *ast.CallExpr:
		return c.checkCallExpr(pass, n)
	case *ast.CompositeLit:
		return c.checkCompositeLit(pass, n)
	}
	return true
}

func (c *Checker) checkGenDecl(pass *analysis.Pass, node *ast.GenDecl) bool {
	var prefix string
	switch node.Tok {
	case token.CONST:
		prefix = c.Config.Constants.Prefix
	case token.VAR:
		prefix = c.Config.Variables.Prefix
	}

	isConst := node.Tok == token.CONST
	isVar := node.Tok == token.VAR

	shouldCheck := isConst && c.Config.Constants.Enabled || isVar && c.Config.Variables.Enabled
	if !shouldCheck {
		return true
	}

	valueSpecs := make([]*ast.ValueSpec, 0, len(node.Specs))
	for _, spec := range node.Specs {
		valueSpecs = append(valueSpecs, spec.(*ast.ValueSpec))
	}

	metadata := extractMetadata(pass, valueSpecs, extractGenDecl, c.Config.IgnoreGroups)
	return checkElementsSorted(
		pass,
		metadata,
		prefix,
		c.Config.GlobalPrefix,
		"variable/constant declarations are not sorted",
	)
}

func (c *Checker) checkStructType(pass *analysis.Pass, node *ast.StructType) bool {
	if !c.Config.StructFields.Enabled {
		return true
	}

	metadata := extractMetadata(pass, node.Fields.List, extractStructField, c.Config.IgnoreGroups)
	return checkElementsSorted(
		pass,
		metadata,
		c.Config.StructFields.Prefix,
		c.Config.GlobalPrefix,
		"struct fields are not sorted",
	)
}

func (c *Checker) checkInterfaceType(pass *analysis.Pass, node *ast.InterfaceType) bool {
	if !c.Config.InterfaceMethods.Enabled {
		return true
	}

	metadata := extractMetadata(pass, node.Methods.List, extractInterfaceMethod, c.Config.IgnoreGroups)
	return checkElementsSorted(
		pass,
		metadata,
		c.Config.InterfaceMethods.Prefix,
		c.Config.GlobalPrefix,
		"interface methods are not sorted",
	)
}

func (c *Checker) checkCallExpr(pass *analysis.Pass, node *ast.CallExpr) bool {
	if !c.Config.VariadicArgs.Enabled {
		return true
	}

	metadata := extractVariadicArgMetadata(pass, node, c.Config.IgnoreGroups)
	return checkElementsSorted(
		pass,
		metadata,
		c.Config.VariadicArgs.Prefix,
		c.Config.GlobalPrefix,
		"variadic arguments are not sorted",
	)
}

func (c *Checker) checkCompositeLit(pass *analysis.Pass, node *ast.CompositeLit) bool {
	if !c.Config.MapKeys.Enabled {
		return true
	}

	keyValueExprs := make([]*ast.KeyValueExpr, 0, len(node.Elts))
	for _, elt := range node.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			keyValueExprs = append(keyValueExprs, kv)
		}
	}

	metadata := extractMetadata(pass, keyValueExprs, extractMapKey, c.Config.IgnoreGroups)
	return checkElementsSorted(
		pass,
		metadata,
		c.Config.MapKeys.Prefix,
		c.Config.GlobalPrefix,
		"map keys are not sorted",
	)
}

func checkElementsSorted[T ast.Node](
	pass *analysis.Pass,
	groups [][]metadata[T],
	prefix, globalPrefix, msg string,
) bool {

	allSorted := true

	for _, group := range groups {
		if len(group) <= 1 {
			continue
		}

		for i := 1; i < len(group); i++ {
			if !hasPrefixOrGlobal(group[i].Value, prefix, globalPrefix) {
				continue
			}

			if group[i].Value < group[i-1].Value {
				allSorted = false

				pass.Report(analysis.Diagnostic{
					Pos:     group[i].Position,
					Message: msg,
				})
				break
			}
		}
	}

	return allSorted
}

func hasPrefixOrGlobal(name, prefix, globalPrefix string) bool {
	if prefix == "" {
		return hasPrefix(name, globalPrefix)
	}
	return hasPrefix(name, prefix)
}

// hasPrefix checks if a name starts with the specified prefix.
// Returns true if prefix is empty (no filtering) or if the name starts with the prefix.
func hasPrefix(name, prefix string) bool {
	if prefix == "" {
		return true // No filtering, all names match
	}

	if name == "" {
		return false // Empty name never matches a prefix
	}

	return strings.HasPrefix(name, prefix)
}
