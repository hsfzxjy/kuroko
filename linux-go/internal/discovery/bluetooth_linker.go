package discovery

import (
	"encoding/gob"
	"kmux"
	"sync"

	"kuroko-linux/bluez"
	"kuroko-linux/endpoints/outward"
	"kuroko-linux/internal"
	"kuroko-linux/internal/session"
	"kuroko-linux/util"

	"github.com/muka/go-bluetooth/bluez/profile"
	"github.com/muka/go-bluetooth/bluez/profile/device"
)

type BluetoothLinker struct {
	l  *sync.Mutex
	ch chan LinkState
}

func newBluetoothLinker() *BluetoothLinker {
	return &BluetoothLinker{
		l:  &sync.Mutex{},
		ch: make(chan LinkState),
	}
}

func (bl *BluetoothLinker) StartLink(pr ProbeResult) (res *util.StreamResult[LinkState], err error) {
	if !bl.l.TryLock() {
		err = ErrAlreadyStarted
		return
	}
	res = util.NewStreamResult[LinkState]()

	go func() {
		var err error
		var sess *session.Session
		defer bl.l.Unlock()
		defer close(res.C)

		bd := pr.(*BluetoothDevice)
		var dev *device.Device1
		dev, err = device.NewDevice1(bd.ObjPath)
		if err != nil {
			goto EXCEPT
		}

		if !dev.Properties.Paired {
			err = dev.SetTrusted(true)
			if err != nil {
				goto EXCEPT
			}

			res.C <- LS_BONDING
			err = dev.Pair()
			if bluez.ErrAlreadyExists.Is(err) {
				res.C <- LS_BONDED
			} else if err != nil {
				res.C <- LS_UNBONDED
				goto EXCEPT
			}
		}
		res.C <- LS_BONDED

		res.C <- LS_VERIFYING
		sess, err = session.DialSession(internal.Bi(
			kmux.TT_BLUETOOTH, dev.Properties.Address,
		))
		if err != nil {
			res.C <- LS_ERROR
			goto EXCEPT
		}
		if err = outward.ExchangeIdentity(sess); err != nil {
			res.C <- LS_UNVERIFIED
			goto EXCEPT
		}
		res.C <- LS_VERIFIED

	EXCEPT:
		res.Err = err
	}()

	return
}

func (bl *BluetoothLinker) StopLink(pr ProbeResult) {

}

func init() {
	gob.Register(profile.ErrDoesNotExist)
}
