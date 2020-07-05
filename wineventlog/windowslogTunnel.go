package wineventlog

import (
	"github.com/lucky-abc/cleat/config"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/output"
	"github.com/lucky-abc/cleat/record"
	"github.com/lucky-abc/cleat/tunnel"
	"github.com/lucky-abc/cleat/wineventlog/wineventapi"
)

type WindowslogTunnel struct {
	tunnel.TunnelModel
	queue chan string
}

func NewWindowslogTunnel(ck *record.RecordPoint) *WindowslogTunnel {
	available, _ := wineventapi.IsAvailable()
	if !available {
		logger.Loggers().Warn("Windows API is not supported on the current platform")
		return nil
	}
	q := make(chan string, 1024)
	winlogSource := NewWinLogSource(q, ck)
	udpServerIP := config.Config().GetString("output.udp.serverIP")
	udpServerPort := config.Config().GetInt("output.udp.serverPort")
	udpOutput := output.NewUDPOutput(udpServerIP, udpServerPort, q)
	tunnel := &WindowslogTunnel{
		TunnelModel: tunnel.TunnelModel{
			Source: winlogSource,
			Output: udpOutput,
		},
		queue: q,
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
