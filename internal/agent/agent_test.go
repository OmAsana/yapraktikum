package agent

import (
	"context"
)

func ExampleAgent_Server() {
	ctx := context.Background()
	agent := NewDefaultAgent()
	agent.Server(ctx)
}
