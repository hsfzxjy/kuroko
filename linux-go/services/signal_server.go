package services

import (
	"context"
	"kuroko-linux/internal/log"
	"kuroko-linux/util/signals"
	"os"
	"syscall"
)

type signalServer int

func (signalServer) Name() string { return "SignalServer" }

func termHandler(sig os.Signal) error {
	switch sig {
	case syscall.SIGINT:
	case syscall.SIGTERM:
	case syscall.SIGQUIT:
	default:
		return nil
	}
	log.For(signalServer(0)).Info("Recieved signal %v, terminating...\n", sig)
	markStop()
	return signals.ErrStop
}

func (signalServer) Run(barrier *barrier) error {
	defer barrier.markStopped()

	ctx, cancel := context.WithCancel(context.Background())
	go barrier.markStartedAndWaitForStop(cancel)

	signals.SetSigHandler(termHandler, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	return signals.ServeSignals(ctx)
}
