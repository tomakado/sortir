package config

var defaults = map[string]any{
	FlagConstants:        true,
	FlagVariables:        true,
	FlagStructFields:     true,
	FlagInterfaceMethods: true,
	FlagMapKeys:          true,
}

func Default[T any](param string) T {
	if val, ok := defaults[param]; ok {
		return val.(T)
	}

	var zero T
	return zero
}
