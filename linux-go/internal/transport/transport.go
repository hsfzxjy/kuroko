package transport

import (
	"kmux"

	"kuroko-linux/internal"
	"kuroko-linux/internal/transport/bluetooth"
)

func BridgeInfoToTransportKey(bi internal.BridgeInfo) *kmux.TransportKey {
	typ := bi.GetType()
	tk := new(kmux.TransportKey)
	tk[0] = byte(typ)
	switch typ {
	case kmux.TT_BLUETOOTH:
		addr := bluetooth.Str2Ba(bi.GetAddr())
		copy(tk[1:7], addr[:])
	}
	return tk
}

type info struct {
	typ  kmux.TransportType
	addr string
}

func (i *info) GetType() kmux.TransportType { return i.typ }
func (i *info) GetAddr() string             { return i.addr }

func TransportKeyToBridgeInfo(tk *kmux.TransportKey) internal.BridgeInfo {
	i := new(info)
	i.typ = tk.Type()
	switch i.typ {
	case kmux.TT_BLUETOOTH:
		i.addr = bluetooth.Ba2Str(tk.BluetoothMAC())
	}
	return i
}
