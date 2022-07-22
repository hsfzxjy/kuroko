package rpc

import (
	"kuroko-linux/services"
)

func (Core) ServiceOn(name string, reply *int) error {
	return services.Manager.SwitchOn(name)
}

func (Core) ServiceOff(name string, reply *int) error {
	return services.Manager.SwitchOff(name)
}
