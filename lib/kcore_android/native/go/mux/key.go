package mux

// #include <stdint.h>
import "C"
import (
	"encoding/binary"
	"kmux"
)

func Ctk2Tk[T ~uint64](k1, k2, k3 T) *kmux.TransportKey {
	ret := new(kmux.TransportKey)
	binary.BigEndian.PutUint64(ret[0:8], uint64(k1))
	binary.BigEndian.PutUint64(ret[8:16], uint64(k2))
	binary.BigEndian.PutUint64(ret[16:24], uint64(k3))
	return ret
}

func Tk2Ctk[T ~uint64](tk *kmux.TransportKey) (T, T, T) {
	k1 := binary.BigEndian.Uint64(tk[0:8])
	k2 := binary.BigEndian.Uint64(tk[8:16])
	k3 := binary.BigEndian.Uint64(tk[16:24])
	return T(k1), T(k2), T(k3)
}
