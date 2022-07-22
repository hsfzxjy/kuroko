package kmux

import (
	"sync"
)

type AccepterState int32

const (
	AS_STARTED AccepterState = iota
	AS_STARTING
	AS_ERROR
	AS_ENDED
	AS_ENDING
)

func (s AccepterState) IsTerminal() bool {
	return s == AS_ERROR || s == AS_STARTED || s == AS_ENDED
}

type accepterManager struct {
	m      map[TransportType]*Accepter
	rtrCh  chan RawTransport
	sessCh chan SessionId
}

var AccepterManager = newAccepterManager()

func newAccepterManager() *accepterManager {
	ret := &accepterManager{
		m:      make(map[TransportType]*Accepter),
		rtrCh:  make(chan RawTransport),
		sessCh: make(chan SessionId),
	}
	ret.Register(TT_LANv4, newLANv4Listener)
	return ret
}

func (am *accepterManager) SessionCh() <-chan SessionId { return am.sessCh }

func (am *accepterManager) Register(typ TransportType, newListener NewListenerFunc) {
	am.m[typ] = NewAccepter(newListener, am.rtrCh, typ)
}

func (am *accepterManager) StartAll() {
	for _, accepter := range am.m {
		accepter.Start()
	}
}

func (am *accepterManager) Get(typ TransportType) *Accepter {
	return am.m[typ]
}

func (am *accepterManager) Loop() {
	for {
		conns.putRtr(<-am.rtrCh)
	}
}

type Accepter struct {
	newListener NewListenerFunc
	listener    Listener

	typ TransportType

	rtrCh     chan<- RawTransport
	state     AccepterState
	stateCh   chan AccepterState
	lastErr   error
	stopFlag  bool
	startFlag bool
	l         sync.Mutex
}

func NewAccepter(
	newListener NewListenerFunc,
	rtrCh chan<- RawTransport,
	typ TransportType,
) *Accepter {
	return &Accepter{
		typ:         typ,
		newListener: newListener,
		rtrCh:       rtrCh,
		stateCh:     make(chan AccepterState, 8),
		state:       AS_ENDED,
	}
}

func (a *Accepter) LastError() error { return a.lastErr }
func (a *Accepter) putState(st AccepterState) {
	a.state = st
	a.stateCh <- st
}
func (a *Accepter) StateCh() <-chan AccepterState { return a.stateCh }
func (a *Accepter) State() AccepterState          { return a.state }

func (a *Accepter) Start() {
	a.l.Lock()
	switch a.state {
	case AS_STARTED:
	case AS_STARTING:
		a.l.Unlock()
		return
	case AS_ENDING:
		a.startFlag = true
		a.l.Unlock()
		return
	}
	a.putState(AS_STARTING)
	a.stopFlag = false
	a.lastErr = nil
	a.l.Unlock()

	var err error
	listener, err := a.newListener()
	a.l.Lock()
	if err != nil {
		a.putState(AS_ERROR)
		a.lastErr = err
		a.l.Unlock()
		return
	}
	if a.stopFlag {
		listener.Close()
		a.putState(AS_ENDED)
		a.l.Unlock()
		return
	}
	a.putState(AS_STARTED)
	a.listener = listener
	a.l.Unlock()

	go func() {
		for {
			rtr, err := listener.Accept()
			a.l.Lock()
			if err != nil {
				switch a.state {
				case AS_ENDING:
					a.putState(AS_ENDED)
					if a.startFlag {
						defer a.Start()
					}
				default:
					a.putState(AS_ERROR)
					a.lastErr = err
				}
				a.l.Unlock()
				return
			}
			a.l.Unlock()
			a.rtrCh <- rtr
		}
	}()
}

func (a *Accepter) Stop() {
	a.l.Lock()
	defer a.l.Unlock()
	switch a.state {
	case AS_ENDED:
	case AS_ENDING:
	case AS_ERROR:
		return
	case AS_STARTING:
		a.stopFlag = true
		return
	}
	a.startFlag = false
	a.putState(AS_ENDING)
	a.listener.Close()
	a.listener = nil
}
