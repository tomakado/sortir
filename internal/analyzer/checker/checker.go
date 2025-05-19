// Package checker implements the checking logic for sorting of Go code elements.
package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"go.tomakado.io/sortir/internal/config"
	"go.tomakado.io/sortir/internal/log"
)

type Checker struct {
	logger *log.Logger

	Config *config.SortConfig
}

func New(cfg *config.SortConfig) *Checker {
	return &Checker{Config: cfg}
}

func (c *Checker) WithLogger(logger *log.Logger) *Checker {
	return &Checker{
		logger: logger,
		Config: c.Config,
	}
}

func (c *Checker) CheckNode(pass *analysis.Pass, node ast.Node) bool {
	switch n := node.(type) {
	case *ast.GenDecl:
		c.logger.Verbose("Checking GenDecl node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return c.checkGenDecl(pass, n)
	case *ast.StructType:
		c.logger.Verbose("Checking StructType node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return c.checkStructType(pass, n)
	case *ast.InterfaceType:
		c.logger.Verbose("Checking InterfaceType node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return c.checkInterfaceType(pass, n)
	case *ast.CallExpr:
		c.logger.Verbose("Checking CallExpr node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return c.checkCallExpr(pass, n)
	case *ast.CompositeLit:
		c.logger.Verbose("Checking CompositeLit node", log.FieldPosition, pass.Fset.Position(n.Pos()))
		return c.checkCompositeLit(pass, n)
	}
	return true
}

func (c *Checker) checkGenDecl(pass *analysis.Pass, node *ast.GenDecl) bool {
	var prefix string
	switch node.Tok {
	case token.CONST:
		prefix = c.Config.Constants.Prefix
		c.logger.Verbose("Processing constants", log.FieldEnabled, c.Config.Constants.Enabled, log.FieldPrefix, prefix)
	case token.VAR:
		prefix = c.Config.Variables.Prefix
		c.logger.Verbose("Processing variables", log.FieldEnabled, c.Config.Variables.Enabled, log.FieldPrefix, prefix)
	}

	isConst := node.Tok == token.CONST
	isVar := node.Tok == token.VAR

	shouldCheck := isConst && c.Config.Constants.Enabled || isVar && c.Config.Variables.Enabled
	if !shouldCheck {
		c.logger.Verbose("Skipping checks", log.FieldNodeType, node.Tok.String())
		return true
	}

	valueSpecs := make([]*ast.ValueSpec, 0, len(node.Specs))
	for _, spec := range node.Specs {
		valueSpecs = append(valueSpecs, spec.(*ast.ValueSpec))
	}

	c.logger.Verbose("Extracting metadata", log.FieldSpecsCount, len(valueSpecs), log.FieldIgnoreGroups, c.Config.IgnoreGroups)
	metadata := extractMetadata(pass, valueSpecs, extractGenDecl, c.Config.IgnoreGroups)
	c.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, prefix, log.FieldGlobalPrefix, c.Config.GlobalPrefix)
	return checkElementsSorted(
		pass,
		metadata,
		prefix,
		c.Config.GlobalPrefix,
		"variable/constant declarations are not sorted",
		c.logger,
	)
}

func (c *Checker) checkStructType(pass *analysis.Pass, node *ast.StructType) bool {
	c.logger.Verbose("Processing struct fields", log.FieldEnabled, c.Config.StructFields.Enabled, log.FieldPrefix, c.Config.StructFields.Prefix)
	if !c.Config.StructFields.Enabled {
		c.logger.Verbose("Skipping struct field checks")
		return true
	}

	c.logger.Verbose("Extracting metadata", log.FieldFieldsCount, len(node.Fields.List), log.FieldIgnoreGroups, c.Config.IgnoreGroups)
	metadata := extractMetadata(pass, node.Fields.List, extractStructField, c.Config.IgnoreGroups)
	c.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, c.Config.StructFields.Prefix, log.FieldGlobalPrefix, c.Config.GlobalPrefix)
	return checkElementsSorted(
		pass,
		metadata,
		c.Config.StructFields.Prefix,
		c.Config.GlobalPrefix,
		"struct fields are not sorted",
		c.logger,
	)
}

func (c *Checker) checkInterfaceType(pass *analysis.Pass, node *ast.InterfaceType) bool {
	c.logger.Verbose("Processing interface methods", log.FieldEnabled, c.Config.InterfaceMethods.Enabled, log.FieldPrefix, c.Config.InterfaceMethods.Prefix)
	if !c.Config.InterfaceMethods.Enabled {
		c.logger.Verbose("Skipping interface method checks")
		return true
	}

	c.logger.Verbose("Extracting metadata", log.FieldMethodsCount, len(node.Methods.List), log.FieldIgnoreGroups, c.Config.IgnoreGroups)
	metadata := extractMetadata(pass, node.Methods.List, extractInterfaceMethod, c.Config.IgnoreGroups)
	c.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, c.Config.InterfaceMethods.Prefix, log.FieldGlobalPrefix, c.Config.GlobalPrefix)
	return checkElementsSorted(
		pass,
		metadata,
		c.Config.InterfaceMethods.Prefix,
		c.Config.GlobalPrefix,
		"interface methods are not sorted",
		c.logger,
	)
}

func (c *Checker) checkCallExpr(pass *analysis.Pass, node *ast.CallExpr) bool {
	c.logger.Verbose("Processing variadic arguments", log.FieldEnabled, c.Config.VariadicArgs.Enabled, log.FieldPrefix, c.Config.VariadicArgs.Prefix)
	if !c.Config.VariadicArgs.Enabled {
		c.logger.Verbose("Skipping variadic argument checks")
		return true
	}

	c.logger.Verbose("Extracting metadata", log.FieldArgsCount, len(node.Args), log.FieldIgnoreGroups, c.Config.IgnoreGroups)
	metadata := extractVariadicArgMetadata(pass, node, c.Config.IgnoreGroups)
	c.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, c.Config.VariadicArgs.Prefix, log.FieldGlobalPrefix, c.Config.GlobalPrefix)
	return checkElementsSorted(
		pass,
		metadata,
		c.Config.VariadicArgs.Prefix,
		c.Config.GlobalPrefix,
		"variadic arguments are not sorted",
		c.logger,
	)
}

