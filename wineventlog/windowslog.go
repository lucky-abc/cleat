package wineventlog

import (
	"bytes"
	"fmt"
	"github.com/beevik/etree"
	"github.com/lucky-abc/cleat/logger"
	"github.com/lucky-abc/cleat/metrics"
	"github.com/lucky-abc/cleat/record"
	"github.com/lucky-abc/cleat/wineventlog/wineventapi"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/sys/windows"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	bookmarkTemplate   = `<BookmarkList><Bookmark Channel="%s" RecordId="%d" IsCurrent="True"/></BookmarkList>`
	checkpointTemplate = `window-event-[%s]`
)
const (
	ERROR_INSUFFICIENT_BUFFER syscall.Errno = 122
	ERROR_NO_MORE_ITEMS       syscall.Errno = 259
	RPC_S_INVALID_BOUND       syscall.Errno = 1734
	ERROR_INVALID_OPERATION   syscall.Errno = 4317
)

type WindowsLog struct {
	LogName       string
	RecordNumber  uint64
	eventHandle   wineventapi.EvtHandle
	outputBuf     *bytes.Buffer
	renderBuf     []byte
	queue         chan string
	cancelContext context.Context
	cancelFun     func()
	waitGroup     sync.WaitGroup
	ck            *record.RecordPoint
	runFlag       int32
	metricMeter   *metrics.Meter
	recorCounter  *metrics.Counter
}

func NewWindowsLog(logName string, queue chan string, ck *record.RecordPoint, metricRegistry *metrics.MetricRegistry) *WindowsLog {
	l := &WindowsLog{
		LogName:   logName,
		outputBuf: bytes.NewBuffer(make([]byte, 1<<14)),
		renderBuf: make([]byte, 1<<14),
		queue:     queue,
		ck:        ck,
		runFlag:   0,
	}
	metricMeter := metrics.NewMeter("windowevent[" + logName + "]-read-rate")
	metricRegistry.RegisterMetric(metricMeter)
	context, cancelf := context.WithCancel(context.Background())
	l.cancelContext = context
	l.cancelFun = cancelf
	l.metricMeter = metricMeter
	l.recorCounter = metricRegistry.GetCounter("windowevent-record-total")
	return l
}

func (log *WindowsLog) Open() {
	atomic.StoreInt32(&log.runFlag, 1)
	handle, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		logger.Loggers().Errorf("create windows event fail:%v", err)
		return
	}
	defer windows.CloseHandle(handle)

	q, err := syscall.UTF16PtrFromString("*")
	if err != nil {
		logger.Loggers().Errorf("create windows event query fail:%v", err)
		return
	}
	//var cp *uint16
	cp, err := syscall.UTF16PtrFromString(log.LogName)
	if err != nil {
		logger.Loggers().Errorf("create windows event channel fail:%v", err)
		return
	}
	ckVal, err := log.ck.GetCheckpoint(fmt.Sprintf(checkpointTemplate, log.LogName))
	if err != nil {
		logger.Loggers().Errorf("window event get checkpoint error:%v", err)
		return
	}
	logger.Loggers().Debugf("window event %s checkpoint: %d", log.LogName, ckVal)
	bookmark, err := CreateBookmarkFromRecordID(log.LogName, ckVal)
	if err != nil {
		logger.Loggers().Errorf("create windows event bookmark fail:%v", err)
	}
	var flags uint32
	if bookmark > 0 {
		flags = 3
	} else {
		flags = 2
	}
	eventHandle, err := wineventapi.EvtSubscribe(0, uintptr(handle), cp, q, bookmark, 0, 0, uintptr(flags))
	if err != nil {
		logger.Loggers().Errorf("windows event subscribe fail:%v", err)
	}
	log.eventHandle = eventHandle
}

func (log *WindowsLog) Read() {
	log.waitGroup.Add(1)
	defer func() {
		atomic.StoreInt32(&log.runFlag, 0)
		log.waitGroup.Done()
	}()
	var maxHandles = 100
	eventHandles := make([]uintptr, maxHandles)
	var numRead uint32
	for {
		err := wineventapi.EvtNext(log.eventHandle, uint32(len(eventHandles)), &eventHandles[0], 0, 0, &numRead)
		if atomic.LoadInt32(&log.runFlag) == 0 {
			logger.Loggers().Info("window event read over")
			return
		}
		if err != nil {
			if err == ERROR_INVALID_OPERATION && numRead == 0 || err == ERROR_NO_MORE_ITEMS {
				logger.Loggers().Warn("windows event has no more record, sleep a little")
				time.Sleep(time.Second * 3)
				continue
			}
			logger.Loggers().Errorf("windows event read next fail:%v", err)
			return
		}
		eventHandles = eventHandles[:numRead]
		err = log.eventlogRender(eventHandles)
		if err != nil {
			return
		}
	}
}

