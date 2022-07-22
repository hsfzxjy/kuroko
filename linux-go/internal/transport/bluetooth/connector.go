package bluetooth

import (
	"context"
	"kmux"
	"kuroko-linux/bluez"
	"kuroko-linux/internal"
	"kuroko-linux/internal/exc"
	"sync"
	"syscall"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/muka/go-bluetooth/bluez/profile"
	"golang.org/x/sys/unix"
)

func ConnectBluetooth(ctx context.Context, ba [6]byte) (tr kmux.RawTransport, err error) {
	addr := Ba2Str(ba)

	adapter, err := bluez.Adapter1.Get()
	if err != nil {
		err = exc.Wrap(err)
		return
	}
	dev, err := adapter.GetDeviceByAddress(addr)
	if err != nil {
		err = exc.Wrap(err)
		return
	}

	path := dev.Path()
	connManager.l.Lock()
	if _, ok := connManager.m[path]; ok {
		connManager.l.Unlock()
		err = exc.New("bluetooth connection not available")
		return
	}

	ch := make(chan dbus.UnixFDIndex)
	connManager.m[path] = ch

	connManager.l.Unlock()

	go dev.ConnectProfile(internal.BLUETOOTH_UUID)

	var fd dbus.UnixFDIndex
	var found = false
	select {
	case fd = <-ch:
		found = true
		connManager.l.Lock()
	case <-time.After(5 * time.Second):
		err = exc.New("connection time out")
		connManager.l.Lock()
		select {
		case fd = <-ch:
			found = true
			err = nil
		default:
		}
	}
	delete(connManager.m, path)
	connManager.l.Unlock()

	if found {
		var sa unix.Sockaddr
		var fd int = int(fd)
		sa, err = unix.Getsockname(fd)
		if err != nil {
			err = exc.Wrap(err)
			goto CLOSE
		}
		<-time.After(100 * time.Millisecond)
		if err = syscall.SetNonblock(fd, false); err != nil {
			err = exc.Wrap(err)
			goto CLOSE
		}
		tr = NewBluetoothSocket(fd, sa.(*unix.SockaddrRFCOMM), dev)
		return
	CLOSE:
		unix.Close(fd)
	}

	if err != nil {
		dev.DisconnectProfile(internal.BLUETOOTH_UUID)
		err = exc.Wrap(err)
	}

	return
}

type Profile int

type connectionManager struct {
	m map[dbus.ObjectPath]chan<- dbus.UnixFDIndex
	l sync.Mutex
}

var connManager = connectionManager{
	m: make(map[dbus.ObjectPath]chan<- dbus.UnixFDIndex),
	l: sync.Mutex{},
}

func (Profile) NewConnection(path dbus.ObjectPath, fd dbus.UnixFDIndex, props map[string]any) *dbus.Error {
	connManager.l.Lock()
	defer connManager.l.Unlock()

	ch, ok := connManager.m[path]
	if !ok {
		unix.Close(int(fd))
		return &profile.ErrRejected
	}
	ch <- fd
	close(ch)
	return nil
}

func (Profile) RequestDisconnection(path dbus.ObjectPath) *dbus.Error {
	return nil
}

func init() {
	kmux.SetBluetoothDialer(ConnectBluetooth)
}
