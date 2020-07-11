package filelog

import (
	"github.com/lucky-abc/cleat/config"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/metrics"
	"github.com/lucky-abc/cleat/record"
	"os"
	"time"
)

type FileLogSource struct {
	logChan        chan string
	ck             *record.RecordPoint
	fileReaders    []FileReader
	timeTicker     *time.Ticker
	metricRegistry *metrics.MetricRegistry
}

func NewFileLogSource(c chan string, ck *record.RecordPoint, metricRegistry *metrics.MetricRegistry) *FileLogSource {
	s := &FileLogSource{
		logChan:        c,
		ck:             ck,
		fileReaders:    make([]FileReader, 0),
		metricRegistry: metricRegistry,
	}
	return s
}

func (s *FileLogSource) Start() {
	paths, ok := config.Config().Get("files.paths").([]interface{})
	if !ok {
		logger.Loggers().Error("no file path in config file")
		return
	}
	for _, pathInfo := range paths {
		pathMap, ok := pathInfo.(map[interface{}]interface{})
		if !ok {
			logger.Loggers().Error("no file path in config file2")
			return
		}
		path := pathMap["path"].(string)
		charset := pathMap["charset"].(string)
		finfo, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Loggers().Errorf("the file is not exist: %v", finfo.Name())
				continue
			}
			logger.Loggers().Errorf("get file info error: %v", finfo.Name())
			continue
		}
		var fileReader FileReader
		if finfo.IsDir() {
			fileReader = CreateDirReader(path, charset, s.ck, s.logChan, s.metricRegistry)
		} else {
			fileReader = CreateFileLogReader(path, charset, s.ck, s.logChan, s.metricRegistry)
		}
		s.fileReaders = append(s.fileReaders, fileReader)
	}
	s.timeTicker = time.NewTicker(20 * time.Second)
	go func() {
		for t := range s.timeTicker.C {
			logger.Loggers().Debugf("file reader exec duration: %v", t.Format("2006-01-02 15:04:05.000"))
			for _, reader := range s.fileReaders {
				if !reader.Reading() {
					go reader.Read()
				}
			}
		}
	}()

}

func (s *FileLogSource) Process() {
}

func (s *FileLogSource) Stop() {
	s.timeTicker.Stop()
	for _, reader := range s.fileReaders {
		reader.Close()
	}
	logger.Loggers().Debug("closed file source")
}
