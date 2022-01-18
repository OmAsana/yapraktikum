package agent

import (
	"net/url"
	"strings"
	"time"
)

type AgentOption func(*Agent) error

func WithAddress(address string) AgentOption {
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

func WithReportInterval(seconds int64) AgentOption {
	return func(agent *Agent) error {
		agent.cfg.ReportInterval = time.Second * time.Duration(seconds)
		return nil
	}
}

func WithPollInterval(seconds int64) AgentOption {
	return func(agent *Agent) error {
		agent.cfg.PollInterval = time.Second * time.Duration(seconds)
		return nil
	}
}
