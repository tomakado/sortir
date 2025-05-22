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
						End: d.Suggestion.To, NewText: []byte(d.Suggestion.Replacement), Pos: d.Suggestion.From,
					},
				},
			},
		}
	}

	return analysis.Diagnostic{
		End: d.To, Message: d.Message, Pos: d.From, SuggestedFixes: suggestedFixes,
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
