package main

/*
#include <stdint.h>
#include "../c/for_go.h"
*/
import "C"
import (
	"context"
	"errors"
	"io"
	"kcore_android/dffi"
	"kcore_android/mux"
	"kmux"
	"sync"
	"sync/atomic"
	"unsafe"
)

//export KCore_SessionRelease
func KCore_SessionRelease(extrap uintptr) {
	extra := (*mux.SessionExtra)(unsafe.Pointer(extrap))
	extra.DecRefCnt()
}

var (
	dialCancelerCounter uint32 = 0
	dialCancelers       sync.Map
)

//export KCore_CancelDial
func KCore_CancelDial(token C.uint32_t) {
	canceler, loaded := dialCancelers.LoadAndDelete(uint32(token))
	if loaded {
		canceler.(context.CancelFunc)()
	}
}

//export KCore_SessionDial
func KCore_SessionDial(x1, x2, x3 C.uint64_t, ccb C.DartCallback) C.uint32_t {
	tk := mux.Ctk2Tk(x1, x2, x3)
	cb := dffi.WrapDartCallback(ccb)

	ctx, canceler := context.WithCancel(context.Background())
	cancelToken := atomic.AddUint32(&dialCancelerCounter, 1)
	dialCancelers.Store(cancelToken, canceler)

	go func() {
		sid, sw, err := kmux.DialSession(ctx, tk)
		dialCancelers.Delete(cancelToken)
		if err != nil {
			cb.RejectFuture(err)
		} else {
			extra := sw.Extra.(*mux.SessionExtra)
			extra.IncRefCnt()
			cb.ResolveFuture([]any{
				(uint32)(sid),
				(uintptr)(unsafe.Pointer(extra)),
				(uintptr)(unsafe.Pointer(extra.ReadBuf)),
				(uintptr)(unsafe.Pointer(extra.WriteBuf)),
			})
		}
	}()

	return C.uint32_t(cancelToken)
}

//export KCore_SessionRead
func KCore_SessionRead(csid C.uint32_t, buf_ptr *C.uint8_t, buf_size C.uint64_t, ccb C.DartCallback) {
	sid := kmux.SessionId(csid)
	cb := dffi.WrapDartCallback(ccb)

	sw, err := sid.Get()
	if err != nil {
		cb.CompleteFuture(err, -1)
		return
	}
	extra := sw.Extra.(*mux.SessionExtra)
	if !extra.SetReading() {
		cb.CompleteFuture(errors.New("concurrent reading"), -1)
		return
	}

	buf := unsafe.Slice((*byte)(buf_ptr), buf_size)

	go func() {
		n, err := io.ReadFull(sw.Session, buf)
		extra.ClearReading()
		cb.CompleteFuture(err, n)
	}()
}

//export KCore_SessionWrite
func KCore_SessionWrite(csid C.uint32_t, buf_ptr *C.uint8_t, buf_size C.uint64_t, ccb C.DartCallback) {
	sid := kmux.SessionId(csid)
	cb := dffi.WrapDartCallback(ccb)

	sw, err := sid.Get()
	if err != nil {
		cb.CompleteFuture(err, -1)
		return
	}
	extra := sw.Extra.(*mux.SessionExtra)
	if !extra.SetWriting() {
		cb.CompleteFuture(errors.New("concurrent writing"), -1)
		return
	}

	buf := unsafe.Slice((*byte)(buf_ptr), buf_size)

	go func() {
		n, err := sw.Session.Write(buf)
		extra.ClearWriting()
		cb.CompleteFuture(err, n)
	}()
}

//export KCore_SessionClose
func KCore_SessionClose(csid C.uint32_t, ccb C.DartCallback) {
	sid := kmux.SessionId(csid)
	cb := dffi.WrapDartCallback(ccb)

	go func() {
		err := sid.Close()
		cb.CompleteFuture(err, nil)
	}()
}

//export KCore_SessionGetTransportKey
func KCore_SessionGetTransportKey(csid C.uint32_t, buf_ptr *C.uint8_t) int32 {
	buf := unsafe.Slice((*byte)(buf_ptr), kmux.TK_LEN)
	sid := kmux.SessionId(csid)
	tk, err := sid.TransportKey()
	if err != nil {
		return dffi.RETCODE_ERROR
	}
	copy(buf, tk[:])
	return dffi.RETCODE_OK
}
