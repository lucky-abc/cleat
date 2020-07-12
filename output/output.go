package output

import (
	"github.com/lucky-abc/cleat/config"
	"github.com/lucky-abc/cleat/metrics"
	"github.com/pkg/errors"
	"strings"
)

type Output interface {
	Start()
	Process()
	Stop()
}

type TCPOutputConfig struct {
	Server     string
	ServerPort int
}

type UDPOutputConfig struct {
	Server     string
	ServerPort int
}

func BuildOutput(queue chan string, metricRegistry *metrics.MetricRegistry, tunnelName string) (Output, error) {
	outputType, udpConig, tcpConfig, err := parseConfig()
	if err != nil {
		return nil, err
	}
	var output Output
	if outputType == "udp" {
		output = NewUDPOutput(udpConig, queue, metricRegistry, tunnelName)
	} else if outputType == "tcp" {
		output = NewTCPOutput(tcpConfig, queue, metricRegistry, tunnelName)
	}
	return output, nil
}

func parseConfig() (string, *UDPOutputConfig, *TCPOutputConfig, error) {
	stringMap := config.Config().GetStringMap("output")
	for key, value := range stringMap {
		if key == "udp" {
			udpConfig := &UDPOutputConfig{}
			valueMap, ok := value.(map[string]interface{})
			if !ok {
				return "", nil, nil, errors.New("parse output config error")
			}
			for k, v := range valueMap {
				if strings.ToLower(k) == "serverip" {
					udpConfig.Server = v.(string)
				}
				if strings.ToLower(k) == "serverport" {
					udpConfig.ServerPort = v.(int)
				}
			}
			return key, udpConfig, nil, nil
		}
		if key == "tcp" {
			tcpConfig := &TCPOutputConfig{}
			valueMap, ok := value.(map[string]interface{})
			if !ok {
				return "", nil, nil, errors.New("parse output config error")
			}
			for k, v := range valueMap {
				if strings.ToLower(k) == "serverip" {
					tcpConfig.Server = v.(string)
				}
				if strings.ToLower(k) == "serverport" {
					tcpConfig.ServerPort = v.(int)
				}
			}
			return key, nil, tcpConfig, nil
		}
	}
	return "", nil, nil, errors.New("parse output config error")
}
