package wrapper

var (
	MetricClientCall         = "client"
	MetricServerCall         = "server"
	MetricClientCallDuration = "client.duration"
	MetricServerCallDuration = "server.duration"
	MetricClientCallErr      = "client.err"
	MetricServerCallErr      = "server.err"
)

//MetricReporter define the reporter to report metrics
type MetricReporter interface {
	Meter(name string, value int64)
	Histogram(name string, value int64)
	Gauge(name string, value int64)
}
