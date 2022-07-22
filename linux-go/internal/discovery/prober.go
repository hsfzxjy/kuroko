package discovery

import (
	"kmux"
	"kuroko-linux/models"
	"net/rpc"

	"github.com/hsfzxjy/go-srpc"
	"github.com/pkg/errors"
)

type ProbeResult interface {
	Type() kmux.TransportType
	DisplayAddr() string
	DisplayName() string
	Peer() *models.Peer
	LinkState() LinkState
}

type Prober interface {
	Start(config any) error
	Stop() error
	Recieve() <-chan ProbeResult
	Stopped() <-chan struct{}
}

var ErrAlreadyStarted = errors.New("already started")

var probers = map[string]Prober{
	"bluetooth": newBluetoothProber(),
}

type ProberManager int

func (*ProberManager) Start(name string, s *srpc.Session) error {
	return srpc.S(func() error {
		var d Prober
		d, ok := probers[name]
		if !ok {
			return errors.New("no such Prober")
		}
		if err := d.Start(nil); err != nil {
			return err
		}
		defer d.Stop()
	LOOP:
		for {
			select {
			case dev := <-d.Recieve():
				s.PushValue(dev)
			case <-d.Stopped():
				break LOOP
			case <-s.Canceled():
				break LOOP
			}
		}
		return nil
	}, s, nil)
}

func init() {
	rpc.RegisterName("ProberManager", new(ProberManager))
}
