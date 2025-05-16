package sortir

import (
	"github.com/golangci/plugin-module-register/register"
	"go.tomakado.io/sortir/internal/analyzer"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("sortir", func(_ any) (register.LinterPlugin, error) {
		return &Plugin{}, nil
	})
}

type Plugin struct{}

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{analyzer.New().Analyzer}, nil
}

func (f *Plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
