package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Metric interface {
	Name() string
}

type Gauge struct {
	name     string
	value    int64
	getValue func() int64
}

func NewGauge(name string, getValue func() int64) *Gauge {
	g := &Gauge{
		name:     name,
		getValue: getValue,
	}
	return g
}

func (gauge *Gauge) Name() string {
	return gauge.name
}

//func (gauge *Gauge) Mark(v int64) {
//	atomic.StoreInt64(&gauge.value, v)
//}

func (gauge *Gauge) Value() int64 {
	return gauge.getValue()
}

type Meter struct {
	name       string
	value      int64
	recordTime int64
}

func NewMeter(name string) *Meter {
	m := &Meter{
		name: name,
	}
	return m
}

func (meter *Meter) Name() string {
	return meter.name
}

func (meter *Meter) Update(v int64) {
	atomic.AddInt64(&meter.value, v)
	atomic.CompareAndSwapInt64(&meter.recordTime, 0, time.Now().Unix())
}

func (meter *Meter) Value() int64 {
	value := atomic.SwapInt64(&meter.value, 0)
	currentTime := time.Now().Unix()
	rTime := atomic.SwapInt64(&meter.recordTime, currentTime)
	if currentTime - rTime == 0{
		return 0
	}
	return value / (currentTime - rTime)
}

type Counter struct {
	name  string
	value int64
}

func NewCounter(name string) *Counter {
	c := &Counter{
		name: name,
	}
	return c
}

func (counter *Counter) Name() string {
	return counter.name
}

func (counter *Counter) Incr(value int64) {
	atomic.AddInt64(&counter.value, value)
}

func (counter *Counter) Decr(value int64) {
	atomic.AddInt64(&counter.value, (0 - value))
}

func (counter *Counter) Value() int64 {
	return counter.value
}

type InfoSheet struct {
	name    string
	info    *sync.Map
	getInfo func(infoValues *sync.Map)
}

func NewInfoSheet(name string, getInfo func(infoValues *sync.Map)) *InfoSheet {
	infos := &InfoSheet{
		name:    name,
		info:    &sync.Map{},
		getInfo: getInfo,
	}
	return infos
}

func (info *InfoSheet) Name() string {
	return info.name
}

func (info *InfoSheet) AddInfo(key string, value interface{}) {
	info.info.Store(key, value)
}

func (info *InfoSheet) Info() *sync.Map {
	info.getInfo(info.info)
	return info.info
}
