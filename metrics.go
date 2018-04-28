package wrapper

var (
	MetricClientCall         = "client.call"
	MetricServerCall         = "server.call"
	MetricClientCallDuration = "client.call.duration"
	MetricServerCallDuration = "server.call.duration"
	MetricClientCallErr      = "client.call.err"
	MetricServerCallErr      = "server.call.err"
)

//MetricReporter define the reporter to report metrics
type MetricReporter interface {
	Meter(name string, value int64)
	Histogram(name string, value int64)
	Gauge(name string, value int64)
}
