package magic

import "unsafe"

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

type stringHeader struct {
	data unsafe.Pointer
	len  int
}

func UnpackSlice[T any](s []T) (unsafe.Pointer, int) {
	var h = (*sliceHeader)(unsafe.Pointer(&s))
	return h.data, h.len
}

func UnpackString(s string) (unsafe.Pointer, int) {
	var h = (*stringHeader)(unsafe.Pointer(&s))
	return h.data, h.len
}