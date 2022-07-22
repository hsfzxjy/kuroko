package discovery

import (
	"encoding/gob"
	"errors"
	"kuroko-linux/internal/log"
	"kuroko-linux/util"
	"net/rpc"

	"github.com/hsfzxjy/go-srpc"
)

type LinkState int32

const (
	LS_NONE LinkState = iota
	LS_BONDING
	LS_BONDED
	LS_UNBONDED
	LS_VERIFYING
	LS_RETRYING
	LS_VERIFIED
	LS_ERROR
	LS_UNVERIFIED
)

var linkStateMessages = map[LinkState]string{
	LS_NONE:       "",
	LS_BONDING:    "bonding",
	LS_BONDED:     "bonded",
	LS_UNBONDED:   "unbonded",
	LS_VERIFYING:  "verifying",
	LS_VERIFIED:   "verified",
	LS_ERROR:      "error",
	LS_UNVERIFIED: "unverified",
}

func (ls LinkState) String() string {
	if str, ok := linkStateMessages[ls]; ok {
		return str
	} else {
		return ""
	}
}

type Linker interface {
	StartLink(ProbeResult) (*util.StreamResult[LinkState], error)
	StopLink(ProbeResult)
}

var linkers = map[string]Linker{
	"bluetooth": newBluetoothLinker(),
}

type LinkerManager int

type LinkerArg struct {
	Typ    string
	Result ProbeResult
}

func (lm *LinkerManager) Name() string { return "LinkerManager" }

func (lm *LinkerManager) StartLink(arg *LinkerArg, s *srpc.Session) error {
	return srpc.S(func() error {
		linker, ok := linkers[arg.Typ]
		if !ok {
			return errors.New("no such Linker")
		}
		res, err := linker.StartLink(arg.Result)
		if err != nil {
			return err
		}
	LOOP:
		for {
			select {
			case state, ok := <-res.C:
				if !ok {
					break LOOP
				}
				s.PushValue(state)
			case <-s.Canceled():
				linker.StopLink(arg.Result)
				break LOOP
			}
		}
		log.For(lm).Errorf(res.Err, "Error while linking")
		return res.Err
	}, s, nil)
}

func init() {
	gob.Register(LinkState(0))
	gob.Register(new(LinkerArg))
	rpc.RegisterName("LinkerManager", new(LinkerManager))
}
