package rpc

import (
	"kuroko-linux/internal/exc"
	"net/rpc"

	"github.com/hsfzxjy/go-srpc"
)

func Client() (ret *srpc.Client, err error) {
	var c *rpc.Client
	if c, err = rpc.DialHTTP("tcp", "127.0.0.1:27182"); err != nil {
		exc.Wrap(err)
		return
	}
	ret = srpc.WrapClient(c)
	return
}

type Core int

func (Core) Name() string { return "RpcCore" }

func init() {
	rpc.RegisterName("Core", Core(0))
}
