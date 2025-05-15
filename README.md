# Sortir

Sortir is a Go linter and formatter that checks and fixes sorting of various Go code elements.

## Features

Sortir enforces consistent ordering within the following code elements:

- Constant groups
- Variable groups
- Struct fields
- Interface methods
- Variadic arguments
- Map values

By default, Sortir only checks sorting within groups (elements not separated by empty lines), with an option to sort across empty lines.

## Installation

### Using go install

```bash
go install go.tomakado.io/sortir/cmd/sortir@latest
```

### From source

```bash
git clone https://github.com/tomakado/sortir.git
cd sortir
go install ./cmd/sortir
```

## Usage

### Command-line

```bash
# Check a package
sortir ./...

# Fix sorting issues automatically
sortir -fix ./...

# Use a custom configuration file
sortir -config=custom-config.yaml ./...
```

### As a golangci-lint plugin

Add to your `.golangci.yml`:

```yaml
linters:
  enable:
    - sortir

linters-settings:
  sortir:
    sort-across-empty-lines: false
    enabled-checks:
      constants: true
      variables: true
      struct-fields: true
      interface-methods: true
      variadic-args: true
      map-values: true
```

## Configuration

Sortir can be configured via command-line flags or a YAML configuration file (`.sortir.yaml`).

### Configuration File Example

```yaml
# .sortir.yaml
sort_across_empty_lines: false
fix_mode: false
verbose: false
enabled_checks:
  constants: true
  variables: true
  struct_fields: true
  interface_methods: true
  variadic_args: true
  map_values: true
```

### Command-line Options

- `-sort-across-empty-lines`: Sort elements across empty lines
- `-fix`: Automatically fix sorting issues
- `-verbose`: Enable verbose output
- `-config`: Path to configuration file
- `-check-constants`: Enable/disable checking constant declarations
- `-check-variables`: Enable/disable checking variable declarations
- `-check-struct-fields`: Enable/disable checking struct fields
- `-check-interface-methods`: Enable/disable checking interface methods
- `-check-variadic-args`: Enable/disable checking variadic arguments
- `-check-map-values`: Enable/disable checking map values

## Sorting Rules

- **Constants and variables**: Sorted alphabetically by name
- **Struct fields**: Named fields sorted alphabetically, anonymous fields sorted separately
- **Interface methods**: Sorted alphabetically by name
- **Variadic arguments**: Sorted by literal value or identifier name
- **Map values**: Sorted by map key (for sortable key types)

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.