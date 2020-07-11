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
	logFileReporter := metrics.NewLogFileReporter(logger.Logger(), metricRegistry, 10)
	bootTime := time.Now()
	infoSheetMetric := metrics.NewInfoSheet("system_overview", func(infoValues *sync.Map) {
		infoValues.Store("bootTime", bootTime.Format("2006-01-02 15:04:05.999"))
		infoValues.Store("runningTime", time.Since(bootTime).String())
	})
	infoSheetMetric.AddInfo("OS", runtime.GOOS)
	metricRegistry.RegisterMetric(infoSheetMetric)
	metricRegistry.RegisterReporter(logFileReporter)
	logFileReporter.Start()
	return metricRegistry
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
