package filelog

import (
	"bufio"
	"context"
	"fmt"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/metrics"
	"github.com/lucky-abc/cleat/record"
	"golang.org/x/text/encoding"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
)

const recordpointDirLogTemplate = "dirlog-%s"

type DirReader struct {
	dirPath       string
	ck            *record.RecordPoint
	queue         chan string
	cancelContext context.Context
	cancelFun     func()
	decoder       *encoding.Decoder
	readFlag      int32 //0:读取未执行，1：正在读取
	waitGroup     sync.WaitGroup
	readMeter     *metrics.Meter
	fileNumMetric *metrics.Counter
}

func CreateDirReader(path string, charset string, ck *record.RecordPoint, queue chan string, metricRegistry *metrics.MetricRegistry) *DirReader {
	r := &DirReader{
		dirPath: path,
		ck:      ck,
		queue:   queue,
	}
	context, cancelf := context.WithCancel(context.Background())
	r.cancelContext = context
	r.cancelFun = cancelf
	decoder := NewMessageDecoder(charset)
	r.decoder = decoder

	readMeter := metrics.NewMeter("directoryread-rate")
	fileNumMetric := metrics.NewCounter("directory-filenum")
	metricRegistry.RegisterMetric(readMeter)
	metricRegistry.RegisterMetric(fileNumMetric)
	r.readMeter = readMeter
	r.fileNumMetric = fileNumMetric
	return r
}

func (dr *DirReader) Read() {
	atomic.StoreInt32(&dr.readFlag, 1)
	dr.waitGroup.Add(1)
	defer func() {
		atomic.StoreInt32(&dr.readFlag, 0)
		dr.waitGroup.Done()
	}()
	dir := dr.dirPath
	fss, err := ioutil.ReadDir(dir)
	if err != nil {
		logger.Loggers().Errorf("get sub file or directory error：%s,%v", dir, err)
		return
	}
	var ffi []os.FileInfo = make([]os.FileInfo, 0)
	//排除掉目录
	for _, f := range fss {
		if f.IsDir() {
			continue
		}
		ffi = append(ffi, f)
	}
	sort.Slice(ffi, func(i, j int) bool {
		return ffi[i].Name() < ffi[j].Name()
	})
	//for _, info := range ffi {
	//	logger.Loggers().Debug(info.Name())
	//
	//}
	var lastFileIndex int
	for i, info := range ffi {
		offset, err := dr.ck.GetCheckpoint(fmt.Sprintf(recordpointDirLogTemplate, filepath.Join(dr.dirPath, info.Name())))
		if err != nil {
			logger.Loggers().Errorf("get file recordpoint error1：%s,%v", info.Name(), err)
			return
		}
		if offset > 0 {
			logger.Loggers().Error(info.Name(), offset)
			lastFileIndex = i
			break
		}
	}
	ffi = ffi[lastFileIndex:]
	for i, info := range ffi {
		fileAbsPath := filepath.Join(dr.dirPath, info.Name())
		offset, err := dr.ck.GetCheckpoint(fmt.Sprintf(recordpointDirLogTemplate, fileAbsPath))
		if err != nil {
			logger.Loggers().Errorf("get file recordpoint error2：%s,%v", fileAbsPath, err)
			return
		}
		file, err := os.Open(fileAbsPath)
		if err != nil {
			logger.Loggers().Errorf("open file error: %s,%v", fileAbsPath, err)
			return
		}
		if offset > 0 {
			file.Seek(int64(offset), 0)
		}
		reader := bufio.NewReader(file)
		dr.fileNumMetric.Incr(1)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					logger.Loggers().Infof("File read complete：%s", fileAbsPath)
					file.Close()
					dr.fileNumMetric.Decr(1)
					if i < len(ffi)-1 {
						dr.ck.DelCheckpoint(fmt.Sprintf(recordpointDirLogTemplate, fileAbsPath))
					}
					break
				}
				logger.Loggers().Errorf("file read error：%s,%v", fileAbsPath, err)
				return
			}
			offset += uint64(len(line))
			offset += uint64(len(line))
			toline, err := dr.decoder.Bytes(line[:len(line)-1])
			if err != nil {
				logger.Loggers().Warnf("character encoding conversion error：%v", err)
				continue
			}
			msg := string(toline)
			//logger.Loggers().Debug(msg)
		Lbl:
			for {
				select {
				case dr.queue <- msg:
					dr.readMeter.Update(1)
					dr.ck.SetCheckpoint(fmt.Sprintf(recordpointDirLogTemplate, fileAbsPath), offset)
					break Lbl
				case <-dr.cancelContext.Done():
					logger.Loggers().Debugf("end of file read: %s", fileAbsPath)
					return
				}
			}
		}
	}
}

func (dr *DirReader) Reading() bool {
	return atomic.LoadInt32(&dr.readFlag) == 1
}

func (dr *DirReader) Close() {
	logger.Loggers().Debug("close directory reader")
	dr.cancelFun()
	dr.waitGroup.Wait()
}
