package metrics

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"reflect"
	"strconv"
	"time"
)

type Reporter interface {
	Start()
	Stop()
}

type LogFileReporter struct {
	mr             *MetricRegistry
	reportDuration time.Duration
	logger         *zap.Logger
	tiker          *time.Ticker
	printBuffer    bytes.Buffer
}

func NewLogFileReporter(logger *zap.Logger, mr *MetricRegistry, duration int) *LogFileReporter {
	d, err := time.ParseDuration(strconv.Itoa(duration) + "s")
	if err != nil {
		return nil
	}
	reporter := &LogFileReporter{
		mr:             mr,
		logger:         logger,
		reportDuration: d,
	}
	return reporter
}

func (lfr *LogFileReporter) Start() {
	lfr.tiker = time.NewTicker(lfr.reportDuration)
	go func() {
		for {
			<-lfr.tiker.C
			lfr.report()
		}
	}()
}

func (lfr *LogFileReporter) report() {
	metrics := lfr.mr.Metrics()
	for name, metric := range metrics {
		metricType := reflect.ValueOf(metric).Elem().Type()
		switch metricType.Name() {
		case "Gauge":
			lfr.reportGauge(name, metric.(*Gauge))
		case "Meter":
			lfr.reportMeter(name, metric.(*Meter))
		case "Counter":
			lfr.reportCounter(name, metric.(*Counter))
		case "InfoSheet":
			lfr.reportInfoSheet(name, metric.(*InfoSheet))
		}
	}
	lfr.logger.Info(lfr.printBuffer.String())
	lfr.printBuffer.Reset()
}

func (lfr *LogFileReporter) reportGauge(name string, metric *Gauge) {
	lfr.printBuffer.WriteString(fmt.Sprintf("\nGauge-(%s): %d", name, metric.Value()))
}
func (lfr *LogFileReporter) reportMeter(name string, metric *Meter) {
	lfr.printBuffer.WriteString(fmt.Sprintf("\nMeter-(%s): %d /sec", name, metric.Value()))
}
func (lfr *LogFileReporter) reportCounter(name string, metric *Counter) {
	lfr.printBuffer.WriteString(fmt.Sprintf("\nCounter-(%s): %d", name, metric.Value()))
}
func (lfr *LogFileReporter) reportInfoSheet(name string, metric *InfoSheet) {
	lfr.printBuffer.WriteString(fmt.Sprintf("\nInfoSheet-(%s): ", name))
	metric.Info().Range(func(key interface{}, value interface{}) bool {
		lfr.printBuffer.WriteString(fmt.Sprintf("\n\t{%s: %v}  ", key, value))
		return true
	})
}

func (lfr *LogFileReporter) Stop() {
	lfr.tiker.Stop()
}
