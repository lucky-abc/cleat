package main

import (
	"github.com/lucky-abc/cleat/config"
	"github.com/lucky-abc/cleat/filelog"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/metrics"
	"github.com/lucky-abc/cleat/record"
	"github.com/lucky-abc/cleat/wineventlog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func main() {
	_, configPath, dataPath, logPath := getStartupPath()
	config.InitSystemConfig("config", configPath)
	logger.NewLogger(logPath, config.Config())

	metricRegistry := setupMetrics()

	ck, err := record.NewCheckpoint(filepath.Join(dataPath, "recordpoint"))
	if err != nil {
		logger.Loggers().Error("new checkpoint error:", err)
		return
	}

	winlogTunnel := wineventlog.NewWindowslogTunnel(ck, metricRegistry)
	if winlogTunnel != nil {
		winlogTunnel.Start()
		winlogTunnel.Transfer()
	}

	fileTunnel := filelog.NewFilelogTunnel(ck, metricRegistry)
	fileTunnel.Start()
	fileTunnel.Transfer()

	signalsChan := make(chan os.Signal, 1)
	signal.Notify(signalsChan, os.Interrupt, os.Kill)
	signal := <-signalsChan
	logger.Loggers().Infof("termination signal:%v", signal)
	logger.Loggers().Info("Terminating run. Please wait...")
	if winlogTunnel != nil {
		winlogTunnel.Stop()
	}
	fileTunnel.Stop()
	ck.Close()

	logger.Loggers().Infof("it's over")
}

func setupMetrics() *metrics.MetricRegistry {
	metricRegistry := metrics.NewMetricRegstry()
	bootTime := time.Now()
	infoSheetMetric := metrics.NewInfoSheet("system_overview", func(infoValues *sync.Map) {
		infoValues.Store("bootTime", bootTime.Format("2006-01-02 15:04:05.999"))
		infoValues.Store("runningTime", time.Since(bootTime).String())
	})
	infoSheetMetric.AddInfo("OS", runtime.GOOS)
	metricRegistry.RegisterMetric(infoSheetMetric)
	setupMetricReport(metricRegistry)
	return metricRegistry
}

func setupMetricReport(metricRegistry *metrics.MetricRegistry) {
	reportersConfig, ok := config.Config().Get("metrics.reporters").([]interface{})
	if !ok {
		logger.Loggers().Info("there are no report config")
		return
	}
	for _, reporterConfig := range reportersConfig {
		rc, ok := reporterConfig.(map[interface{}]interface{})
		if !ok {
			logger.Loggers().Info("there are no report config")
			continue
		}
		for k1, v1 := range rc {
			if k1.(string) == "logfile" {
				logfileConfig := v1.(map[interface{}]interface{})
				if !ok {
					logger.Loggers().Info("there are no report config")
					break
				}
				var interval = ""
				var level = ""
				for k2, v2 := range logfileConfig {
					switch k2.(string) {
					case "reportInterval":
						interval = v2.(string)
					case "level":
						level = v2.(string)
					}
				}
				logFileReporter := metrics.NewLogFileReporter(logger.Logger(), level, metricRegistry, interval)
				metricRegistry.RegisterReporter(logFileReporter)
				logFileReporter.Start()
			}
		}
	}
}

func getStartupPath() (appPath, configPath, dataPath, logPath string) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Loggers().Error("get boot path error: %v", err)
		panic("get boot path error")
		return
	}
	dir = filepath.Clean(dir)
	dir = filepath.ToSlash(dir)
	appPath = filepath.Dir(dir)
	configPath = filepath.Join(appPath, "config")
	dataPath = filepath.Join(appPath, "data")
	logPath = filepath.Join(appPath, "logs")
	return
}
