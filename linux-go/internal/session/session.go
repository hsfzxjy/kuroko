package session

import (
	"bufio"
	"context"
	"kmux"
	"kuroko-linux/internal"
	"kuroko-linux/internal/transport"
	"kuroko-linux/models"
)

type _sessionId struct {
	id           kmux.SessionId
	panicOnError bool
}

func (s *_sessionId) Read(p []byte) (int, error) {
	n, err := s.id.Read(p)
	if err != nil {
		panic(err)
	}
	return n, err
}

func (s *_sessionId) Write(p []byte) (int, error) {
	n, err := s.id.Write(p)
	if err != nil {
		panic(err)
	}
	return n, err
}

func (s *_sessionId) Close() error {
	return s.id.Close()
}

type Session struct {
	id *_sessionId

	Rw *bufio.ReadWriter
	*models.Bridge
}

func (s *Session) SetPanicOnError(v bool) {
	s.id.panicOnError = v
}

func newSession(id kmux.SessionId) *Session {
	wrapped := &_sessionId{id, false}
	reader := bufio.NewReader(wrapped)
	writer := bufio.NewWriter(wrapped)
	rw := bufio.NewReadWriter(reader, writer)
	return &Session{id: wrapped, Rw: rw}
}

func WrapSession(id kmux.SessionId) (*Session, error) {
	tk, err := id.TransportKey()
	if err != nil {
		return nil, err
	}
	bi := transport.TransportKeyToBridgeInfo(tk)
	sess := newSession(id)
	sess.Bridge = models.NewBridge(bi)
	return sess, nil
}

func DialSession(bi internal.BridgeInfo) (*Session, error) {
	if tr, ok := bi.(*Session); ok {
		return tr, nil
	}

	tk := transport.BridgeInfoToTransportKey(bi)
	id, _, err := kmux.DialSession(context.TODO(), tk)
	if err != nil {
		return nil, err
	}

	sess := newSession(id)
	sess.Bridge = models.NewBridge(bi)
	return sess, nil
}
