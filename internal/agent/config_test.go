package agent

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/pkg"
)

func TestAgentInitConfig(t *testing.T) {
	t.Run("check overrides", func(t *testing.T) {
		newAddress := "127.0.0.1:1234"
		hashKey := "someKey"
		newPollInterval := 2 * time.Second
		newReportInterval := 5 * time.Second

		unsetAdd, _ := pkg.SetEnv(t, "ADDRESS", newAddress)
		unsetPoll, _ := pkg.SetEnv(t, "POLL_INTERVAL", fmt.Sprintf("%ds", int(newPollInterval.Seconds())))
		unsetReport, _ := pkg.SetEnv(t, "REPORT_INTERVAL", fmt.Sprintf("%ds", int(newReportInterval.Seconds())))
		unsetHashKey, _ := pkg.SetEnv(t, "KEY", hashKey)
		defer func() {
			unsetReport()
			unsetPoll()
			unsetAdd()
			unsetHashKey()
		}()

		cfg := DefaultConfig
		err := cfg.initEnvArgs()
		require.NoError(t, err)
		assert.Equal(t, cfg.Address, newAddress)
		assert.Equal(t, cfg.PollInterval, newPollInterval)
		assert.Equal(t, cfg.ReportInterval, newReportInterval)
		assert.Equal(t, cfg.HaskKey, hashKey)

	})

	t.Run("check error", func(t *testing.T) {
		unset, _ := pkg.SetEnv(t, "POLL_INTERVAL", "10")
		err := DefaultConfig.initEnvArgs()
		require.Error(t, err)
		unset()

		unset, _ = pkg.SetEnv(t, "REPORT_INTERVAL", "10")
		err = DefaultConfig.initEnvArgs()
		require.Error(t, err)
		unset()

	})
}

func Test_initCmdFlags(t *testing.T) {
	t.Run("test default args", func(t *testing.T) {
		cfg := DefaultConfig
		err := cfg.initCmdFlagsWithArgs([]string{})
		assert.NoError(t, err)
		assert.EqualValues(t, DefaultConfig, cfg)

	})

	t.Run("test override", func(t *testing.T) {
		addr := "localhost:1234"
		hashKey := "someKey"

		cfg := DefaultConfig
		err := cfg.initCmdFlagsWithArgs([]string{"-a", addr, "-p", "100s", "-r", "1s", "-k", hashKey})
		assert.NoError(t, err)

		targetCfg := Config{
			Address:        addr,
			ReportInterval: 1 * time.Second,
			PollInterval:   100 * time.Second,
			HaskKey:        hashKey,
			LogLevel:       DefaultLogLevel,
		}
		assert.EqualValues(t, targetCfg, cfg)

	})
}
