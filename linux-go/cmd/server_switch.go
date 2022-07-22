package cmd

import "kuroko-linux/internal/rpc"

type onCommand struct {
	Service struct {
		Name string `positional-arg-name:"<SERVICE>" choice:"btserver"`
	} `positional-args:"true" required:"true"`
}

func (c *onCommand) Execute(args []string) error {
	client, err := rpc.Client()
	if err != nil {
		return err
	}
	var reply int
	return client.Call("Core.ServiceOn", c.Service.Name, &reply)
}

type offCommand struct {
	Service struct {
		Name string `positional-arg-name:"<SERVICE>" choice:"btserver"`
	} `positional-args:"true" required:"true"`
}

func (c *offCommand) Execute(args []string) error {
	client, err := rpc.Client()
	if err != nil {
		return err
	}
	var reply int
	return client.Call("Core.ServiceOff", c.Service.Name, &reply)
}

func init() {
	CLIParser.AddCommand("on", "Turn on a service", "", &onCommand{})
	CLIParser.AddCommand("off", "Turn on a service", "", &offCommand{})
}
