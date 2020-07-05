package wineventapi

import (
	"syscall"
	"unsafe"
	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer
type EvtHandle uintptr

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modwevtapi = windows.NewLazySystemDLL("wevtapi.dll")

	procEvtOpenLog                      = modwevtapi.NewProc("EvtOpenLog")
	procEvtClearLog                     = modwevtapi.NewProc("EvtClearLog")
	procEvtQuery                        = modwevtapi.NewProc("EvtQuery")
	procEvtSubscribe                    = modwevtapi.NewProc("EvtSubscribe")
	procEvtCreateBookmark               = modwevtapi.NewProc("EvtCreateBookmark")
	procEvtUpdateBookmark               = modwevtapi.NewProc("EvtUpdateBookmark")
	procEvtCreateRenderContext          = modwevtapi.NewProc("EvtCreateRenderContext")
	procEvtRender                       = modwevtapi.NewProc("EvtRender")
	procEvtClose                        = modwevtapi.NewProc("EvtClose")
	procEvtSeek                         = modwevtapi.NewProc("EvtSeek")
	procEvtNext                         = modwevtapi.NewProc("EvtNext")
	procEvtOpenChannelEnum              = modwevtapi.NewProc("EvtOpenChannelEnum")
	procEvtNextChannelPath              = modwevtapi.NewProc("EvtNextChannelPath")
	procEvtFormatMessage                = modwevtapi.NewProc("EvtFormatMessage")
	procEvtOpenPublisherMetadata        = modwevtapi.NewProc("EvtOpenPublisherMetadata")
	procEvtGetPublisherMetadataProperty = modwevtapi.NewProc("EvtGetPublisherMetadataProperty")
	procEvtGetEventMetadataProperty     = modwevtapi.NewProc("EvtGetEventMetadataProperty")
	procEvtOpenEventMetadataEnum        = modwevtapi.NewProc("EvtOpenEventMetadataEnum")
	procEvtNextEventMetadata            = modwevtapi.NewProc("EvtNextEventMetadata")
	procEvtGetObjectArrayProperty       = modwevtapi.NewProc("EvtGetObjectArrayProperty")
	procEvtGetObjectArraySize           = modwevtapi.NewProc("EvtGetObjectArraySize")
)

func IsAvailable() (bool, error) {
	err := modwevtapi.Load()
	if err != nil {
		return false, err
	}
	return true, nil
}

