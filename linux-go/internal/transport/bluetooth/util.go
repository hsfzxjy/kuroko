package bluetooth

import (
	"fmt"
	"strconv"
	"strings"
)

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
		addr[5], addr[4], addr[3], addr[2], addr[1], addr[0])
}
