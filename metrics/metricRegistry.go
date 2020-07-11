package metrics

import "sync"

type MetricRegistry struct {
	metrics   sync.Map
	reporters []Reporter
}

func NewMetricRegstry() *MetricRegistry {
	mr := &MetricRegistry{
		reporters: make([]Reporter, 0),
		metrics:   sync.Map{},
	}
	return mr
}
func (mr *MetricRegistry) RegisterMetric(metric Metric) {
	mr.metrics.Store(metric.Name(), metric)
}

func (mr *MetricRegistry) RegisterReporter(reporter Reporter) {
	mr.reporters = append(mr.reporters, reporter)
}
func (mr *MetricRegistry) GetMeter(name string) *Meter {
	metric, _ := mr.metrics.Load(name)
	return metric.(*Meter)
}
func (mr *MetricRegistry) GetCounter(name string) *Counter {
	metric, _ := mr.metrics.Load(name)
	return metric.(*Counter)
}
func (mr *MetricRegistry) GetGauge(name string) *Gauge {
	metric, _ := mr.metrics.Load(name)
	return metric.(*Gauge)
}
func (mr *MetricRegistry) GetInfoSheet(name string) *InfoSheet {
	metric, _ := mr.metrics.Load(name)
	return metric.(*InfoSheet)
}

func (mr *MetricRegistry) Metrics() map[string]interface{} {
	ms := make(map[string]interface{}, 0)
	mr.metrics.Range(func(key, value interface{}) bool {
		ms[key.(string)] = value
		return true
	})
	return ms
}
