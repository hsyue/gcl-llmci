package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("llmci", New)
}

func New(settings any) (register.LinterPlugin, error) {
	// The configuration type will be map[string]any or []interface, it depends on your configuration.
	// You can use https://github.com/go-viper/mapstructure to convert map to struct.

	s, err := register.DecodeSettings[LLMCiPluginConfig](settings)
	if err != nil {
		return nil, err
	}

	return &LLMCiPlugin{config: &s}, nil
}

type LLMCiPlugin struct {
	config *LLMCiPluginConfig
}

func (f *LLMCiPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	analyzer := &LLMCiAnalyzer{config: f.config}
	return []*analysis.Analyzer{
		{
			Name: "llmci-analyzer",
			Doc:  "LLMCi Analyzer",
			Run:  analyzer.run,
		},
	}, nil
}

func (f *LLMCiPlugin) GetLoadMode() string {
	// NOTE: the mode can be `register.LoadModeSyntax` or `register.LoadModeTypesInfo`.
	// - `register.LoadModeSyntax`: if the linter doesn't use types information.
	// - `register.LoadModeTypesInfo`: if the linter uses types information.

	return register.LoadModeSyntax
}
