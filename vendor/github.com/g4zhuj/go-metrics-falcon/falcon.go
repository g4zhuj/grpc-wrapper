package falconmetrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	gometrics "github.com/rcrowley/go-metrics"
)

//DefaultFalcon default falcon created by default config
var DefaultFalcon = NewFalcon(&DefaultFalconConfig)

type FalconMetric struct {
	Endpoint  string      `json:"endpoint"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Step      int64       `json:"step"`
	Type      string      `json:"counterType"`
	Tags      string      `json:"tags"`
	Timestamp int64       `json:"timestamp"`
}

type Falcon struct {
	cfg *FalconConfig
}

func NewFalcon(cfg *FalconConfig) *Falcon {
	return &Falcon{
		cfg: cfg,
	}
}

//NewFalconMetric create open-falcon metric with args
func NewFalconMetric(metricType, valueType, name, endpoint string, step int64, value interface{}) *FalconMetric {
	tags := NewDefaultTags()
	MetricType.Set(tags, metricType)
	ValueType.Set(tags, valueType)
	return &FalconMetric{
		Endpoint:  endpoint,
		Metric:    name,
		Value:     value,
		Step:      step,
		Type:      "GAUGE",
		Tags:      tags.ToStr(),
		Timestamp: time.Now().Unix(),
	}
}

func ReportRegistry(r gometrics.Registry) {
	DefaultFalcon.ReportRegistry(r)
}

func (f *Falcon) post(ms []*FalconMetric) error {
	mb, err := json.Marshal(ms)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewBuffer(mb)
	_, err = http.Post(f.cfg.HostName, "application/json", bodyReader)
	if err != nil {
		return err
	}

	return nil
}

func (f *Falcon) print(ms []*FalconMetric) error {
	for _, m := range ms {
		mb, err := json.Marshal(m)
		if err != nil {
			return err
		}
		fmt.Printf("falcon metric: %v\n", string(mb))
	}
	return nil
}

func (f *Falcon) ReportRegistry(r gometrics.Registry) {
	for _ = range time.Tick(time.Duration(f.cfg.Step) * time.Second) {
		r.Each(func(name string, i interface{}) {
			var fmetrics []*FalconMetric
			switch metric := i.(type) {
			case gometrics.Counter:
				mcount := NewFalconMetric(MetricCounter, ValueCount, name, f.cfg.EndPoint, f.cfg.Step, metric.Count())
				fmetrics = append(fmetrics, mcount)

			case gometrics.GaugeFloat64:
				mcount := NewFalconMetric(MetricGaugeFloat64, Value, name, f.cfg.EndPoint, f.cfg.Step, metric.Value())
				fmetrics = append(fmetrics, mcount)

			case gometrics.Histogram:
				h := metric.Snapshot()
				ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99})
				mcount := NewFalconMetric(MetricGaugeFloat64, ValueCount, name, f.cfg.EndPoint, f.cfg.Step, h.Count())
				fmetrics = append(fmetrics, mcount)
				mmin := NewFalconMetric(MetricGaugeFloat64, ValueMin, name, f.cfg.EndPoint, f.cfg.Step, h.Min())
				fmetrics = append(fmetrics, mmin)
				mmax := NewFalconMetric(MetricGaugeFloat64, ValueMax, name, f.cfg.EndPoint, f.cfg.Step, h.Max())
				fmetrics = append(fmetrics, mmax)
				mmean := NewFalconMetric(MetricGaugeFloat64, ValueMean, name, f.cfg.EndPoint, f.cfg.Step, h.Mean())
				fmetrics = append(fmetrics, mmean)
				mmedian := NewFalconMetric(MetricGaugeFloat64, ValueMedian, name, f.cfg.EndPoint, f.cfg.Step, ps[0])
				fmetrics = append(fmetrics, mmedian)
				m75 := NewFalconMetric(MetricGaugeFloat64, Value75, name, f.cfg.EndPoint, f.cfg.Step, ps[1])
				fmetrics = append(fmetrics, m75)
				m95 := NewFalconMetric(MetricGaugeFloat64, Value95, name, f.cfg.EndPoint, f.cfg.Step, ps[2])
				fmetrics = append(fmetrics, m95)
				m99 := NewFalconMetric(MetricGaugeFloat64, Value99, name, f.cfg.EndPoint, f.cfg.Step, ps[3])
				fmetrics = append(fmetrics, m99)

			case gometrics.Meter:
				m := metric.Snapshot()
				mcount := NewFalconMetric(MetricMeter, ValueCount, name, f.cfg.EndPoint, f.cfg.Step, m.Count())
				fmetrics = append(fmetrics, mcount)
				mrate1 := NewFalconMetric(MetricMeter, ValueRate1, name, f.cfg.EndPoint, f.cfg.Step, m.Rate1())
				fmetrics = append(fmetrics, mrate1)
				mrate5 := NewFalconMetric(MetricMeter, ValueRate5, name, f.cfg.EndPoint, f.cfg.Step, m.Rate5())
				fmetrics = append(fmetrics, mrate5)
				mrate15 := NewFalconMetric(MetricMeter, ValueRate15, name, f.cfg.EndPoint, f.cfg.Step, m.Rate15())
				fmetrics = append(fmetrics, mrate15)
				mratemean := NewFalconMetric(MetricMeter, ValueRateMean, name, f.cfg.EndPoint, f.cfg.Step, m.RateMean())
				fmetrics = append(fmetrics, mratemean)

			case gometrics.Timer:
				du := float64(time.Millisecond)
				t := metric.Snapshot()
				ps := t.Percentiles([]float64{0.5, 0.75, 0.95, 0.99})
				mcount := NewFalconMetric(MetricTimer, ValueCount, name, f.cfg.EndPoint, f.cfg.Step, t.Count())
				fmetrics = append(fmetrics, mcount)
				mmin := NewFalconMetric(MetricTimer, ValueMin, name, f.cfg.EndPoint, f.cfg.Step, float64(t.Min())/du)
				fmetrics = append(fmetrics, mmin)
				mmax := NewFalconMetric(MetricTimer, ValueMax, name, f.cfg.EndPoint, f.cfg.Step, float64(t.Max())/du)
				fmetrics = append(fmetrics, mmax)
				mmean := NewFalconMetric(MetricTimer, ValueMean, name, f.cfg.EndPoint, f.cfg.Step, t.Mean()/du)
				fmetrics = append(fmetrics, mmean)
				mmedian := NewFalconMetric(MetricTimer, ValueMedian, name, f.cfg.EndPoint, f.cfg.Step, ps[0]/du)
				fmetrics = append(fmetrics, mmedian)
				m75 := NewFalconMetric(MetricTimer, Value75, name, f.cfg.EndPoint, f.cfg.Step, ps[1]/du)
				fmetrics = append(fmetrics, m75)
				m95 := NewFalconMetric(MetricTimer, Value95, name, f.cfg.EndPoint, f.cfg.Step, ps[2]/du)
				fmetrics = append(fmetrics, m95)
				m99 := NewFalconMetric(MetricTimer, Value99, name, f.cfg.EndPoint, f.cfg.Step, ps[3]/du)
				fmetrics = append(fmetrics, m99)
				mrate1 := NewFalconMetric(MetricTimer, ValueRate1, name, f.cfg.EndPoint, f.cfg.Step, metric.Rate1())
				fmetrics = append(fmetrics, mrate1)
				mrate5 := NewFalconMetric(MetricTimer, ValueRate5, name, f.cfg.EndPoint, f.cfg.Step, metric.Rate5())
				fmetrics = append(fmetrics, mrate5)
				mrate15 := NewFalconMetric(MetricTimer, ValueRate15, name, f.cfg.EndPoint, f.cfg.Step, metric.Rate15())
				fmetrics = append(fmetrics, mrate15)
				mratemean := NewFalconMetric(MetricTimer, ValueRateMean, name, f.cfg.EndPoint, f.cfg.Step, metric.RateMean())
				fmetrics = append(fmetrics, mratemean)
			}

			if f.cfg.Debug {
				f.print(fmetrics)
			}
			f.post(fmetrics)
		})
	}
}
