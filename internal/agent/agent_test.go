package agent

import (
	"context"
	"testing"
)

func TestAgent_Server(t *testing.T) {
	ctx := context.Background()
	agent := NewDefaultAgent()
	agent.Server(ctx)
}
