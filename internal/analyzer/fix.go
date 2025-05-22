package analyzer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

func (a *Analyzer) generateFix(pass *analysis.Pass, group []Metadata, elementType string) *FixSuggestion {
	if len(group) <= 1 {
		return nil
	}

	sorted := make([]Metadata, len(group))
	copy(sorted, group)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value < sorted[j].Value
	})

	var replacement []byte
	var from, to token.Pos

	switch group[0].Node.(type) {
	case *ast.ValueSpec:
		replacement, from, to = a.generateGenDeclFix(pass, group, sorted)
	case *ast.Field:
		replacement, from, to = a.generateFieldFix(pass, group, sorted)
	case ast.Expr:
		replacement, from, to = a.generateExprFix(pass, group, sorted)
	case *ast.KeyValueExpr:
		replacement, from, to = a.generateKeyValueFix(pass, group, sorted)
	}

	if replacement == nil {
		return nil
	}

	return &FixSuggestion{
		From:        from,
		To:          to,
		Message:     fmt.Sprintf("Sort %s", elementType),
		Replacement: replacement,
	}
}

func (a *Analyzer) generateGenDeclFix(pass *analysis.Pass, original, sorted []Metadata) ([]byte, token.Pos, token.Pos) {
	if len(original) == 0 {
		return nil, 0, 0
	}

	sourceMap := a.buildSourceMap(pass, original)
	
	var buf bytes.Buffer
	from := original[0].Node.Pos()
	to := original[len(original)-1].Node.End()

	for i, meta := range sorted {
		spec := meta.Node.(*ast.ValueSpec)
		
		if i > 0 {
			buf.WriteByte('\n')
			if original[i].Line-original[i-1].Line > 1 {
				buf.WriteByte('\n')
			}
		}

		srcText := sourceMap[spec]
		
		if spec.Names != nil && len(spec.Names) > 1 {
			srcText = a.sortNamesInDecl(srcText, spec)
		}
		
		buf.WriteString(srcText)
	}

	return buf.Bytes(), from, to
}

func (a *Analyzer) generateFieldFix(pass *analysis.Pass, original, sorted []Metadata) ([]byte, token.Pos, token.Pos) {
	if len(original) == 0 {
		return nil, 0, 0
	}

	sourceMap := a.buildSourceMap(pass, original)
	
	var buf bytes.Buffer
	from := original[0].Node.Pos()
	to := original[len(original)-1].Node.End()

	indentLevel := a.getCommonIndent(sourceMap)

	for i, meta := range sorted {
		field := meta.Node.(*ast.Field)

		if i > 0 {
			buf.WriteByte('\n')
			if original[i].Line-original[i-1].Line > 1 {
				buf.WriteByte('\n')
			}
		}

		srcText := sourceMap[field]
		if !strings.HasPrefix(srcText, indentLevel) && indentLevel != "" {
			buf.WriteString(indentLevel)
		}
		buf.WriteString(srcText)
	}

	return buf.Bytes(), from, to
}

func (a *Analyzer) generateExprFix(pass *analysis.Pass, original, sorted []Metadata) ([]byte, token.Pos, token.Pos) {
	if len(original) == 0 {
		return nil, 0, 0
	}

	sourceMap := a.buildSourceMap(pass, original)
	
	var buf bytes.Buffer
	from := original[0].Node.Pos()
	to := original[len(original)-1].Node.End()

	for i, meta := range sorted {
		expr := meta.Node.(ast.Expr)

		if i > 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(sourceMap[expr])
	}

	return buf.Bytes(), from, to
}

func (a *Analyzer) generateKeyValueFix(pass *analysis.Pass, original, sorted []Metadata) ([]byte, token.Pos, token.Pos) {
	if len(original) == 0 {
		return nil, 0, 0
	}

	sourceMap := a.buildSourceMap(pass, original)
	
	var buf bytes.Buffer
	from := original[0].Node.Pos()
	to := original[len(original)-1].Node.End()

	indentLevel := a.getCommonIndent(sourceMap)

	for i, meta := range sorted {
		kv := meta.Node.(*ast.KeyValueExpr)

		if i > 0 {
			buf.WriteString(",\n")
			if original[i].Line-original[i-1].Line > 1 {
				buf.WriteByte('\n')
			}
		}

		srcText := sourceMap[kv]
		if !strings.HasPrefix(srcText, indentLevel) && indentLevel != "" {
			buf.WriteString(indentLevel)
		}
		buf.WriteString(srcText)
	}

	return buf.Bytes(), from, to
}

func (a *Analyzer) buildSourceMap(pass *analysis.Pass, metadata []Metadata) map[ast.Node]string {
	result := make(map[ast.Node]string)
	
	for _, meta := range metadata {
		srcText := a.extractNodeSource(pass, meta.Node)
		result[meta.Node] = srcText
	}
	
	return result
}

func (a *Analyzer) extractNodeSource(pass *analysis.Pass, node ast.Node) string {
	startPos := pass.Fset.Position(node.Pos())
	
	if pass.ReadFile == nil {
		var buf bytes.Buffer
		format.Node(&buf, pass.Fset, node)
		return buf.String()
	}
	
	content, err := pass.ReadFile(startPos.Filename)
	if err != nil {
		var buf bytes.Buffer
		format.Node(&buf, pass.Fset, node)
		return buf.String()
	}
	
	file := pass.Fset.File(node.Pos())
	if file == nil {
		var buf bytes.Buffer
		format.Node(&buf, pass.Fset, node)
		return buf.String()
	}
	
	start := file.Offset(node.Pos())
	end := file.Offset(node.End())
	
	if start >= 0 && end <= len(content) && start < end {
		return string(content[start:end])
	}
	
	var buf bytes.Buffer
	format.Node(&buf, pass.Fset, node)
	return buf.String()
}

func (a *Analyzer) sortNamesInDecl(srcText string, spec *ast.ValueSpec) string {
	if spec.Names == nil || len(spec.Names) <= 1 {
		return srcText
	}
	
	names := make([]string, len(spec.Names))
	for i, name := range spec.Names {
		names[i] = name.Name
	}
	
	originalNames := strings.Join(names, ", ")
	sort.Strings(names)
	sortedNames := strings.Join(names, ", ")
	
	return strings.Replace(srcText, originalNames, sortedNames, 1)
}

func (a *Analyzer) getCommonIndent(sourceMap map[ast.Node]string) string {
	var minIndent *string
	
	for _, src := range sourceMap {
		lines := strings.Split(src, "\n")
		if len(lines) == 0 {
			continue
		}
		
		firstLine := lines[0]
		indent := ""
		for _, ch := range firstLine {
			if ch == ' ' || ch == '\t' {
				indent += string(ch)
			} else {
				break
			}
		}
		
		if minIndent == nil || len(indent) < len(*minIndent) {
			minIndent = &indent
		}
	}
	
	if minIndent == nil {
		return ""
	}
	return *minIndent
}