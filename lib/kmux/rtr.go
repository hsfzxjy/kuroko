package kmux

import (
	"io"
	"net"
	"sync"
)

type RawTransport interface {
	io.ReadWriteCloser
	TransportKey() *TransportKey
	ClosedCh() <-chan struct{}
}

type tcpRawTransport struct {
	*net.TCPConn
	typ      TransportType
	key      *TransportKey
	closedCh chan struct{}
	dieOnce  sync.Once
}

func NewtcpRawTransport(conn *net.TCPConn, typ TransportType) *tcpRawTransport {
	rtr := new(tcpRawTransport)

	rtr.TCPConn = conn
	rtr.typ = typ
	rtr.key = keyFromTCPAddr(conn.RemoteAddr().(*net.TCPAddr), typ)
	rtr.closedCh = make(chan struct{})
	return rtr
}

func (t *tcpRawTransport) Close() error {
	t.dieOnce.Do(func() { close(t.closedCh) })
	return t.TCPConn.Close()

}

func (t *tcpRawTransport) TransportKey() *TransportKey { return t.key }
func (t *tcpRawTransport) ClosedCh() <-chan struct{}   { return t.closedCh }