func EvtOpenLog(session EvtHandle, path *uint16, flags uint32) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall(procEvtOpenLog.Addr(), 3, uintptr(session), uintptr(unsafe.Pointer(path)), uintptr(flags))
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtClearLog(session EvtHandle, channelPath *uint16, targetFilePath *uint16, flags uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtClearLog.Addr(), 4, uintptr(session), uintptr(unsafe.Pointer(channelPath)), uintptr(unsafe.Pointer(targetFilePath)), uintptr(flags), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtQuery(session EvtHandle, path *uint16, query *uint16, flags uint32) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall6(procEvtQuery.Addr(), 4, uintptr(session), uintptr(unsafe.Pointer(path)), uintptr(unsafe.Pointer(query)), uintptr(flags), 0, 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtSubscribe(session EvtHandle, signalEvent uintptr, channelPath *uint16, query *uint16, bookmark uintptr, context uintptr, callback syscall.Handle, flags uintptr) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall9(procEvtSubscribe.Addr(), 8, uintptr(session), uintptr(signalEvent), uintptr(unsafe.Pointer(channelPath)), uintptr(unsafe.Pointer(query)), uintptr(bookmark), uintptr(context), uintptr(callback), uintptr(flags), 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtCreateBookmark(bookmarkXML *uint16) (handle uintptr, err error) {
	r0, _, e1 := syscall.Syscall(procEvtCreateBookmark.Addr(), 1, uintptr(unsafe.Pointer(bookmarkXML)), 0, 0)
	//handle = EvtHandle(r0)
	handle = r0
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtUpdateBookmark(bookmark EvtHandle, event EvtHandle) (err error) {
	r1, _, e1 := syscall.Syscall(procEvtUpdateBookmark.Addr(), 2, uintptr(bookmark), uintptr(event), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtCreateRenderContext(ValuePathsCount uint32, valuePaths uintptr, flags uintptr) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall(procEvtCreateRenderContext.Addr(), 3, uintptr(ValuePathsCount), uintptr(valuePaths), flags)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtRender(context EvtHandle, fragment uintptr, flags uintptr, bufferSize uint32, buffer *byte, bufferUsed *uint32, propertyCount *uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procEvtRender.Addr(), 7, uintptr(context), uintptr(fragment), flags, uintptr(bufferSize), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(bufferUsed)), uintptr(unsafe.Pointer(propertyCount)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtClose(object uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procEvtClose.Addr(), 1, uintptr(object), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtSeek(resultSet EvtHandle, position int64, bookmark EvtHandle, timeout uint32, flags uint32) (success bool, err error) {
	r0, _, e1 := syscall.Syscall6(procEvtSeek.Addr(), 5, uintptr(resultSet), uintptr(position), uintptr(bookmark), uintptr(timeout), uintptr(flags), 0)
	success = r0 != 0
	if !success {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtNext(resultSet EvtHandle, eventArraySize uint32, eventArray *uintptr, timeout uint32, flags uint32, numReturned *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtNext.Addr(), 6, uintptr(resultSet), uintptr(eventArraySize), uintptr(unsafe.Pointer(eventArray)), uintptr(timeout), uintptr(flags), uintptr(unsafe.Pointer(numReturned)))
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtOpenChannelEnum(session EvtHandle, flags uint32) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall(procEvtOpenChannelEnum.Addr(), 2, uintptr(session), uintptr(flags), 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtNextChannelPath(channelEnum EvtHandle, channelPathBufferSize uint32, channelPathBuffer *uint16, channelPathBufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtNextChannelPath.Addr(), 4, uintptr(channelEnum), uintptr(channelPathBufferSize), uintptr(unsafe.Pointer(channelPathBuffer)), uintptr(unsafe.Pointer(channelPathBufferUsed)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtFormatMessage(publisherMetadata EvtHandle, event EvtHandle, messageID uint32, valueCount uint32, values uintptr, flags uintptr, bufferSize uint32, buffer *byte, bufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procEvtFormatMessage.Addr(), 9, uintptr(publisherMetadata), uintptr(event), uintptr(messageID), uintptr(valueCount), uintptr(values), flags, uintptr(bufferSize), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtOpenPublisherMetadata(session EvtHandle, publisherIdentity *uint16, logFilePath *uint16, locale uint32, flags uint32) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall6(procEvtOpenPublisherMetadata.Addr(), 5, uintptr(session), uintptr(unsafe.Pointer(publisherIdentity)), uintptr(unsafe.Pointer(logFilePath)), uintptr(locale), uintptr(flags), 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtGetPublisherMetadataProperty(publisherMetadata EvtHandle, propertyID uintptr, flags uint32, bufferSize uint32, variant *uint32, bufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtGetPublisherMetadataProperty.Addr(), 6, uintptr(publisherMetadata), propertyID, uintptr(flags), uintptr(bufferSize), uintptr(unsafe.Pointer(variant)), uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtGetEventMetadataProperty(eventMetadata EvtHandle, propertyID uintptr, flags uint32, bufferSize uint32, variant *uint32, bufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procEvtGetEventMetadataProperty.Addr(), 6, uintptr(eventMetadata), propertyID, uintptr(flags), uintptr(bufferSize), uintptr(unsafe.Pointer(variant)), uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtOpenEventMetadataEnum(publisherMetadata EvtHandle, flags uint32) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall(procEvtOpenEventMetadataEnum.Addr(), 2, uintptr(publisherMetadata), uintptr(flags), 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtNextEventMetadata(enumerator EvtHandle, flags uint32) (handle EvtHandle, err error) {
	r0, _, e1 := syscall.Syscall(procEvtNextEventMetadata.Addr(), 2, uintptr(enumerator), uintptr(flags), 0)
	handle = EvtHandle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtGetObjectArrayProperty(objectArray uintptr, propertyID uintptr, arrayIndex uint32, flags uint32, bufferSize uint32, evtVariant *uint32, bufferUsed *uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procEvtGetObjectArrayProperty.Addr(), 7, objectArray, propertyID, uintptr(arrayIndex), uintptr(flags), uintptr(bufferSize), uintptr(unsafe.Pointer(evtVariant)), uintptr(unsafe.Pointer(bufferUsed)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EvtGetObjectArraySize(objectArray uintptr, arraySize *uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procEvtGetObjectArraySize.Addr(), 2, uintptr(objectArray), uintptr(unsafe.Pointer(arraySize)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
