package wineventlog

import (
	"github.com/lucky-abc/cleat/config"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/record"
)

type WinLogSource struct {
	logChan     chan string
	windowsLogs []*WindowsLog
}

func NewWinLogSource(c chan string, ck *record.RecordPoint) *WinLogSource {
	eventNames := config.Config().GetStringSlice("windows.event.eventname")
	logger.Loggers().Infof("window event channel: %v", eventNames)
	if len(eventNames) == 0 {
		logger.Loggers().Errorf("no window event channel")
		return nil
	}
	windowLogs := make([]*WindowsLog, len(eventNames))
	for i, eventname := range eventNames {
		l := NewWindowsLog(eventname, c, ck)
		windowLogs[i] = l
	}
	s := &WinLogSource{
		logChan:     c,
		windowsLogs: windowLogs,
	}
	return s
}

func (s *WinLogSource) Start() {
	for _, log := range s.windowsLogs {
		logger.Loggers().Debugf("open window event channel:%s", log.LogName)
		log.Open()
	}
}

func (s *WinLogSource) Process() {
	for _, log := range s.windowsLogs {
		go log.Read()
	}
}

func (s *WinLogSource) Stop() {
	for _, log := range s.windowsLogs {
		log.Close()
	}
	close(s.logChan)
}
