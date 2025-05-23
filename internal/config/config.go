package config

import "go.tomakado.io/sortir/internal/log"

type CheckConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

type SortConfig struct {
	FixModeEnabled bool   `yaml:"fix"`
	GlobalPrefix   string `yaml:"prefix"`
	IgnoreGroups   bool   `yaml:"ignoreGroups"`
	Verbose        bool   `yaml:"verbose"`

	Constants        *CheckConfig `yaml:"constants"`
	InterfaceMethods *CheckConfig `yaml:"interfaceMethods"`
	MapKeys          *CheckConfig `yaml:"mapKeys"`
	StructFields     *CheckConfig `yaml:"structFields"`
	Variables        *CheckConfig `yaml:"variables"`
	VariadicArgs     *CheckConfig `yaml:"variadicArgs"`
}

func New() *SortConfig {
	return &SortConfig{
		FixModeEnabled: Default[bool](FlagFix), GlobalPrefix: Default[string](FlagFilterPrefix), IgnoreGroups: Default[bool](FlagIgnoreGroups), Verbose: Default[bool](FlagVerbose),

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
