package main

import (
	"github.com/lucky-abc/cleat/config"
	"github.com/lucky-abc/cleat/filelog"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/record"
	"github.com/lucky-abc/cleat/wineventlog"
	"os"
	"os/signal"
	"path/filepath"
)

func main() {
	_, configPath, dataPath, logPath := getStartupPath()
	config.InitSystemConfig("config", configPath)
	logger.NewLogger(logPath, config.Config())

	ck, err := record.NewCheckpoint(filepath.Join(dataPath, "recordpoint"))
	if err != nil {
		logger.Loggers().Error("new checkpoint error:", err)
		return
	}

	winlogTunnel := wineventlog.NewWindowslogTunnel(ck)
	if winlogTunnel != nil {
		winlogTunnel.Start()
		winlogTunnel.Transfer()
	}

	fileTunnel := filelog.NewFilelogTunnel(ck)
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
