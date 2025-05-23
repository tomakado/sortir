package config

var defaults = map[string]any{
	FlagConstants: true, FlagInterfaceMethods: true, FlagMapKeys: true, FlagStructFields: true, FlagVariables: true,
}

func Default[T any](param string) T {
	if val, ok := defaults[param]; ok {
		return val.(T)
	}

	var zero T
	return zero
}
