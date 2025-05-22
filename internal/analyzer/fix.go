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
	case *ast.KeyValueExpr:
		replacement, from, to = a.generateKeyValueFix(pass, group, sorted)
	case ast.Expr:
		replacement, from, to = a.generateExprFix(pass, group, sorted)
	}

	if replacement == nil {
		return nil
	}

	return &FixSuggestion{
		From:        from,
		Message:     fmt.Sprintf("Sort %s", elementType),
		Replacement: replacement,
		To:          to,
	}
}

func (a *Analyzer) generateGenDeclFix(pass *analysis.Pass, original, sorted []Metadata) ([]byte, token.Pos, token.Pos) {
	if len(original) == 0 {
		return nil, 0, 0
	}

	// Try to preserve original formatting by extracting with line context
	if pass.ReadFile != nil {
		if result := a.generateGenDeclFixPreserveFormat(pass, original, sorted); result != nil {
			return result, original[0].Node.Pos(), original[len(original)-1].Node.End()
		}
	}

	// Fallback to node-by-node extraction
	return a.generateGenDeclFixNodeByNode(pass, original, sorted)
}

func (a *Analyzer) generateGenDeclFixPreserveFormat(pass *analysis.Pass, original, sorted []Metadata) []byte {
	if len(original) == 0 {
		return nil
	}

	content, file := a.getFileContent(pass, original[0].Node.Pos())
	if content == nil || file == nil {
		return nil
	}

	declLines := a.extractDeclLines(original, file, content)
	result := a.buildSortedResult(original, sorted, declLines)

	return []byte(strings.Join(result, "\n"))
}

func (a *Analyzer) getFileContent(pass *analysis.Pass, pos token.Pos) ([]byte, *token.File) {
	startPos := pass.Fset.Position(pos)
	content, err := pass.ReadFile(startPos.Filename)
	if err != nil {
		return nil, nil
	}

	file := pass.Fset.File(pos)
	return content, file
}

type declInfo struct {
	fullLine string
	meta     Metadata
}

func (a *Analyzer) extractDeclLines(original []Metadata, file *token.File, content []byte) map[*ast.ValueSpec]declInfo {
	declLines := make(map[*ast.ValueSpec]declInfo)

	for _, meta := range original {
		spec := meta.Node.(*ast.ValueSpec)
		fullLine := a.extractFullLine(file, content, meta.Line)

		if fullLine != "" {
			// Sort names within multi-name declarations
			if len(spec.Names) > 1 {
				fullLine = a.sortNamesInFullLine(fullLine, spec)
			}

			declLines[spec] = declInfo{
				fullLine: fullLine,
				meta:     meta,
			}
		}
	}

	return declLines
}

func (a *Analyzer) extractFullLine(file *token.File, content []byte, lineNum int) string {
	lineStart := file.LineStart(lineNum)
	lineEnd := a.findLineEnd(file, content, lineStart)

	startOffset := file.Offset(lineStart)
	endOffset := file.Offset(lineEnd)

	if startOffset >= 0 && endOffset <= len(content) {
		return string(content[startOffset:endOffset])
	}

	return ""
}

func (a *Analyzer) findLineEnd(file *token.File, content []byte, lineStart token.Pos) token.Pos {
	offset := file.Offset(lineStart)
	lineEnd := lineStart

	for offset < len(content) && content[offset] != '\n' {
		offset++
		lineEnd = file.Pos(offset)
	}

	return lineEnd
}

func (a *Analyzer) buildSortedResult(original, sorted []Metadata, declLines map[*ast.ValueSpec]declInfo) []string {
	var result []string

	for i, meta := range sorted {
		spec := meta.Node.(*ast.ValueSpec)
		if info, ok := declLines[spec]; ok {
			if i > 0 && a.shouldAddEmptyLine(original, sorted, i) {
				result = append(result, "")
			}
			result = append(result, info.fullLine)
		}
	}

	return result
}

func (a *Analyzer) shouldAddEmptyLine(original, sorted []Metadata, currentIdx int) bool {
	prevIdx := a.findOriginalIndex(original, sorted[currentIdx-1].Node)
	currIdx := a.findOriginalIndex(original, sorted[currentIdx].Node)

	if prevIdx >= 0 && currIdx >= 0 && prevIdx < len(original)-1 {
		return original[prevIdx+1].Line-original[prevIdx].Line > 1
	}

	return false
}

func (a *Analyzer) findOriginalIndex(original []Metadata, node ast.Node) int {
	for j, orig := range original {
		if orig.Node == node {
			return j
		}
	}
	return -1
}

func (a *Analyzer) generateGenDeclFixNodeByNode(pass *analysis.Pass, original, sorted []Metadata) ([]byte, token.Pos, token.Pos) {
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

		if len(spec.Names) > 1 {
			srcText = a.sortNamesInDecl(srcText, spec)
		}

		buf.WriteString(srcText)
	}

	return buf.Bytes(), from, to
}

func (a *Analyzer) generateFieldFix(pass *analysis.Pass, original, sorted []Metadata) ([]byte, token.Pos, token.Pos) {
	return a.generateIndentedFix(pass, original, sorted, "", "\n")
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

	for i, meta := range sorted {
		kv := meta.Node.(*ast.KeyValueExpr)

		if i > 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(sourceMap[kv])
	}

	return buf.Bytes(), from, to
}

func (a *Analyzer) generateIndentedFix(pass *analysis.Pass, original, sorted []Metadata, prefix, separator string) ([]byte, token.Pos, token.Pos) {
	if len(original) == 0 {
		return nil, 0, 0
	}

	sourceMap := a.buildSourceMap(pass, original)

	var buf bytes.Buffer
	from := original[0].Node.Pos()
	to := original[len(original)-1].Node.End()

	indentLevel := a.getCommonIndent(sourceMap)

	for i, meta := range sorted {
		if i > 0 {
			buf.WriteString(prefix + separator)
			if original[i].Line-original[i-1].Line > 1 {
				buf.WriteByte('\n')
			}
		}

		srcText := sourceMap[meta.Node]
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
	// Try to extract from source file first
	if src := a.extractFromFile(pass, node); src != "" {
		return src
	}
	
	// Fallback to formatting the node
	return a.formatNode(pass, node)
}

func (a *Analyzer) extractFromFile(pass *analysis.Pass, node ast.Node) string {
	if pass.ReadFile == nil {
		return ""
	}
	
	startPos := pass.Fset.Position(node.Pos())
	content, err := pass.ReadFile(startPos.Filename)
	if err != nil {
		return ""
	}
	
	file := pass.Fset.File(node.Pos())
	if file == nil {
		return ""
	}
	
	start := file.Offset(node.Pos())
	end := file.Offset(node.End())
	
	if start >= 0 && end <= len(content) && start < end {
		return string(content[start:end])
	}
	
	return ""
}

func (a *Analyzer) formatNode(pass *analysis.Pass, node ast.Node) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, pass.Fset, node); err != nil {
		return ""
	}
	return buf.String()
}

func (a *Analyzer) sortNamesInDecl(srcText string, spec *ast.ValueSpec) string {
	if len(spec.Names) <= 1 {
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

func (a *Analyzer) sortNamesInFullLine(line string, spec *ast.ValueSpec) string {
	if len(spec.Names) <= 1 {
		return line
	}

	names := make([]string, len(spec.Names))
	for i, name := range spec.Names {
		names[i] = name.Name
	}

	originalNames := strings.Join(names, ", ")
	sort.Strings(names)
	sortedNames := strings.Join(names, ", ")

	return strings.Replace(line, originalNames, sortedNames, 1)
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
