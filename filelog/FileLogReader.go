package filelog

import (
	"bufio"
	"context"
	"fmt"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/record"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"os"
	"strings"
	"sync/atomic"
)

const recordpointFileLogTemplate = "filelog-%s"

type FileReader interface {
	Read()
	Close()
	Reading() bool
}

func NewMessageDecoder(charset string) (decoder *encoding.Decoder) {
	cset := strings.ToLower(charset)
	switch cset {
	case "gbk":
		decoder = simplifiedchinese.GBK.NewDecoder()
	case "gb1830":
		decoder = simplifiedchinese.GB18030.NewDecoder()
	case "gb2312":
		decoder = simplifiedchinese.HZGB2312.NewDecoder()
	default:
		decoder = encoding.Nop.NewDecoder()
	}
	return
}

type FileLogReader struct {
	filePath      string
	ck            *record.RecordPoint
	queue         chan string
	cancelContext context.Context
	cancelFun     func()
	decoder       *encoding.Decoder
	readFlag      int32 //0:读取未执行，1：正在读取
}

func CreateFileLogReader(path string, charset string, ck *record.RecordPoint, queue chan string) *FileLogReader {
	r := &FileLogReader{
		filePath: path,
		ck:       ck,
		queue:    queue,
	}
	context, cancelf := context.WithCancel(context.Background())
	r.cancelContext = context
	r.cancelFun = cancelf
	decoder := NewMessageDecoder(charset)
	r.decoder = decoder
	return r
}

func (fr *FileLogReader) Read() {
	atomic.StoreInt32(&fr.readFlag, 1)
	defer func() {
		atomic.StoreInt32(&fr.readFlag, 0)
	}()
	file, err := os.Open(fr.filePath)
	if err != nil {
		logger.Loggers().Errorf("open file error: %s,%v", file.Name(), err)
		return
	}
	defer file.Close()
	offset, err := fr.ck.GetCheckpoint(fmt.Sprintf(recordpointFileLogTemplate, file.Name()))
	if err != nil {
		logger.Loggers().Errorf("get file recordpoint error： %s", file.Name())
		return
	}
	if offset > 0 {
		file.Seek(int64(offset), 0)
	}
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				logger.Loggers().Infof("File read complete：%s", file.Name())
				file.Close()
				break
			}
			logger.Loggers().Errorf("file read error：%s,%v", file.Name(), err)
		}
		offset += uint64(len(line))
		toline, err := fr.decoder.Bytes(line[:len(line)-1])
		if err != nil {
			logger.Loggers().Warnf("character encoding conversion error：%v", err)
			continue
		}
		msg := string(toline)
		//logger.Loggers().Debug(msg)
		//logger.Loggers().Debug(offset)
	Lbl:
		for {
			select {
			case fr.queue <- msg:
				fr.ck.SetCheckpoint(fmt.Sprintf(recordpointFileLogTemplate, file.Name()), offset)
				break Lbl
			case <-fr.cancelContext.Done():
				logger.Loggers().Debugf("end of file read: %s", file.Name())
				return
			}
		}
	}
}

func (fr *FileLogReader) Reading() bool {
	return atomic.LoadInt32(&fr.readFlag) == 1
}

func (fr *FileLogReader) Close() {
	logger.Loggers().Debug("close filelog reader")
	fr.cancelFun()
}
