// Package config provides configuration management for the sortir linter.
// It handles loading configuration from files and command-line flags.
package config

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// CheckConfig holds configuration for a specific check type
type CheckConfig struct {
	Enabled bool   `yaml:"enabled"` // Whether this check type is enabled
	Prefix  string `yaml:"prefix"`  // Filter prefix specific to this check type
}

// SortConfig holds configuration for sorting behavior and enabled checks.
type SortConfig struct {
	IgnoreGroups     bool         `yaml:"ignoreGroups"`
	FilterPrefix     string       `yaml:"filterPrefix"`     // Global filter prefix
	Constants        *CheckConfig `yaml:"constants"`        // Constants check configuration
	Variables        *CheckConfig `yaml:"variables"`        // Variables check configuration
	StructFields     *CheckConfig `yaml:"structFields"`     // Struct fields check configuration
	InterfaceMethods *CheckConfig `yaml:"interfaceMethods"` // Interface methods check configuration
	VariadicArgs     *CheckConfig `yaml:"variadicArgs"`     // Variadic arguments check configuration
	MapValues        *CheckConfig `yaml:"mapValues"`        // Map values check configuration
	FixMode          bool         `yaml:"fixMode"`
	Verbose          bool         `yaml:"verbose"`
}

func New() *SortConfig {
	return &SortConfig{
		IgnoreGroups: false,
		FilterPrefix: "",
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
		MapValues: &CheckConfig{
			Enabled: false,
			Prefix:  "",
		},
		FixMode: false,
	}
}

// LoadFromFile loads configuration from a YAML file at the given path.
// If the file doesn't exist, it returns nil without an error.
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

// RegisterFlags registers all configuration options as command-line flags.
// If flagSet is nil, it uses flag.CommandLine.
func (c *SortConfig) RegisterFlags(flagSet *flag.FlagSet) {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}

	// General config flags
	flagSet.BoolVar(
		&c.IgnoreGroups,
		"ignore-groups",
		c.IgnoreGroups,
		"Sort elements across empty lines",
	)
	flagSet.StringVar(
		&c.FilterPrefix,
		"filter-prefix",
		"",
		"Only check sorting for symbols starting with specified prefix (global)",
	)
	// flagSet.BoolVar(&c.FixMode, "fix", c.FixMode, "Automatically fix sorting issues")
	flagSet.BoolVar(&c.Verbose, "verbose", c.Verbose, "Enable verbose output")

	// Specific check flags
	flagSet.BoolVar(
		&c.Constants.Enabled,
		"check-constants",
		c.Constants.Enabled,
		"Check constant declarations",
	)
	flagSet.StringVar(
		&c.Constants.Prefix,
		"constants-prefix",
		c.Constants.Prefix,
		"Only check sorting for constants starting with specified prefix",
	)

	flagSet.BoolVar(
		&c.Variables.Enabled,
		"check-variables",
		c.Variables.Enabled,
		"Check variable declarations",
	)
	flagSet.StringVar(
		&c.Variables.Prefix,
		"variables-prefix",
		c.Variables.Prefix,
		"Only check sorting for variables starting with specified prefix",
	)

	flagSet.BoolVar(
		&c.StructFields.Enabled,
		"check-struct-fields",
		c.StructFields.Enabled,
		"Check struct fields",
	)
	flagSet.StringVar(
		&c.StructFields.Prefix,
		"struct-fields-prefix",
		c.StructFields.Prefix,
		"Only check sorting for struct fields starting with specified prefix",
	)

	flagSet.BoolVar(
		&c.InterfaceMethods.Enabled,
		"check-interface-methods",
		c.InterfaceMethods.Enabled,
		"Check interface methods",
	)
	flagSet.StringVar(
		&c.InterfaceMethods.Prefix,
		"interface-methods-prefix",
		c.InterfaceMethods.Prefix,
		"Only check sorting for interface methods starting with specified prefix",
	)

	flagSet.BoolVar(
		&c.VariadicArgs.Enabled,
		"check-variadic-args",
		c.VariadicArgs.Enabled,
		"Check variadic arguments",
	)
	flagSet.StringVar(
		&c.VariadicArgs.Prefix,
		"variadic-args-prefix",
		c.VariadicArgs.Prefix,
		"Only check sorting for variadic arguments starting with specified prefix",
	)

	flagSet.BoolVar(
		&c.MapValues.Enabled,
		"check-map-values",
		c.MapValues.Enabled,
		"Check map values",
	)
	flagSet.StringVar(
		&c.MapValues.Prefix,
		"map-values-prefix",
		c.MapValues.Prefix,
		"Only check sorting for map values starting with specified prefix",
	)
}

// Load creates a new configuration by loading from file and flags.
// It loads from configPath if provided, otherwise checks default paths.
// If flagSet is provided, it registers flags on that FlagSet.
func Load(configPath string, flagSet *flag.FlagSet) (*SortConfig, error) {
	config := &SortConfig{}

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

	if flagSet != nil {
		config.RegisterFlags(flagSet)
		flag.Parse()
	}

	return config, nil
}
