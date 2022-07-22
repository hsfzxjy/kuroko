package bluetooth

import (
	"fmt"
	"kmux"
	"kuroko-linux/bluez"
	"kuroko-linux/internal"
	"kuroko-linux/internal/exc"
	"sync"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"

	"golang.org/x/sys/unix"
)

var serviceRecordTemplate string = `
<record>
<attribute id="0x0000">
	<uint32 value="0x00010009" />
</attribute>
<attribute id="0x0001">
	<sequence>
		<uuid value="%s" />
		<uuid value="0x1101" />
	</sequence>
</attribute>
<attribute id="0x0003">
	<uuid value="%s" />
</attribute>
<attribute id="0x0004">
	<sequence>
		<sequence>
			<uuid value="0x0100" />
		</sequence>
		<sequence>
			<uuid value="0x0003" />
			<uint8 value="0x%02x" />
		</sequence>
	</sequence>
</attribute>
<attribute id="0x0005">
	<sequence>
		<uuid value="0x1002" />
	</sequence>
</attribute>
<attribute id="0x0009">
	<sequence>
		<sequence>
			<uuid value="0x1101" />
			<uint16 value="0x0100" />
		</sequence>
	</sequence>
</attribute>
<attribute id="0x0100">
	<text value="KurokoServer" />
</attribute>
</record>`
var KurokoProfilePath = dbus.ObjectPath("/kuroko/bluetooth")

type BluetoothServerSocket struct {
	Fd      int
	Channel uint8

	adving  bool
	advLock sync.Mutex

	pollFds [1]unix.PollFd
}

func NewBluetoothServerSocket(addr string, channel uint8, backlog int) (ret *BluetoothServerSocket, err error) {
	sock, err := unix.Socket(syscall.AF_BLUETOOTH, syscall.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	unix.FcntlInt(uintptr(sock), unix.F_SETFL, unix.O_NONBLOCK)
	if err != nil {
		err = exc.Wrap(err)
		return
	}

	sock_addr := &unix.SockaddrRFCOMM{
		Addr:    Str2Ba(addr),
		Channel: channel, // PORT_ANY
	}

	err = unix.Bind(sock, sock_addr)
	if err != nil {
		err = exc.Wrap(err)
		return
	}

	err = unix.Listen(sock, backlog)
	if err != nil {
		err = exc.Wrap(err)
		return
	}

	sa, err := unix.Getsockname(sock)
	if err != nil {
		err = exc.Wrap(err)
		return
	}
	ret = &BluetoothServerSocket{Fd: sock, Channel: sa.(*unix.SockaddrRFCOMM).Channel}
	ret.pollFds[0] = unix.PollFd{Fd: int32(sock), Events: unix.POLLIN}

	err = ret.Advertise()

	if err != nil {
		return nil, exc.Wrap(err)
	}

	return
}

func (s *BluetoothServerSocket) Close() error {
	s.StopAdvertise()
	unix.Shutdown(s.Fd, syscall.SHUT_RDWR)
	return unix.Close(s.Fd)
}

func (s *BluetoothServerSocket) Accept() (tr kmux.RawTransport, err error) {
	for {
		var n int
		n, err = unix.Poll(s.pollFds[:], 100)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			err = exc.Wrap(err)
			return
		}
		if n == 0 {
			continue
		}
		break
	}
	fd, sa, err := unix.Accept(s.Fd)
	if err != nil {
		err = exc.Wrap(err)
		return
	}
	tr = NewBluetoothSocket(fd, sa.(*unix.SockaddrRFCOMM), nil)
	return
}

func (s *BluetoothServerSocket) Advertise() error {
	s.advLock.Lock()
	defer s.advLock.Unlock()

	if s.adving {
		return exc.New("Already advertising")
	}

	pm, err := bluez.ProfileManager1.Get()
	if err != nil {
		return exc.Wrap(err)
	}

	err = pm.UnregisterProfile(KurokoProfilePath)
	if err != nil && !bluez.ErrDoesNotExist.Is(err) {
		return exc.Wrap(err)
	}

	serviceUUID := internal.BLUETOOTH_UUID
	serviceRecord := fmt.Sprintf(serviceRecordTemplate, serviceUUID, serviceUUID, s.Channel)
	err = pm.RegisterProfile(KurokoProfilePath,
		serviceUUID,
		map[string]any{
			"ServiceRecord": serviceRecord,
		},
	)

	if err == nil {
		s.adving = true
	}

	return exc.Wrap(err)
}

func (s *BluetoothServerSocket) StopAdvertise() error {
	s.advLock.Lock()
	defer s.advLock.Unlock()

	if !s.adving {
		return exc.New("not advertising")
	}
	s.adving = false

	pm, err := bluez.ProfileManager1.Get()
	if err != nil {
		return exc.Wrap(err)
	}
	err = pm.UnregisterProfile(KurokoProfilePath)

	if err != nil && !bluez.ErrDoesNotExist.Is(err) {
		return exc.Wrap(err)
	}

	err = pm.RegisterProfile(KurokoProfilePath, internal.BLUETOOTH_UUID, map[string]interface{}{"Role": "client"})

	return exc.Wrap(err)
}

func init() {
	kmux.AccepterManager.Register(
		kmux.TT_BLUETOOTH,
		func() (kmux.Listener, error) {
			return NewBluetoothServerSocket(
				"00:00:00:00:00:00",
				0, 1,
			)
		},
	)
}
