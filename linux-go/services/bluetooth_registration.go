package services

import (
	"kuroko-linux/bluez"
	"kuroko-linux/internal/exc"
	"kuroko-linux/internal/transport/bluetooth"

	"github.com/muka/go-bluetooth/bluez/profile/agent"
)

type bluetoothReg int

func (bluetoothReg) Name() string { return "BluetoothReg" }

func (bluetoothReg) Run(barrier *barrier) error {
	defer barrier.markStopped()

	adapter, err := bluez.Adapter1.Get()
	if err != nil {
		return err
	}
	conn := adapter.Client().GetConnection()

	err = conn.Export(
		bluetooth.Profile(0),
		bluetooth.KurokoProfilePath,
		"org.bluez.Profile1")
	if err != nil {
		return exc.Wrap(err)
	}

	err = conn.Export(
		bluetooth.Agent(0),
		bluetooth.KurokoAgentPath,
		"org.bluez.Agent1",
	)
	if err != nil {
		return exc.Wrap(err)
	}

	am, err := agent.NewAgentManager1()
	if err != nil {
		return exc.Wrap(err)
	}
	if err = am.RegisterAgent(
		bluetooth.KurokoAgentPath, "NoInputNoOutput",
	); err != nil {
		return exc.Wrap(err)
	}

	barrier.markStartedAndWaitForStop(func() {})

	conn.Export(nil, bluetooth.KurokoProfilePath, "org.bluez.Profile1")
	conn.Export(nil, bluetooth.KurokoAgentPath, "org.bluez.Agent1")
	am.UnregisterAgent(bluetooth.KurokoAgentPath)

	return nil
}

type Agent int
