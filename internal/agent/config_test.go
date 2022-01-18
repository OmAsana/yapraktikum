package agent

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/OmAsana/yapraktikum/internal/pkg"
)

func TestAgentInitConfig(t *testing.T) {
	t.Run("check default", func(t *testing.T) {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("REPORT_INTERVAL")
		cfg, err := InitConfig()

		require.NoError(t, err)
		assert.Equal(t, cfg.PollInterval, int64(2))
		assert.Equal(t, cfg.ReportInterval, int64(10))
		assert.Equal(t, cfg.Address, "127.0.0.1:8080")
	})

	t.Run("check overrides", func(t *testing.T) {
		newAddress := "127.0.0.1:1234"
		newPollInterval := int64(2)
		newReportInterval := int64(5)
		unsetAdd, _ := pkg.SetEnv(t, "ADDRESS", newAddress)
		unsetPoll, _ := pkg.SetEnv(t, "POLL_INTERVAL", strconv.FormatInt(newPollInterval, 10))
		unsetRerport, _ := pkg.SetEnv(t, "REPORT_INTERVAL", strconv.FormatInt(newReportInterval, 10))
		defer func() {
			unsetRerport()
			unsetPoll()
			unsetAdd()
		}()

		cfg, err := InitConfig()
		require.NoError(t, err)
		assert.Equal(t, cfg.Address, newAddress)
		assert.Equal(t, cfg.PollInterval, newPollInterval)
		assert.Equal(t, cfg.ReportInterval, newReportInterval)

	})

	t.Run("check error", func(t *testing.T) {
		unset, _ := pkg.SetEnv(t, "POLL_INTERVAL", "10s")
		_, err := InitConfig()
		require.Error(t, err)
		unset()

		unset, _ = pkg.SetEnv(t, "REPORT_INTERVAL", "10s")
		_, err = InitConfig()
		require.Error(t, err)
		unset()

	})
}