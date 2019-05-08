package plugins

import (
	"context"
	"sync"
	"time"

	falconmetrics "github.com/g4zhuj/go-metrics-falcon"
	wrapper "github.com/g4zhuj/grpc-wrapper"
	metrics "github.com/rcrowley/go-metrics"
	"google.golang.org/grpc"
)

const (
	MeterPre = iota
	HistogramPre
	GaugePre
)

type FalconReporter struct {
	mu       sync.RWMutex
	falcon   *falconmetrics.Falcon
	registry metrics.Registry
}

func NewDefaultFalconReporter() *FalconReporter {
	cfg := falconmetrics.DefaultFalconConfig
	falcon := falconmetrics.NewFalcon(&cfg)
	go falcon.ReportRegistry(metrics.DefaultRegistry)

	fr := &FalconReporter{
		falcon:   falcon,
		registry: metrics.NewRegistry(),
	}
	go fr.falcon.ReportRegistry(fr.registry)
	return fr
}

func (f *FalconReporter) Meter(name string, value int64) {
	meter := metrics.GetOrRegisterMeter(name, f.registry)
	if meter != nil {
		meter.Mark(value)
	}
}

func (f *FalconReporter) Histogram(name string, value int64) {
	histogram := metrics.GetOrRegisterHistogram(name, f.registry, metrics.NewExpDecaySample(1028, 0.015))
	if histogram != nil {
		histogram.Update(value)
	}
}

func (f *FalconReporter) Gauge(name string, value int64) {
	gauge := metrics.GetOrRegisterGauge(name, f.registry)
	if gauge != nil {
		gauge.Update(value)
	}
}

//MetricClientInterceptor  rewrite client's interceptor to report metrics
func MetricClientInterceptor(reporter wrapper.MetricReporter) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ts := time.Now()
		defer func() {
			//report time duration in millisecond
			duration := time.Since(ts) / time.Millisecond
			reporter.Histogram(wrapper.MetricClientCallDuration+method, int64(duration))
		}()
		reporter.Meter(wrapper.MetricClientCall+method, 1)
		err := invoker(ctx, method, req, resp, cc, opts...)
		if err != nil {
			reporter.Meter(wrapper.MetricClientCallErr+method, 1)
		}
		return err
	}
}

//MetricServerInterceptor rewrite server's interceptor to report metric
func MetricServerInterceptor(reporter wrapper.MetricReporter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		ts := time.Now()
		defer func() {
			//report time duration in millisecond
			duration := time.Since(ts) / time.Millisecond
			reporter.Histogram(wrapper.MetricServerCallDuration+info.FullMethod, int64(duration))
		}()

		reporter.Meter(wrapper.MetricServerCall+info.FullMethod, 1)
		resp, err = handler(ctx, req)
		if err != nil {
			reporter.Meter(wrapper.MetricServerCallErr+info.FullMethod, 1)
		}
		return
	}
}
