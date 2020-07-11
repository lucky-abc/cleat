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

func (mr *MetricRegistry) Metrics() map[string]interface{} {
	ms := make(map[string]interface{}, 0)
	mr.metrics.Range(func(key, value interface{}) bool {
		ms[key.(string)] = value
		return true
	})
	return ms
}
