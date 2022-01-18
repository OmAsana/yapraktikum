package agent

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type unsetFunc func()

func setEnv(t *testing.T, key, value string) (unsetFunc, error) {
	t.Helper()

	err := os.Setenv(key, value)
	if err != nil {
		return nil, err
	}

	return func() {
		_ = os.Unsetenv(key)

	}, nil

}
func TestInitConfig(t *testing.T) {
	t.Run("check default", func(t *testing.T) {
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
		setEnv(t, "ADDRESS", newAddress)
		setEnv(t, "POLL_INTERVAL", strconv.FormatInt(newPollInterval, 10))
		setEnv(t, "REPORT_INTERVAL", strconv.FormatInt(newReportInterval, 10))

		cfg, err := InitConfig()
		require.NoError(t, err)
		assert.Equal(t, cfg.Address, newAddress)
		assert.Equal(t, cfg.PollInterval, newPollInterval)
		assert.Equal(t, cfg.ReportInterval, newReportInterval)

	})

	t.Run("check error", func(t *testing.T) {
		unset, _ := setEnv(t, "POLL_INTERVAL", "10s")
		_, err := InitConfig()
		require.Error(t, err)
		unset()

		unset, _ = setEnv(t, "REPORT_INTERVAL", "10s")
		_, err = InitConfig()
		require.Error(t, err)
		unset()

	})
}
