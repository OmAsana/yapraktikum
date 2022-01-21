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

var defaultCfg = &Config{
	Address:        DefaultAddress,
	ReportInterval: DefaultReportInterval,
	PollInterval:   DefaultPollInterval,
}

func TestAgentInitConfig(t *testing.T) {
	t.Run("check default", func(t *testing.T) {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("REPORT_INTERVAL")
		cfg, err := initEnvArgs(*defaultCfg)

		require.NoError(t, err)
		assert.Equal(t, cfg.PollInterval, defaultCfg.PollInterval)
		assert.Equal(t, cfg.ReportInterval, defaultCfg.ReportInterval)
		assert.Equal(t, cfg.Address, defaultCfg.Address)
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

		cfg, err := initEnvArgs(*defaultCfg)
		require.NoError(t, err)
		assert.Equal(t, cfg.Address, newAddress)
		assert.Equal(t, cfg.PollInterval, newPollInterval)
		assert.Equal(t, cfg.ReportInterval, newReportInterval)

	})

	t.Run("check error", func(t *testing.T) {
		unset, _ := pkg.SetEnv(t, "POLL_INTERVAL", "10")
		_, err := initEnvArgs(*defaultCfg)
		require.Error(t, err)
		unset()

		unset, _ = pkg.SetEnv(t, "REPORT_INTERVAL", "10")
		_, err = initEnvArgs(*defaultCfg)
		require.Error(t, err)
		unset()

	})
}

func Test_initCmdFlags(t *testing.T) {
	t.Run("test default args", func(t *testing.T) {
		cfg := initCmdFlagsWithArgs([]string{})
		assert.EqualValues(t, defaultCfg, cfg)

	})

	t.Run("test override", func(t *testing.T) {
		cfg := initCmdFlagsWithArgs([]string{"-a", "localhost:1234", "-p", "100s", "-r", "1s"})

		targetCfg := Config{
			Address:        "localhost:1234",
			ReportInterval: 1 * time.Second,
			PollInterval:   100 * time.Second,
		}
		assert.EqualValues(t, &targetCfg, cfg)

	})
}
