package output

import (
	"fmt"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/metrics"
	"net"
	"sync"
)

type UDPOutput struct {
	udpServer         string
	udpServerPort     int
	queue             chan string
	udpConn           *net.UDPConn
	waitGroup         sync.WaitGroup
	sendMeter         *metrics.Meter
	recordTotalMetric *metrics.Counter
}

func NewUDPOutput(udpServer string, udpServerPort int, queue chan string, metricRegistry *metrics.MetricRegistry, tunnelName string) *UDPOutput {
	output := &UDPOutput{
		udpServer:     udpServer,
		udpServerPort: udpServerPort,
		queue:         queue,
	}
	sendMeter := metrics.NewMeter(tunnelName + "-updoutput-rate")
	metricRegistry.RegisterMetric(sendMeter)
	output.sendMeter = sendMeter
	output.recordTotalMetric = metricRegistry.GetCounter(tunnelName + "-output-record-total")
	return output
}

func (output *UDPOutput) Start() {
	var udpAddr *net.UDPAddr
	var err error
	addrs := fmt.Sprintf("%s:%d", output.udpServer, output.udpServerPort)
	logger.Loggers().Info("udp server address：", addrs)
	udpAddr, err = net.ResolveUDPAddr("udp", addrs)
	if err != nil {
		logger.Loggers().Error("udp adress resolve error:", err)
		return
	}
	output.udpConn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		logger.Loggers().Error("connect udp server error:", err)
		return
	}
}

func (output *UDPOutput) Process() {
	output.waitGroup.Add(1)
	defer output.waitGroup.Done()
	for data := range output.queue {
		_, err := output.udpConn.Write([]byte(data))
		if err != nil {
			logger.Loggers().Error("upd send error：", err)
			return
		}
		output.sendMeter.Update(1)
		output.recordTotalMetric.Incr(1)
	}
}

func (output *UDPOutput) Stop() {
	output.waitGroup.Wait()
	if output.udpConn != nil {
		output.udpConn.Close()
	}
	logger.Loggers().Info("udp output closed")
}
