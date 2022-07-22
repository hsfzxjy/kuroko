package dffi

/*
#cgo LDFLAGS: -landroid -llog -L../../build/ -lkcorec
#include "../../c/for_go.h"
*/
import "C"
import (
	"fmt"
	"kcore_android/magic"
	"unsafe"
)

func asDartValue(v any) (ret C.DartValue, ori any) {
	if v == nil {
		ret.kind = C.DartValue_kNull
		return
	}
	ori = v
	switch vv := v.(type) {
	case bool:
		ret.kind = C.DartValue_kBool
		var b int32 = 0
		if vv {
			b = 1
		}
		*(*int32)(unsafe.Pointer(&ret.value)) = b
	case int32:
		ret.kind = C.DartValue_kInt32
		*(*int32)(unsafe.Pointer(&ret.value)) = int32(vv)
		// fallthrough
	case uint32:
		ret.kind = C.DartValue_kInt32
		*(*int32)(unsafe.Pointer(&ret.value)) = int32(vv)
	case int:
		ret.kind = C.DartValue_kInt64
		*(*int64)(unsafe.Pointer(&ret.value)) = (int64)(vv)
	case int64:
		ret.kind = C.DartValue_kInt64
		*(*int64)(unsafe.Pointer(&ret.value)) = (int64)(vv)
	case uint64:
		ret.kind = C.DartValue_kInt64
		*(*int64)(unsafe.Pointer(&ret.value)) = (int64)(vv)
	case uintptr:
		ret.kind = C.DartValue_kInt64
		*(*int64)(unsafe.Pointer(&ret.value)) = (int64)(vv)
	case []byte:
		ret.kind = C.DartValue_kUint8Array
		var s = (*C.DartValue_vUint8Array)(unsafe.Pointer(&ret))
		data, l := magic.UnpackSlice(vv)
		s.length = C.intptr_t(l)
		s.values = (*C.uint8_t)(unsafe.Pointer(data))
	case string:
		ret.kind = C.DartValue_kString
		if vv[len(vv)-1] != '\x00' {
			vv = vv + "\x00"
			ori = vv
		}
		data, _ := magic.UnpackString(vv)
		*(*C.intptr_t)(unsafe.Pointer(&ret.value)) = C.intptr_t(uintptr(data))
	default:
		panic(fmt.Sprintf("cannot convert %v(%t) to DartValue", vv, vv))
	}
	return
}
