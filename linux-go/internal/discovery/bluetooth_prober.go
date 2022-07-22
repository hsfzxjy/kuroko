package discovery

import (
	"encoding/gob"
	"kmux"
	"kuroko-linux/bluez"
	"kuroko-linux/internal/exc"
	"kuroko-linux/internal/log"
	"kuroko-linux/models"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
)

type BluetoothDevice struct {
	ObjPath dbus.ObjectPath
	Address string
	Name    string
	State   LinkState
}

func (dev *BluetoothDevice) DisplayName() string      { return dev.Name }
func (dev *BluetoothDevice) DisplayAddr() string      { return dev.Address }
func (dev *BluetoothDevice) Type() kmux.TransportType { return kmux.TT_BLUETOOTH }
func (dev *BluetoothDevice) LinkState() LinkState     { return dev.State }
func (dev *BluetoothDevice) Peer() *models.Peer       { return nil }

func newBluetoothDevice(dev *device.Device1) *BluetoothDevice {
	prop := dev.Properties
	state := LS_NONE
	if paired, err := dev.GetPaired(); err == nil && paired {
		state = LS_BONDED
	}
	return &BluetoothDevice{
		ObjPath: dev.Path(),
		Name:    prop.Name,
		Address: prop.Address,
		State:   state,
	}
}

type BluetoothProber struct {
	l       *sync.Mutex
	ch      chan ProbeResult
	stopCh  chan struct{}
	adapter *adapter.Adapter1

	cancelNotifier func()
}

func newBluetoothProber() *BluetoothProber {
	return &BluetoothProber{
		l:       &sync.Mutex{},
		ch:      make(chan ProbeResult),
		stopCh:  nil,
		adapter: bluez.Adapter1.MustGet(),

		cancelNotifier: nil,
	}
}

func (p *BluetoothProber) Name() string { return "BluetoothProber" }

func (p *BluetoothProber) setupNotifier() error {
	discovered, err := p.adapter.GetDevices()
	if err != nil {
		return err
	}

	ch, cancel, err := p.adapter.OnDeviceDiscovered()
	if err != nil {
		return err
	}
	p.cancelNotifier = cancel
	p.stopCh = make(chan struct{})
	done := make(chan struct{})
	go func() {
		var err error
		var dev *device.Device1

		close(done)

		for _, dev := range discovered {
			p.ch <- newBluetoothDevice(dev)
		}

		for obj := range ch {
			if obj.Type == adapter.DeviceRemoved {
				continue
			}
			dev, err = device.NewDevice1(obj.Path)
			if err != nil {
				log.For(p).Errorf(exc.Wrap(err), "unable to new device")
				continue
			}
			p.ch <- newBluetoothDevice(dev)
		}
	}()
	<-done
	return nil
}

func (p *BluetoothProber) Start(config any) error {
	if !p.l.TryLock() {
		return ErrAlreadyStarted
	}
	var err error
	if err = p.setupNotifier(); err != nil {
		goto EXCEPT
	}
	if err = p.adapter.StartDiscovery(); err != nil {
		p.cancelNotifier()
		close(p.stopCh)
		goto EXCEPT
	}
	return nil
EXCEPT:
	p.l.Unlock()
	return err
}

func (p *BluetoothProber) Stop() error {
	defer p.l.Unlock()
	if p.l.TryLock() {
		return nil
	}
	defer func() {
		p.cancelNotifier()
		close(p.stopCh)
		p.stopCh = nil
	}()

	if err := p.adapter.StopDiscovery(); err != nil {
		return err
	}

	return nil
}

func (p *BluetoothProber) Recieve() <-chan ProbeResult {
	return p.ch
}

func (p *BluetoothProber) Stopped() <-chan struct{} {
	return p.stopCh
}

func init() {
	gob.Register(dbus.ObjectPath(""))
	gob.Register(new(device.Device1Properties))
	gob.Register(new(BluetoothDevice))
}
