package agent

import (
	"net/url"
	"strings"
	"time"

	"github.com/OmAsana/yapraktikum/internal/logging"
)

type Option func(*Agent) error

func WithAddress(address string) Option {
	return func(agent *Agent) error {
		if !strings.HasPrefix(address, "http://") {
			address = "http://" + address

		}

		defaultBaseURL, err := url.Parse(address)
		if err != nil {
			return err
		}
		agent.cfg.BaseURL = defaultBaseURL
		return nil
	}
}

func WithReportInterval(t time.Duration) Option {
	return func(agent *Agent) error {
		agent.cfg.ReportInterval = t
		return nil
	}
}

func WithLogger(l *logging.Logger) Option {
	return func(agent *Agent) error {
		agent.log = l
		return nil
	}
}

func WithPollInterval(t time.Duration) Option {
	return func(agent *Agent) error {
		agent.cfg.PollInterval = t
		return nil
	}
}

func WithHashKey(key string) Option {
	return func(agent *Agent) error {
		agent.cfg.HashKey = key
		return nil
	}
}
