package metrics

import (
	"fmt"

	"github.com/OmAsana/yapraktikum/internal/handlers"
)

type Counter struct {
	Name  string
	Value int64
}

func (c Counter) String() string {
	return fmt.Sprintf("<Counter: Name: %s, Value: %d>", c.Name, c.Value)
}

func (c Counter) IsValid() error {
	if c.Value < 0 {
		return fmt.Errorf("counter can not be negative")
	}
	return nil
}

func CounterToHandlerScheme(c Counter) handlers.Metrics {
	return handlers.Metrics{
		ID:    c.Name,
		MType: "counter",
		Delta: &c.Value,
		Value: nil,
	}
}

func CounterFromHandler(h handlers.Metrics) Counter {
	return Counter{
		Name:  h.ID,
		Value: *h.Delta,
	}
}
