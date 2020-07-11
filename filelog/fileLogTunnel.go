package filelog

import (
	"github.com/lucky-abc/cleat/config"
	"github.com/lucky-abc/cleat/metrics"
	"github.com/lucky-abc/cleat/output"
	"github.com/lucky-abc/cleat/record"
	"github.com/lucky-abc/cleat/tunnel"
)

type FilelogTunnel struct {
	tunnel.TunnelModel
	queue chan string
}

func NewFilelogTunnel(ck *record.RecordPoint, metricRegistry *metrics.MetricRegistry) *FilelogTunnel {
	var tunnelName = "filelog"
	q := make(chan string, 1024)
	metricGauge := metrics.NewGauge("filelog-channal-size", func() int64 {
		return int64(len(q))
	})
	metricRegistry.RegisterMetric(metricGauge)
	outputRecordTotalMetric := metrics.NewCounter(tunnelName + "-output-record-total")
	metricRegistry.RegisterMetric(outputRecordTotalMetric)

	s := NewFileLogSource(q, ck, metricRegistry)
	udpServerIP := config.Config().GetString("output.udp.serverIP")
	udpServerPort := config.Config().GetInt("output.udp.serverPort")
	o := output.NewUDPOutput(udpServerIP, udpServerPort, q, metricRegistry, tunnelName)

	tunnel := &FilelogTunnel{
		queue: q,
		TunnelModel: tunnel.TunnelModel{
			Source: s,
			Output: o,
		},
	}
	return tunnel
}

func (ft *FilelogTunnel) Start() {
	ft.Output.Start()
	ft.Source.Start()
}

func (ft *FilelogTunnel) Transfer() {
	go ft.Output.Process()
}

func (ft *FilelogTunnel) Stop() {
	ft.Source.Stop()
	close(ft.queue)
	ft.Output.Stop()
}
