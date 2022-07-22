package bluetooth

import (
	"fmt"
	"kmux"
	"kuroko-linux/internal"
	"kuroko-linux/internal/exc"
	"sync"

	"github.com/muka/go-bluetooth/bluez/profile/device"
	"golang.org/x/sys/unix"
)

type BluetoothSocket struct {
	Fd  int
	dev *device.Device1

	sockAddr *unix.SockaddrRFCOMM
	tk       kmux.TransportKey

	closedCh chan struct{}
	dieOnce  sync.Once
}

func NewBluetoothSocket(fd int, sockAddr *unix.SockaddrRFCOMM, dev *device.Device1) *BluetoothSocket {
	ret := &BluetoothSocket{
		Fd:  fd,
		dev: dev,

		sockAddr: sockAddr,

		closedCh: make(chan struct{}),
	}
	ret.tk[0] = byte(kmux.TT_BLUETOOTH)
	copy(ret.tk[1:7], sockAddr.Addr[:])
	return ret
}

func (s *BluetoothSocket) Name() string { return "BluetoothSocket" }

func (s *BluetoothSocket) Close() error {
	if s.dev != nil {
		s.dev.DisconnectProfile(internal.BLUETOOTH_UUID)
	}
	err := exc.Wrap(unix.Close(s.Fd))
	s.dieOnce.Do(func() { close(s.closedCh) })
	return err
}

func (s *BluetoothSocket) Read(p []byte) (n int, err error) {
	n, err = unix.Read(s.Fd, p)
	if err != nil {
		fmt.Printf("read %+v", string(p))
	}
	return n, exc.Wrap(err)
}

func (s *BluetoothSocket) Write(p []byte) (n int, err error) {
	n, err = unix.Write(s.Fd, p)
	return n, exc.Wrap(err)
}

func (s *BluetoothSocket) ClosedCh() <-chan struct{}        { return s.closedCh }
func (s *BluetoothSocket) TransportKey() *kmux.TransportKey { return &s.tk }
