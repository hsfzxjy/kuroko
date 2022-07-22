package services

import (
	"kmux"
	"kuroko-linux/internal/exc"
	"sync"
)

type ServiceManager struct {
	sers map[string]*Service
	wg   *sync.WaitGroup
}

var Manager = newServiceManager()

func newServiceManager() ServiceManager {
	ret := ServiceManager{wg: &sync.WaitGroup{}}
	ret.sers = map[string]*Service{
		"btserver": NewService(&accepter{kmux.TT_BLUETOOTH, "BluetoothWatcher"}, ret.wg),
		"signal":   NewService(signalServer(0), ret.wg),
		"rpc":      NewService(rpcServer(0), ret.wg),
		"btreg":    NewService(bluetoothReg(0), ret.wg),
	}
	return ret
}

func (m *ServiceManager) ServeAll() {
	m.wg.Add(len(m.sers))
	for _, ser := range m.sers {
		go func(ser *Service) {
			ser.Switch(SS_STARTING)
			ser.Serve()
		}(ser)
	}
	m.wg.Wait()
}

var ErrNoSuchService = exc.New("no such service")

var allowedServices = map[string]bool{
	"btserver": true,
}

func (m *ServiceManager) switchService(name string, state serviceState) error {
	if _, ok := allowedServices[name]; !ok {
		return ErrNoSuchService
	}
	return m.sers[name].Switch(state)
}

func (m *ServiceManager) SwitchOn(name string) error {
	return m.switchService(name, SS_STARTING)
}

func (m *ServiceManager) SwitchOff(name string) error {
	return m.switchService(name, SS_STOPPING)
}
