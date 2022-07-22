package cmd

import (
	"fmt"
	"kuroko-linux/internal/rpc"
	"kuroko-linux/models"
	"path/filepath"
)

type sendCommand struct {
	Filename struct {
		Value string `positional-arg-name:"<FILE>"`
	} `positional-args:"true" required:"true"`
}

func (c *sendCommand) Execute(args []string) error {
	client, err := rpc.Client()
	if err != nil {
		return err
	}
	var bridges []*models.Bridge
	err = client.Call("Core.GetBridgeInfoList", 0, &bridges)
	if err != nil {
		return err
	}
	filename := c.Filename.Value
	filename, err = filepath.Abs(filename)
	if err != nil {
		return err
	}
	var success bool
	fmt.Printf("%+v\n", bridges[0])
	return client.Call("Core.SendFile", rpc.SendArgs{
		Filename: filename,
		Bridge:   bridges[0],
	}, &success)
}

func init() {
	CLIParser.AddCommand("send", "send a file", "", new(sendCommand))
}
