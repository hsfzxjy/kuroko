package kmux

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
)

const TK_LEN = 24

type TransportKey [TK_LEN]byte

func (tk *TransportKey) Type() TransportType {
	return TransportType(tk[0])
}

func (tk *TransportKey) LANv4IpBytes() (ret [4]byte) {
	copy(ret[:], tk[1:5])
	return
}

func (tk *TransportKey) LANv4Ip() net.IP {
	return net.IPv4(tk[2], tk[3], tk[4], tk[5])
}

func (tk *TransportKey) LANv4Port() uint16 {
	return binary.BigEndian.Uint16(tk[5:7])
}

func (tk *TransportKey) LANv4TCPAddr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   tk.LANv4Ip(),
		Port: int(tk.LANv4Port()),
	}
}

func (tk *TransportKey) BluetoothMAC() (ret [6]byte) {
	copy(ret[:], tk[1:7])
	return
}

func (tk *TransportKey) Dial(ctx context.Context) (rtr RawTransport, err error) {
	var typ = tk.Type()
	switch typ {
	case TT_BLUETOOTH:
		addr := tk.BluetoothMAC()
		return bluetoothDialer(ctx, addr)
	case TT_LANv4:
		conn, err := net.DialTCP("tcp", nil, tk.LANv4TCPAddr())
		if err != nil {
			return nil, err
		}
		return NewtcpRawTransport(conn, typ), nil
	default:
		return nil, ErrUnknownAddrType
	}

}

var ErrUnknownAddrType = errors.New("unknown addr type")

func keyFromTCPAddr(addr *net.TCPAddr, typ TransportType) *TransportKey {
	var key = new([TK_LEN]byte)
	key[0] = byte(typ)
	switch typ {
	case TT_LANv4:
		copy(key[1:5], addr.IP.To4())
		binary.BigEndian.PutUint16(key[5:7], uint16(addr.Port))
	default:
		panic(ErrUnknownAddrType)
	}
	return (*TransportKey)(key)
}
