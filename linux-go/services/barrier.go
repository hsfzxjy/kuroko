package services

import "sync"

type barrier struct {
	state serviceState
	cond  sync.Cond

	stopFlag bool
}

func newBarrier() barrier {
	return barrier{
		state:    SS_STOPPED,
		cond:     *sync.NewCond(&sync.Mutex{}),
		stopFlag: false,
	}
}

func (b *barrier) _setStateUnsafe(new serviceState) error {
	if new != b.state.Next() {
		return ErrBadState
	}

	if new == SS_STARTED && b.stopFlag {
		new = SS_STOPPING
	}
	b.state = new

	if new == SS_STARTING || new == SS_STOPPING {
		b.cond.Broadcast()
	}
	return nil
}

func (b *barrier) setState(new serviceState) error {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	return b._setStateUnsafe(new)
}

func (b *barrier) waitFor(target serviceState) bool {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	for target != b.state && !b.stopFlag {
		b.cond.Wait()
	}

	return target == b.state
}

func (b *barrier) setStopFlag() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	b.stopFlag = true
	if b.state == SS_STARTED {
		b._setStateUnsafe(SS_STOPPING)
	} else if b.state == SS_STOPPED {
		b.cond.Broadcast()
	}
}

func (b *barrier) markStartedAndWaitForStop(onStop func()) {
	b.setState(SS_STARTED)
	b.waitFor(SS_STOPPING)
	onStop()
}

func (b *barrier) markStopped() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	b.state = SS_STOPPED
}
