package services

import (
	"kuroko-linux/internal/exc"
	"net"
	"net/http"
	"net/rpc"
)


type rpcServer int

func (rpcServer) Name() string { return "RpcServer" }

func (rpcServer) Run(barrier *barrier) error {
	defer barrier.markStopped()

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":27182")
	if err != nil {
		return exc.Wrap(err)
	}

	go barrier.markStartedAndWaitForStop(func() {
		listener.Close()
	})

	http.Serve(listener, nil)

	return nil
}
