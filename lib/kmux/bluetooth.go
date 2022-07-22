package kmux

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type BluetoothDialerFunc func(context.Context, [6]byte) (RawTransport, error)

var bluetoothDialer BluetoothDialerFunc

func SetBluetoothDialer(f BluetoothDialerFunc) {
	bluetoothDialer = f
}

func Str2Ba(addr string) [6]byte {
	a := strings.Split(addr, ":")
	var b [6]byte
	for i, tmp := range a {
		u, _ := strconv.ParseUint(tmp, 16, 8)
		b[len(b)-1-i] = byte(u)
	}
	return b
}

func Ba2Str(addr [6]byte) string {
	return fmt.Sprintf("%2.2X:%2.2X:%2.2X:%2.2X:%2.2X:%2.2X",
		addr[0], addr[1], addr[2], addr[3], addr[4], addr[5])
}
