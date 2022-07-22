package services

import "errors"

var ErrBadState = errors.New("bad state")

type serviceState int32

const (
	SS_STARTING serviceState = iota
	SS_STOPPING
	SS_STARTED
	SS_STOPPED
)

var ssTransMap = map[serviceState][2]serviceState{
	SS_STARTED:  {SS_STARTING, SS_STOPPING},
	SS_STARTING: {SS_STOPPED, SS_STARTED},
	SS_STOPPED:  {SS_STOPPING, SS_STARTING},
	SS_STOPPING: {SS_STARTED, SS_STOPPED},
}

func (ss *serviceState) Next() serviceState {
	return ssTransMap[*ss][1]
}

func (ss *serviceState) Prev() serviceState {
	return ssTransMap[*ss][0]
}
