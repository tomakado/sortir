package analyzer

import (
	"go/token"

	"golang.org/x/tools/go/analysis"
)

type Diagnostic struct {
	From, To   token.Pos
	Message    string
	Suggestion *FixSuggestion
}

func (d Diagnostic) AsGoAnalysisDiagnostic() analysis.Diagnostic {
	var suggestedFixes []analysis.SuggestedFix
	if d.HasFixSuggestion() {
		suggestedFixes = []analysis.SuggestedFix{
			{
				Message: d.Suggestion.Message,
				TextEdits: []analysis.TextEdit{
					{
						Pos:     d.Suggestion.From,
						End:     d.Suggestion.To,
						NewText: []byte(d.Suggestion.Replacement),
					},
				},
			},
		}
	}

	return analysis.Diagnostic{
		Pos:            d.From,
		End:            d.To,
		Message:        d.Message,
		SuggestedFixes: suggestedFixes,
	}
}

func (d Diagnostic) HasFixSuggestion() bool {
	return d.Suggestion != nil && len(d.Suggestion.Replacement) > 0
}

type FixSuggestion struct {
	From, To    token.Pos
	Message     string
	Replacement []byte
}
