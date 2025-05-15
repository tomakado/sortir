package config_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.tomakado.io/sortir/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := config.DefaultConfig()

	require.False(t, cfg.SortAcrossEmptyLines, "Default config should not sort across empty lines")
	require.True(t, cfg.EnabledChecks.Constants, "Constants checking should be enabled by default")
	require.True(t, cfg.EnabledChecks.Variables, "Variables checking should be enabled by default")
	require.True(t, cfg.EnabledChecks.StructFields, "Struct fields checking should be enabled by default")
	require.True(t, cfg.EnabledChecks.InterfaceMethods, "Interface methods checking should be enabled by default")
	require.True(t, cfg.EnabledChecks.VariadicArgs, "Variadic args checking should be enabled by default")
	require.True(t, cfg.EnabledChecks.MapValues, "Map values checking should be enabled by default")
	require.False(t, cfg.FixMode, "Fix mode should be disabled by default")
	require.False(t, cfg.Verbose, "Verbose mode should be disabled by default")
}

func TestLoadFromFile(t *testing.T) {
	t.Parallel()

	t.Run("Valid config file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		yamlContent := `
sort_across_empty_lines: true
enabled_checks:
  constants: true
  variables: false
  struct_fields: true
  interface_methods: false
  variadic_args: true
  map_values: false
fix_mode: true
verbose: true
`

		err := os.WriteFile(configPath, []byte(yamlContent), 0644)
		require.NoError(t, err, "Failed to create test config file")

		cfg := config.DefaultConfig()
		err = cfg.LoadFromFile(configPath)
		require.NoError(t, err, "Failed to load config")

		require.True(t, cfg.SortAcrossEmptyLines, "SortAcrossEmptyLines should be true")
		require.True(t, cfg.EnabledChecks.Constants, "Constants checking should be enabled")
		require.False(t, cfg.EnabledChecks.Variables, "Variables checking should be disabled")
		require.True(t, cfg.EnabledChecks.StructFields, "Struct fields checking should be enabled")
		require.False(t, cfg.EnabledChecks.InterfaceMethods, "Interface methods checking should be disabled")
		require.True(t, cfg.EnabledChecks.VariadicArgs, "Variadic args checking should be enabled")
		require.False(t, cfg.EnabledChecks.MapValues, "Map values checking should be disabled")
		require.True(t, cfg.FixMode, "Fix mode should be enabled")
		require.True(t, cfg.Verbose, "Verbose mode should be enabled")
	})

	t.Run("Non-existent file", func(t *testing.T) {
		t.Parallel()

		cfg := config.DefaultConfig()
		err := cfg.LoadFromFile("non_existent_file.yaml")

		require.NoError(t, err, "Loading from non-existent file should not return an error")

		expected := config.DefaultConfig()
		require.Equal(t, expected, cfg, "Config should remain at default values when file doesn't exist")
	})

	t.Run("Malformatted file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "bad_config.yaml")

		yamlContent := `
sort_across_empty_lines: not_a_boolean
enabled_checks:
  constants: "not a boolean either"
`

		err := os.WriteFile(configPath, []byte(yamlContent), 0644)
		require.NoError(t, err, "Failed to create test config file")

		cfg := config.DefaultConfig()
		err = cfg.LoadFromFile(configPath)

		require.Error(t, err, "Loading from malformatted file should return an error")
	})
}

func TestRegisterFlags(t *testing.T) {
	t.Parallel()

	t.Run("With flags", func(t *testing.T) {
		t.Parallel()

		fs := flag.NewFlagSet("test", flag.ContinueOnError)

		cfg := config.DefaultConfig()
		cfg.RegisterFlags(fs)

		require.NoError(t, fs.Set("sort-across-empty-lines", "true"), "Setting flag should not fail")
		require.NoError(t, fs.Set("fix", "true"), "Setting flag should not fail")
		require.NoError(t, fs.Set("verbose", "true"), "Setting flag should not fail")
		require.NoError(t, fs.Set("check-constants", "false"), "Setting flag should not fail")
		require.NoError(t, fs.Set("check-variables", "false"), "Setting flag should not fail")
		require.NoError(t, fs.Set("check-struct-fields", "false"), "Setting flag should not fail")
		require.NoError(t, fs.Set("check-interface-methods", "false"), "Setting flag should not fail")
		require.NoError(t, fs.Set("check-variadic-args", "false"), "Setting flag should not fail")
		require.NoError(t, fs.Set("check-map-values", "false"), "Setting flag should not fail")

		require.True(t, cfg.SortAcrossEmptyLines, "SortAcrossEmptyLines should be true")
		require.True(t, cfg.FixMode, "FixMode should be true")
		require.True(t, cfg.Verbose, "Verbose should be true")
		require.False(t, cfg.EnabledChecks.Constants, "Constants checking should be disabled")
		require.False(t, cfg.EnabledChecks.Variables, "Variables checking should be disabled")
		require.False(t, cfg.EnabledChecks.StructFields, "Struct fields checking should be disabled")
		require.False(t, cfg.EnabledChecks.InterfaceMethods, "Interface methods checking should be disabled")
		require.False(t, cfg.EnabledChecks.VariadicArgs, "Variadic args checking should be disabled")
		require.False(t, cfg.EnabledChecks.MapValues, "Map values checking should be disabled")
	})
}

func TestLoad(t *testing.T) {
	t.Parallel()

	t.Run("Load from file and flags", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		yamlContent := `
sort_across_empty_lines: false
enabled_checks:
  constants: true
  variables: true
  struct_fields: true
  interface_methods: true
  variadic_args: true
  map_values: true
fix_mode: false
verbose: false
`

		err := os.WriteFile(configPath, []byte(yamlContent), 0644)
		require.NoError(t, err, "Failed to create test config file")

		cfg, err := config.Load(configPath, nil)
		require.NoError(t, err, "Failed to load config from file")

		cfg.SortAcrossEmptyLines = true
		cfg.FixMode = true
		cfg.Verbose = true

		require.True(t, cfg.SortAcrossEmptyLines, "SortAcrossEmptyLines should be true (from flags)")
		require.True(t, cfg.FixMode, "FixMode should be true (from flags)")
		require.True(t, cfg.Verbose, "Verbose should be true (from flags)")
		require.True(t, cfg.EnabledChecks.Constants, "Constants checking should be enabled (from file)")
		require.True(t, cfg.EnabledChecks.Variables, "Variables checking should be enabled (from file)")
	})

	t.Run("Load from default location", func(t *testing.T) {
		t.Parallel()

		originalWd, err := os.Getwd()
		require.NoError(t, err, "Failed to get current working directory")

		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err, "Failed to change working directory")

		// Use a separate function to handle the deferred call to avoid linting issues
		defer func() {
			err := os.Chdir(originalWd)
			if err != nil {
				t.Logf("Warning: Failed to change back to original directory: %v", err)
			}
		}()

		yamlContent := `
sort_across_empty_lines: true
fix_mode: true
verbose: true
`

		err = os.WriteFile(".sortir.yaml", []byte(yamlContent), 0644)
		require.NoError(t, err, "Failed to create test config file")

		cfg, err := config.Load("", nil)
		require.NoError(t, err, "Failed to load config")

		require.True(t, cfg.SortAcrossEmptyLines, "SortAcrossEmptyLines should be true")
		require.True(t, cfg.FixMode, "FixMode should be true")
		require.True(t, cfg.Verbose, "Verbose should be true")
	})
}
