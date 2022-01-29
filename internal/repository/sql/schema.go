package sql

type Counter struct {
	Name  string
	Value int64
}

type Gauge struct {
	Name  string
	Delta float64
}
