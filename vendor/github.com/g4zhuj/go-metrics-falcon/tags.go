package falconmetrics

import (
	"strings"
)

var (
	//ProjectName name of project
	ProjectName = stringTagName("project")
	MetricType  = stringTagName("metricType")
	ValueType   = stringTagName("valueType")

	MetricCounter      = "counter"
	MetricGaugeFloat64 = "gaugeFloat64"
	MetricHistogram    = "histogram"
	MetricMeter        = "meter"
	MetricTimer        = "timer"

	ValueCount    = "count"
	ValueMin      = "min"
	ValueMax      = "max"
	ValueMean     = "mean"
	ValueMedian   = "median"
	Value75       = "75%"
	Value95       = "95%"
	Value99       = "99%"
	ValueRate1    = "rate1"
	ValueRate5    = "rate5"
	ValueRate15   = "rate15"
	ValueRateMean = "ratemean"
	Value         = "value"
)

type stringTagName string

func (tag stringTagName) Set(t *Tags, value string) {
	t.SetTag(string(tag), value)
}

type Tags struct {
	tags map[string]string
}

func NewDefaultTags() *Tags {
	t := &Tags{
		tags: make(map[string]string),
	}
	ProjectName.Set(t, defaultProjectName())
	return t
}

func (t *Tags) SetTag(k, v string) {
	t.tags[k] = v
}

func (t *Tags) ToStr() string {
	var ts []string
	for k, v := range t.tags {
		ts = append(ts, k+"="+v)
	}
	return strings.Join(ts, ",")
}
