package main

/*
#include <stdint.h>
#include "../c/for_go.h"
*/
import "C"
import (
	"kcore_android/dffi"
	"kcore_android/mux"
	"kmux"
	"sync/atomic"
	"unsafe"
)

//export KCore_AccepterSetStateCallback
func KCore_AccepterSetStateCallback(ctt C.uint8_t, ccb C.DartCallback) {
	cb := dffi.WrapDartCallback(ccb)
	tt := kmux.TransportType(ctt)
	go func() {
		accepter := kmux.AccepterManager.Get(tt)
		if !dffi.QueueDartCallback(cb, dffi.CALL_MULTI, int32(accepter.State())) {
			return
		}
		for state := range accepter.StateCh() {
			if !dffi.QueueDartCallback(cb, dffi.CALL_MULTI, int32(state)) {
				break
			}
		}
	}()
}

//export KCore_AccepterStart
func KCore_AccepterStart(ctt C.uint8_t) {
	tt := kmux.TransportType(ctt)
	go kmux.AccepterManager.Get(tt).Start()
}

//export KCore_AccepterStop
func KCore_AccepterStop(ctt C.uint8_t) {
	tt := kmux.TransportType(ctt)
	go kmux.AccepterManager.Get(tt).Stop()
}

var acceptCallback dffi.DartCallback

func init() {
	go func() {
		cb := dffi.DartCallback(atomic.LoadUint64((*uint64)(&acceptCallback)))
		for sid := range kmux.AccepterManager.SessionCh() {
			sw, err := sid.Get()
			if err != nil {
				continue
			}
			k1, k2, k3 := mux.Tk2Ctk[uint64](sw.TransportKey)
			extra := sw.Extra.(*mux.SessionExtra)
			if !dffi.QueueDartCallback(
				cb, dffi.CALL_ARRAY|dffi.CALL_MULTI,
				uint32(sid),
				k1, k2, k3,
				(uintptr)(unsafe.Pointer(extra)),
				(uintptr)(unsafe.Pointer(extra.ReadBuf)),
				(uintptr)(unsafe.Pointer(extra.WriteBuf)),
			) {
				sid.Close()
				cb = dffi.DartCallback(atomic.LoadUint64((*uint64)(&acceptCallback)))
			}
		}
	}()
}

//export KCore_AccepterAccept
func KCore_AccepterAccept(ccb C.DartCallback) {
	cb := dffi.WrapDartCallback(ccb)
	atomic.StoreUint64((*uint64)(&acceptCallback), uint64(cb))
}
