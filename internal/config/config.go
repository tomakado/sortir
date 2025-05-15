package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// SortConfig holds configuration for sorting behavior and enabled checks
type SortConfig struct {
	// When false (default), sorting only happens within groups (elements not separated by empty lines)
	SortAcrossEmptyLines bool `yaml:"sort_across_empty_lines"`

	EnabledChecks EnabledChecksConfig `yaml:"enabled_checks"`
	FixMode       bool                `yaml:"fix_mode"`
	Verbose       bool                `yaml:"verbose"`
}

// EnabledChecksConfig specifies which element types to check for sorting
type EnabledChecksConfig struct {
	Constants        bool `yaml:"constants"`
	Variables        bool `yaml:"variables"`
	StructFields     bool `yaml:"struct_fields"`
	InterfaceMethods bool `yaml:"interface_methods"`
	VariadicArgs     bool `yaml:"variadic_args"`
	MapValues        bool `yaml:"map_values"`
}

func DefaultConfig() *SortConfig {
	return &SortConfig{
		SortAcrossEmptyLines: false,
		EnabledChecks: EnabledChecksConfig{
			Constants:        true,
			Variables:        true,
			StructFields:     true,
			InterfaceMethods: true,
			VariadicArgs:     true,
			MapValues:        true,
		},
		FixMode: false,
		Verbose: false,
	}
}

func (c *SortConfig) LoadFromFile(path string) error {
	if path == "" {
		return nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (c *SortConfig) RegisterFlags(fs *flag.FlagSet) {
	if fs == nil {
		fs = flag.CommandLine
	}

	fs.BoolVar(&c.SortAcrossEmptyLines, "sort-across-empty-lines", c.SortAcrossEmptyLines, "Sort elements across empty lines")
	fs.BoolVar(&c.FixMode, "fix", c.FixMode, "Automatically fix sorting issues")
	fs.BoolVar(&c.Verbose, "verbose", c.Verbose, "Enable verbose output")

	fs.BoolVar(&c.EnabledChecks.Constants, "check-constants", c.EnabledChecks.Constants, "Check constant declarations")
	fs.BoolVar(&c.EnabledChecks.Variables, "check-variables", c.EnabledChecks.Variables, "Check variable declarations")
	fs.BoolVar(&c.EnabledChecks.StructFields, "check-struct-fields", c.EnabledChecks.StructFields, "Check struct fields")
	fs.BoolVar(&c.EnabledChecks.InterfaceMethods, "check-interface-methods", c.EnabledChecks.InterfaceMethods, "Check interface methods")
	fs.BoolVar(&c.EnabledChecks.VariadicArgs, "check-variadic-args", c.EnabledChecks.VariadicArgs, "Check variadic arguments")
	fs.BoolVar(&c.EnabledChecks.MapValues, "check-map-values", c.EnabledChecks.MapValues, "Check map values")
}

func Load(configPath string, fs *flag.FlagSet) (*SortConfig, error) {
	config := DefaultConfig()

	if configPath != "" {
		if err := config.LoadFromFile(configPath); err != nil {
			return nil, err
		}
	} else {
		defaultPaths := []string{
			".sortir.yaml",
			".sortir.yml",
		}

		for _, path := range defaultPaths {
			if err := config.LoadFromFile(path); err == nil {
				break
			}
		}
	}

	if fs != nil {
		config.RegisterFlags(fs)
	}

	return config, nil
}
