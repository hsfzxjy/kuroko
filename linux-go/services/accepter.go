package services

import (
	"kmux"
	"kuroko-linux/endpoints/inward"
	"kuroko-linux/internal/log"
	"kuroko-linux/internal/session"
	"sync"
)

type accepter struct {
	typ  kmux.TransportType
	name string
}

var accepterLoopOnce = sync.Once{}

func (a *accepter) Name() string { return a.name }

func (a *accepter) Run(barrier *barrier) error {
	defer barrier.markStopped()

	accepterLoopOnce.Do(func() {
		go kmux.AccepterManager.Loop()
		go func() {
			for sid := range kmux.AccepterManager.SessionCh() {
				sess, err := session.WrapSession(sid)
				if err != nil {
					log.Name("session").Fields("sid", sid).Errorf(err, "error on WrapSession()")
					continue
				}
				go inward.Route(sess)
			}
		}()
	})

	ac := kmux.AccepterManager.Get(a.typ)
	ac.Start()
	var st kmux.AccepterState
	for st = range ac.StateCh() {
		if st.IsTerminal() {
			break
		}
	}

	switch st {
	case kmux.AS_ERROR:
	case kmux.AS_ENDED:
		return ac.LastError()
	}
	barrier.markStartedAndWaitForStop(func() {
		ac.Stop()
		for st = range ac.StateCh() {
			if st.IsTerminal() {
				break
			}
		}
	})

	return nil
}
