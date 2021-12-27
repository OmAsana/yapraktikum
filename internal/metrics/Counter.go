package metrics

import "fmt"

type Counter struct {
	Name  string
	Value int64
}

func (c Counter) String() string {
	return fmt.Sprintf("<Counter: Name: %s, Value: %d>", c.Name, c.Value)
}
