package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.tomakado.io/sortir/internal/config"
)

func TestDefault(t *testing.T) {
	t.Run("string flags", func(t *testing.T) {
		const want = ""

		got := config.Default[string](config.FlagFilterPrefix)
		require.Equal(t, want, got)
	})

	t.Run("boolean flags", func(t *testing.T) {
		t.Run("existing true", func(t *testing.T) {
			const want = true

			got := config.Default[bool](config.FlagConstants)
			require.Equal(t, want, got)
		})

		t.Run("existing false", func(t *testing.T) {
			const want = false

			got := config.Default[bool](config.FlagVariadicArgs)
			require.Equal(t, want, got)
		})

		t.Run("non-existing", func(t *testing.T) {
			const want = false

			got := config.Default[bool]("dummy-flag")
			require.Equal(t, want, got)
		})
	})
}
