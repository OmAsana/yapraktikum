package agent

import (
	"fmt"
	"os"
	"testing"
	"time"

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
		assert.Equal(t, cfg.PollInterval, 2*time.Second)
		assert.Equal(t, cfg.ReportInterval, 10*time.Second)
		assert.Equal(t, cfg.Address, "127.0.0.1:8080")
	})

	t.Run("check overrides", func(t *testing.T) {
		newAddress := "127.0.0.1:1234"
		newPollInterval := 2 * time.Second
		newReportInterval := 5 * time.Second
		unsetAdd, _ := pkg.SetEnv(t, "ADDRESS", newAddress)
		unsetPoll, _ := pkg.SetEnv(t, "POLL_INTERVAL", fmt.Sprintf("%ds", int(newPollInterval.Seconds())))
		unsetReport, _ := pkg.SetEnv(t, "REPORT_INTERVAL", fmt.Sprintf("%ds", int(newReportInterval.Seconds())))
		//unsetRerport, _ := pkg.SetEnv(t, "REPORT_INTERVAL", strconv.FormatInt(newReportInterval, 10))
		defer func() {
			unsetReport()
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
		unset, _ := pkg.SetEnv(t, "POLL_INTERVAL", "10")
		_, err := InitConfig()
		require.Error(t, err)
		unset()

		unset, _ = pkg.SetEnv(t, "REPORT_INTERVAL", "10")
		_, err = InitConfig()
		require.Error(t, err)
		unset()

	})
}
