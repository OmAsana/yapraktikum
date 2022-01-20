package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/pkg"
)

func TestServerInitConfig(t *testing.T) {
	defaults := Config{
		Address:       "localhost:8080",
		StoreInterval: 300 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
		Restore:       true,
	}

	t.Run("check default", func(t *testing.T) {
		cfg, err := InitConfig()

		require.NoError(t, err)
		assert.Equal(t, cfg.Address, defaults.Address)
		assert.Equal(t, cfg.StoreInterval, defaults.StoreInterval)
		assert.Equal(t, cfg.StoreFile, defaults.StoreFile)
		assert.Equal(t, cfg.Restore, defaults.Restore)
	})

	t.Run("check overrides", func(t *testing.T) {
		overrides := map[string]string{
			"ADDRESS":        "127.0.0.1:1234",
			"STORE_INTERVAL": "1s",
			"STORE_FILE":     "/tmp/random_file_name.json",
			"RESTORE":        "false",
		}

		for k, v := range overrides {
			unset, err := pkg.SetEnv(t, k, v)
			require.NoError(t, err)
			defer unset()
		}

		cfg, err := InitConfig()
		require.NoError(t, err)
		assert.Equal(t, cfg.Address, overrides["ADDRESS"])
		assert.Equal(t, cfg.StoreInterval, func() time.Duration {
			duration, err := time.ParseDuration(overrides["STORE_INTERVAL"])
			require.NoError(t, err)
			return duration
		}())
		assert.Equal(t, cfg.StoreFile, overrides["STORE_FILE"])
		assert.Equal(t, cfg.Restore, func() bool {
			switch v := overrides["RESTORE"]; {
			case v == "true":
				return true
			case v == "false":
				return false
			default:
				assert.NoError(t, fmt.Errorf("value is not a valid boolean"))
				return false
			}
		}())

	})
}