func (log *WindowsLog) eventlogRender(eventHandles []uintptr) error {
	if len(eventHandles) <= 0 {
		logger.Loggers().Errorf("windows event len:", len(eventHandles))
		return errors.New("windows event is empty")
	}
	defer func() {
		for _, h := range eventHandles {
			wineventapi.EvtClose(h)
		}
	}()
	for _, handle := range eventHandles {
		log.outputBuf.Reset()
		var bufferUsed, propertyCount uint32
		err := wineventapi.EvtRender(0, uintptr(handle), uintptr(uint32(1)), uint32(len(log.renderBuf)), &log.renderBuf[0], &bufferUsed, &propertyCount)
		if err != nil {
			if err == ERROR_INSUFFICIENT_BUFFER {
				logger.Loggers().Warnf("windows event insufficient buffer")
				log.renderBuf = make([]byte, bufferUsed)
				log.outputBuf.Reset()
				wineventapi.EvtRender(0, uintptr(handle), uintptr(uint32(1)), uint32(len(log.renderBuf)), &log.renderBuf[0], &bufferUsed, &propertyCount)
			} else {
				logger.Loggers().Errorf("windows event render error:%v", err)
				return err
			}
		}
		UTF16ToUTF8Bytes(log.renderBuf[:bufferUsed], log.outputBuf)
		//logger.Loggers().Debugf("windows event xml:%v", string(log.outputBuf.Bytes()))
		xmlEvent, eventRecordID, err := log.rebuildXml(log.outputBuf.Bytes())
		if err != nil {
			logger.Loggers().Errorf("windows event rebuild error:%v", err)
			return err
		}
	lfor:
		for {
			select {
			case <-log.cancelContext.Done():
				return errors.New("window event close")
			case log.queue <- xmlEvent:
				log.RecordNumber = eventRecordID
				log.metricMeter.Update(1)
				log.recorCounter.Incr(1)
				log.ck.SetCheckpoint(fmt.Sprintf(checkpointTemplate, log.LogName), eventRecordID)
				break lfor
			}
		}
	}
	return nil
}

func (log *WindowsLog) rebuildXml(xmlbytes []byte) (string, uint64, error) {
	doc := etree.NewDocument()
	err := doc.ReadFromBytes(xmlbytes)
	if err != nil {
		logger.Loggers().Errorf("window event rebuild data error:%v", err)
		return "", 0, err
	}

	securityEle := doc.FindElement("//System/Security")
	if securityEle != nil && securityEle.SelectAttr("UserID") != nil && securityEle.SelectAttr("UserID").Value != "" {
		userID := securityEle.SelectAttr("UserID").Value
		userAccount, err := getAccount(userID)
		if err != nil {
			logger.Loggers().Warnf("window event get account fail:%v", err)
		}
		securityEle.CreateAttr("UserAccount", userAccount)
	}
	result, err := doc.WriteToString()
	if err != nil {
		logger.Loggers().Errorf("window event rebuild data to string error:%v", err)
		return "", 0, err
	}
	eventIDStr := doc.FindElement("//System/EventRecordID").Text()
	eventID, err := strconv.ParseUint(eventIDStr, 10, 64)
	if err != nil {
		logger.Loggers().Errorf("window event parse eventID error:%v", err)
		return "", 0, err
	}
	//logger.Loggers().Debug("***********:", result)
	return result, eventID, nil

}

//func (log *WindowsLog) buildEvent(outputBuf *bytes.Buffer) (Event, error) {
//	var e Event
//	decoder := xml.NewDecoder(bytes.NewReader(outputBuf.Bytes()))
//	err := decoder.Decode(&e)
//	if err != nil {
//		logger.Loggers().Errorf("windows event decode from xml error:%v", err)
//		return e, err
//	}
//	err = PopulateAccount(&e.User)
//	if err != nil {
//		logger.Loggers().Warnf("windows event populate account error:%v", err)
//	}
//	logger.Loggers().Debugf("windows event struct:%v", e)
//	return e, nil
//}
func (log *WindowsLog) Close() {
	log.cancelFun()
	atomic.StoreInt32(&log.runFlag, 0)
	log.waitGroup.Wait()
	wineventapi.EvtClose(uintptr(log.eventHandle))
	logger.Loggers().Infof("windows event handle closed: %s", log.LogName)
}

func CreateBookmarkFromRecordID(channel string, recordID uint64) (uintptr, error) {
	xml := fmt.Sprintf(bookmarkTemplate, channel, recordID)
	p, err := syscall.UTF16PtrFromString(xml)
	if err != nil {
		return 0, err
	}

	h, err := wineventapi.EvtCreateBookmark(p)
	if err != nil {
		return 0, err
	}

	return h, nil
}

func getAccount(userid string) (string, error) {
	if userid == "" {
		return "", nil
	}

	s, err := windows.StringToSid(userid)
	if err != nil {
		return "", err
	}

	account, _, _, err := s.LookupAccount("")
	if err != nil {
		return "", err
	}
	return account, nil
}
