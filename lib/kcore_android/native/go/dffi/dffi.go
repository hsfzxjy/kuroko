package dffi

/*
#cgo LDFLAGS: -L../../build/ -lkcorec
#include "../../c/for_go.h"
*/
import "C"
import (
	"runtime"
	"sync/atomic"
)

var dartVersion uint64 = uint64(1) << 32

func IncDartVersion() {
	atomic.AddUint64(&dartVersion, uint64(1)<<32)
}

func WrapDartCallback[T ~uint32](cb T) DartCallback {
	ver := atomic.LoadUint64(&dartVersion)
	return DartCallback(ver + uint64(cb))
}

type DartCallback uint64

func (cb DartCallback) ResolveFuture(res any) bool {
	switch v := res.(type) {
	case []any:
		var args = make([]any, len(v)+1)
		args[0] = CALL_ARRAY | CALL_WITH_CODE | FUT_RESOLVED
		copy(args[1:], v[:])
		return QueueDartCallback(cb, args...)
	default:
		return QueueDartCallback(cb, CALL_SPREAD|CALL_WITH_CODE|FUT_RESOLVED, v)
	}
}

func (cb DartCallback) RejectFuture(err error) bool {
	return QueueDartCallback(cb, CALL_SPREAD|CALL_WITH_CODE|FUT_REJECTED, err.Error())
}

func (cb DartCallback) CompleteFuture(err error, res any) bool {
	if err != nil {
		return cb.RejectFuture(err)
	} else {
		return cb.ResolveFuture(res)
	}
}

func QueueDartCallback(cb DartCallback, args ...any) bool {
	var ccb C.DartCallback
	{
		cbver := uint64(cb) & (((uint64(1) << 32) - 1) << 32)
		if cbver != atomic.LoadUint64(&dartVersion) {
			return false
		}
		ccb = C.DartCallback(uint64(cb) - cbver)
	}

	var cargs [16]C.DartValue
	var n = len(args)

	for i := 0; i < n; i++ {
		cargs[i], args[i] = asDartValue(args[i])
	}

	C.QueueCallback(ccb, C.int(n), &cargs[0])
	runtime.KeepAlive(cargs)
	runtime.KeepAlive(args)
	return true
}
