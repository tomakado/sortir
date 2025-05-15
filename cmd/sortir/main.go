package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/tools/go/analysis/singlechecker"

	"go.tomakado.io/sortir/internal/analyzer"
	"go.tomakado.io/sortir/internal/config"
)

func main() {
	// Create a custom usage function to display help text
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Sortir: A Go linter/formatter for checking and fixing sorting of Go code elements\n\n")
		fmt.Fprintf(os.Stderr, "Usage: sortir [flags] [packages...]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	// Define command-line flags
	configPath := flag.String("config", ".sortir.yaml", "path to configuration file")
	
	// Create default configuration and register all flags
	// This must be done before flag.Parse()
	defaultCfg := config.DefaultConfig()
	defaultCfg.RegisterFlags(flag.CommandLine)
	
	// Parse flags
	flag.Parse()

	// Load configuration from file, flags are already registered and parsed
	cfg, err := config.Load(*configPath, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize analyzer with loaded config
	analyzer := analyzer.New(cfg)

	// Run the analyzer using the singlechecker package
	singlechecker.Main(analyzer)
}