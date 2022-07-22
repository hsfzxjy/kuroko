package kmux

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/hsfzxjy/smux"
)

type (
	Transport = *smux.Session
	Session   = *smux.Stream
	SessionId uint32
)

type connManager struct {
	trs       map[TransportKey]*transportWrapper
	sessions  sync.Map
	trcounter uint32
	counter   uint32
	trl       sync.Mutex
}

type transportWrapper struct {
	Transport
	id     uint32
	dialed chan struct{}
	err    error
}

type sessionWrapper struct {
	Session
	*TransportKey
	Extra any
}

var conns = &connManager{
	trs:      make(map[TransportKey]*transportWrapper),
	sessions: sync.Map{},
}

func (cm *connManager) watchRtr(id uint32, rtr RawTransport) {
	<-rtr.ClosedCh()
	cm.trl.Lock()
	defer cm.trl.Unlock()
	var key = *rtr.TransportKey()
	if trw, exists := cm.trs[key]; exists {
		if trw.id != id {
			return
		}
	} else {
		return
	}
	delete(cm.trs, *rtr.TransportKey())
}

func (cm *connManager) acceptLoop(tr Transport, tk *TransportKey) {
	for {
		sess, err := tr.AcceptStream()
		if err != nil {
			break
		}
		sid, _ := cm.putSession(sess, tk)
		AccepterManager.sessCh <- sid
	}
}

func (cm *connManager) watchSession(sid SessionId, sess Session) {
	<-sess.GetDieCh()
	conns.sessions.Delete(sid)
}

func (cm *connManager) putSession(sess Session, tk *TransportKey) (sid SessionId, sw *sessionWrapper) {
	sid = SessionId(atomic.AddUint32(&conns.counter, 1))

	var extra any
	if SessionNewExtraFunc != nil {
		extra = SessionNewExtraFunc(sess)
	}

	sw = &sessionWrapper{
		Session:      sess,
		TransportKey: tk,
		Extra:        extra,
	}

	conns.sessions.Store(sid, sw)
	go cm.watchSession(sid, sess)
	return
}

func (cm *connManager) putRtr(rtr RawTransport) {
	cm.trl.Lock()
	defer cm.trl.Unlock()
	var key = rtr.TransportKey()
	if _, exists := cm.trs[*key]; exists {
		rtr.Close()
		return
	}
	tr, _ := smux.Client(rtr, SmuxClientConfig)
	trw := &transportWrapper{
		id:        conns.trcounter,
		Transport: tr,
		dialed:    make(chan struct{}),
	}
	conns.trcounter++
	close(trw.dialed)
	cm.trs[*key] = trw
	go cm.watchRtr(trw.id, rtr)
	go cm.acceptLoop(tr, key)
}

func DialSession(ctx context.Context, tk *TransportKey) (SessionId, *sessionWrapper, error) {
	var sid SessionId
	var sw *sessionWrapper
	var err error

	conns.trl.Lock()
	var tr Transport
	trw, exists := conns.trs[*tk]
	if exists && trw.Transport != nil && trw.IsClosed() {
		exists = false
	}
	if !exists {
		trw = &transportWrapper{id: conns.trcounter, dialed: make(chan struct{})}
		conns.trcounter++
		conns.trs[*tk] = trw
		var rtr RawTransport
		conns.trl.Unlock()
		rtr, err = tk.Dial(ctx)
		if err != nil {
			trw.err = err
			close(trw.dialed)
			conns.trl.Lock()
			delete(conns.trs, *tk)
			conns.trl.Unlock()
			return sid, sw, err
		}
		tr, _ = smux.Client(rtr, SmuxClientConfig)
		trw.Transport = tr
		go conns.watchRtr(trw.id, rtr)
		go conns.acceptLoop(tr, tk)
		close(trw.dialed)
	} else if trw.Transport != nil {
		tr = trw.Transport
		conns.trl.Unlock()
	} else {
		conns.trl.Unlock()
		<-trw.dialed
		err = trw.err
		if err != nil {
			return sid, sw, err
		}
		tr = trw.Transport
	}
	sess, err := tr.OpenStream()
	if err != nil {
		return sid, sw, err
	}
	sid, sw = conns.putSession(sess, tk)
	return sid, sw, err
}

var ErrSessionGoneAway = errors.New("session has gone away")

func (sid SessionId) Get() (*sessionWrapper, error) {
	sess, ok := conns.sessions.Load(sid)
	if !ok {
		return nil, ErrSessionGoneAway
	}
	return sess.(*sessionWrapper), nil

}

func (sid SessionId) Write(b []byte) (int, error) {
	sess, ok := conns.sessions.Load(sid)
	if !ok {
		return -1, ErrSessionGoneAway
	}
	return sess.(*sessionWrapper).Write(b)
}

func (sid SessionId) Read(b []byte) (int, error) {
	sess, ok := conns.sessions.Load(sid)
	if !ok {
		return -1, ErrSessionGoneAway
	}
	return sess.(*sessionWrapper).Read(b)
}

func (sid SessionId) Close() error {
	if sess, ok := conns.sessions.Load(sid); ok {
		return sess.(*sessionWrapper).Close()
	}
	return nil
}

func (sid SessionId) TransportKey() (*TransportKey, error) {
	if sess, ok := conns.sessions.Load(sid); ok {
		return sess.(*sessionWrapper).TransportKey, nil
	}
	return nil, ErrSessionGoneAway
}

func (sid SessionId) GetExtra() (any, error) {
	if sess, ok := conns.sessions.Load(sid); ok {
		return sess.(*sessionWrapper).Extra, nil
	}
	return nil, ErrSessionGoneAway
}
