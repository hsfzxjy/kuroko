package kmux

import "net"

type Listener interface {
	Close() error
	Accept() (RawTransport, error)
}
type NewListenerFunc func() (Listener, error)

type listenerLANv4 struct {
	*net.TCPListener
}

func newLANv4Listener() (Listener, error) {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 5432, IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return nil, err
	}
	return &listenerLANv4{l}, err
}

func (l *listenerLANv4) Accept() (RawTransport, error) {
	conn, err := l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	return NewtcpRawTransport(conn, TT_LANv4), nil
}
