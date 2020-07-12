package output

import (
	"bytes"
	"fmt"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/metrics"
	"net"
	"strings"
	"sync"
)

type TCPOutput struct {
	server            string
	serverPort        int
	queue             chan string
	tcpConn           *net.TCPConn
	waitGroup         sync.WaitGroup
	sendMeter         *metrics.Meter
	recordTotalMetric *metrics.Counter
	dataBuffer        bytes.Buffer
}

func NewTCPOutput(tcpConfig *TCPOutputConfig, queue chan string, metricRegistry *metrics.MetricRegistry, tunnelName string) *TCPOutput {
	output := &TCPOutput{
		server:     tcpConfig.Server,
		serverPort: tcpConfig.ServerPort,
		queue:      queue,
	}
	sendMeter := metrics.NewMeter(tunnelName + "-tcpoutput-rate")
	metricRegistry.RegisterMetric(sendMeter)
	output.sendMeter = sendMeter
	output.recordTotalMetric = metricRegistry.GetCounter(tunnelName + "-output-record-total")
	return output
}

func (output *TCPOutput) Start() {
	var addr *net.TCPAddr
	var err error
	addrs := fmt.Sprintf("%s:%d", output.server, output.serverPort)
	logger.Loggers().Info("tcp server address：", addrs)
	addr, err = net.ResolveTCPAddr("tcp", addrs)
	if err != nil {
		logger.Loggers().Error("tcp adress resolve error:", err)
		return
	}
	output.tcpConn, err = net.DialTCP("tcp", nil, addr)
	if err != nil {
		logger.Loggers().Error("connect tcp server error:", err)
		return
	}
}

func (output *TCPOutput) Process() {
	output.waitGroup.Add(1)
	defer output.waitGroup.Done()
	for data := range output.queue {
		output.dataBuffer.Reset()
		output.dataBuffer.WriteString(data)
		if !strings.HasSuffix(data, "\n") {
			output.dataBuffer.WriteString("\n")
		}
		_, err := output.tcpConn.Write(output.dataBuffer.Bytes())
		if err != nil {
			logger.Loggers().Error("tcp send error：", err)
			return
		}
		output.sendMeter.Update(1)
		output.recordTotalMetric.Incr(1)
	}
}

func (output *TCPOutput) Stop() {
	output.waitGroup.Wait()
	if output.tcpConn != nil {
		output.tcpConn.Close()
	}
	logger.Loggers().Info("tcp output closed")
}
