package plugins

import (
	"sync"

	falconmetrics "github.com/g4zhuj/go-metrics-falcon"
	metrics "github.com/rcrowley/go-metrics"
)

type FalconReporter struct {
	mu       sync.RWMutex
	falcon   *falconmetrics.Falcon
	registry metrics.Registry
}

func NewDefalutFalcon() *FalconReporter {
	cfg := falconmetrics.DefaultFalconConfig
	falcon := falconmetrics.NewFalcon(&cfg)
	falcon.ReportRegistry(metrics.DefaultRegistry)

	return &FalconReporter{
		falcon: falcon,
	}
}
func (f *FalconReporter) Meter(name string, value int64) error {

	return nil
}
func (f *FalconReporter) Histogram(name string, value int64) error {

	return nil
}
func (f *FalconReporter) Gauge(name string, value int64) error {

	return nil
}
