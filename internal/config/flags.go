package config

const (
	// Check enabling flags
	CheckConstants        = "check-constants"
	CheckVariables        = "check-variables"
	CheckStructFields     = "check-struct-fields"
	CheckInterfaceMethods = "check-interface-methods"
	CheckVariadicArgs     = "check-variadic-args"
	CheckMapValues        = "check-map-values"

	// General configuration flags
	IgnoreGroups = "ignore-groups"
	FilterPrefix = "filter-prefix"

	// Per-check filter prefix flags
	ConstantsPrefix        = "constants-prefix"
	VariablesPrefix        = "variables-prefix"
	StructFieldsPrefix     = "struct-fields-prefix"
	InterfaceMethodsPrefix = "interface-methods-prefix"
	VariadicArgsPrefix     = "variadic-args-prefix"
	MapValuesPrefix        = "map-values-prefix"
)