func (c *Checker) checkCompositeLit(pass *analysis.Pass, node *ast.CompositeLit) bool {
	c.logger.Verbose("Processing map keys", "composite_lit_type", fmt.Sprintf("%#v", node.Type), log.FieldEnabled, c.Config.MapKeys.Enabled, log.FieldPrefix, c.Config.MapKeys.Prefix)
	if !c.Config.MapKeys.Enabled {
		c.logger.Verbose("Skipping map key checks")
		return true
	}

	keyValueExprs := make([]*ast.KeyValueExpr, 0, len(node.Elts))
	for _, elt := range node.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			keyValueExprs = append(keyValueExprs, kv)
		}
	}

	c.logger.Verbose("Extracting metadata", log.FieldKeyValueCount, len(keyValueExprs), log.FieldIgnoreGroups, c.Config.IgnoreGroups)
	metadata := extractMetadata(pass, keyValueExprs, extractMapKey, c.Config.IgnoreGroups)
	c.logger.Verbose("Checking elements sorted", log.FieldGroupsCount, len(metadata), log.FieldPrefix, c.Config.MapKeys.Prefix, log.FieldGlobalPrefix, c.Config.GlobalPrefix)
	return checkElementsSorted(
		pass,
		metadata,
		c.Config.MapKeys.Prefix,
		c.Config.GlobalPrefix,
		"composite literal elements are not sorted",
		c.logger,
	)
}

func checkElementsSorted[T ast.Node](
	pass *analysis.Pass,
	groups [][]metadata[T],
	prefix, globalPrefix, msg string,
	logger *log.Logger,
) bool {

	allSorted := true

	for groupIdx, group := range groups {
		if len(group) <= 1 {
			logger.Verbose("Skipping group with less than 2 elements", log.FieldGroupIndex, groupIdx, log.FieldGroupSize, len(group))
			continue
		}

		logger.Verbose("Checking group sorting", log.FieldGroupIndex, groupIdx, log.FieldGroupSize, len(group))
		for i := 1; i < len(group); i++ {
			if !hasPrefixOrGlobal(group[i].Value, prefix, globalPrefix) {
				logger.Verbose("Skipping element - no matching prefix", log.FieldElement, group[i].Value, log.FieldPrefix, prefix, log.FieldGlobalPrefix, globalPrefix)
				continue
			}

			if group[i].Value < group[i-1].Value {
				allSorted = false
				if pass.Fset != nil {
					logger.Verbose("Found unsorted elements", log.FieldCurrent, group[i].Value, log.FieldPrevious, group[i-1].Value, log.FieldPosition, pass.Fset.Position(group[i].Position))
				} else {
					logger.Verbose("Found unsorted elements", log.FieldCurrent, group[i].Value, log.FieldPrevious, group[i-1].Value, log.FieldPosition, group[i].Position)
				}

				pass.Report(analysis.Diagnostic{
					Pos:     group[i].Position,
					Message: msg,
				})
				break
			}
		}
	}

	logger.Verbose("Sorting check complete", log.FieldAllSorted, allSorted)
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
