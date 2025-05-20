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
		GlobalPrefix:   Default[string](FlagFilterPrefix),
		IgnoreGroups:   Default[bool](FlagIgnoreGroups),
		FixModeEnabled: Default[bool](FlagFix),
		Verbose:        Default[bool](FlagVerbose),

		Constants: &CheckConfig{
			Enabled: Default[bool](FlagConstants),
			Prefix:  Default[string](FlagConstantsPrefix),
		},
		Variables: &CheckConfig{
			Enabled: Default[bool](FlagVariables),
			Prefix:  Default[string](FlagVariablesPrefix),
		},
		StructFields: &CheckConfig{
			Enabled: Default[bool](FlagStructFields),
			Prefix:  Default[string](FlagStructFieldsPrefix),
		},
		InterfaceMethods: &CheckConfig{
			Enabled: Default[bool](FlagInterfaceMethods),
			Prefix:  Default[string](FlagInterfaceMethodsPrefix),
		},
		VariadicArgs: &CheckConfig{
			Enabled: Default[bool](FlagVariadicArgs),
			Prefix:  Default[string](FlagVariadicArgsPrefix),
		},
		MapKeys: &CheckConfig{
			Enabled: Default[bool](FlagMapKeys),
			Prefix:  Default[string](FlagMapKeysPrefix),
		},
	}
}

func (c *SortConfig) LogLevel() log.Level {
	if c.Verbose {
		return log.Verbose
	}

	return log.Important
}
