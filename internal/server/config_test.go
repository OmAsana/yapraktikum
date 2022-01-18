package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/pkg"
)

func TestServerInitConfig(t *testing.T) {
	t.Run("check default", func(t *testing.T) {
		cfg, err := InitConfig()

		require.NoError(t, err)
		assert.Equal(t, cfg.Address, "localhost:8080")
	})

	t.Run("check overrides", func(t *testing.T) {
		newAddress := "127.0.0.1:1234"
		pkg.SetEnv(t, "ADDRESS", newAddress)

		cfg, err := InitConfig()
		require.NoError(t, err)
		assert.Equal(t, cfg.Address, newAddress)

	})
}
