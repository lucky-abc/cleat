package wineventlog

import (
	"github.com/lucky-abc/cleat/config"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/metrics"
	"github.com/lucky-abc/cleat/output"
	"github.com/lucky-abc/cleat/record"
	"github.com/lucky-abc/cleat/tunnel"
	"github.com/lucky-abc/cleat/wineventlog/wineventapi"
)

type WindowslogTunnel struct {
	tunnel.TunnelModel
	queue      chan string
	tunnelName string
}

func NewWindowslogTunnel(ck *record.RecordPoint, metricRegistry *metrics.MetricRegistry) *WindowslogTunnel {
	available, _ := wineventapi.IsAvailable()
	if !available {
		logger.Loggers().Warn("Windows API is not supported on the current platform")
		return nil
	}
	var tunnelName = "windowevent"
	q := make(chan string, 1024)

	metricGauge := metrics.NewGauge("windowevent-channal-size", func() int64 {
		return int64(len(q))
	})
	metricRegistry.RegisterMetric(metricGauge)
	outputRecordTotalMetric := metrics.NewCounter(tunnelName + "-output-record-total")
	metricRegistry.RegisterMetric(outputRecordTotalMetric)

	winlogSource := NewWinLogSource(q, ck, metricRegistry)
	udpServerIP := config.Config().GetString("output.udp.serverIP")
	udpServerPort := config.Config().GetInt("output.udp.serverPort")
	udpOutput := output.NewUDPOutput(udpServerIP, udpServerPort, q, metricRegistry, tunnelName)

	tunnel := &WindowslogTunnel{
		TunnelModel: tunnel.TunnelModel{
			Source: winlogSource,
			Output: udpOutput,
		},
		queue:      q,
		tunnelName: tunnelName,
	}
	return tunnel
}

func (t *WindowslogTunnel) Start() {
	t.Output.Start()
	t.Source.Start()
}

func (t *WindowslogTunnel) Transfer() {
	go t.Output.Process()
	t.Source.Process()
}

func (t *WindowslogTunnel) Stop() {
	t.Source.Stop()
	t.Output.Stop()
}
