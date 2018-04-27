package wrapper

type Metric interface {
	Meter(name string, value int64) error
	Histogram(name string, value int64) error
	Gauge(name string, value int64) error
}
