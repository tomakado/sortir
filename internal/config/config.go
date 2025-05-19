package config

import "go.tomakado.io/sortir/internal/log"

type CheckConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

type SortConfig struct {
	GlobalPrefix   string `yaml:"prefix"`
	IgnoreGroups   bool   `yaml:"ignoreGroups"`
	FixModeEnabled bool   `yaml:"fix"`
	Verbose        bool   `yaml:"verbose"`

	Constants        *CheckConfig `yaml:"constants"`
	Variables        *CheckConfig `yaml:"variables"`
	StructFields     *CheckConfig `yaml:"structFields"`
	InterfaceMethods *CheckConfig `yaml:"interfaceMethods"`
	VariadicArgs     *CheckConfig `yaml:"variadicArgs"`
	MapKeys          *CheckConfig `yaml:"mapKeys"`
}

func New() *SortConfig {
	return &SortConfig{
		GlobalPrefix:   "",
		IgnoreGroups:   false,
		FixModeEnabled: false,
		Verbose:        false,

		Constants: &CheckConfig{
			Enabled: false,
			Prefix:  "",
		},
		Variables: &CheckConfig{
			Enabled: false,
			Prefix:  "",
		},
		StructFields: &CheckConfig{
			Enabled: false,
			Prefix:  "",
		},
		InterfaceMethods: &CheckConfig{
			Enabled: false,
			Prefix:  "",
		},
		VariadicArgs: &CheckConfig{
			Enabled: false,
			Prefix:  "",
		},
		MapKeys: &CheckConfig{
			Enabled: false,
			Prefix:  "",
		},
	}
}

func (c *SortConfig) LogLevel() log.Level {
	if c.Verbose {
		return log.Verbose
	}

	return log.Important
}
