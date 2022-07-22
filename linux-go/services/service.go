package services

import (
	"context"
	"kuroko-linux/internal/log"
	"sync"
)

type ServiceWorker interface {
	Name() string
	Run(*barrier) error
}

type Service struct {
	worker  ServiceWorker
	wg      *sync.WaitGroup
	barrier barrier
}

var (
	stopCtx, markStop = context.WithCancel(context.Background())
)

func NewService(worker ServiceWorker, wg *sync.WaitGroup) *Service {
	return &Service{worker: worker, wg: wg, barrier: newBarrier()}
}

func (ser *Service) Name() string { return "Service" }

func (ser *Service) Switch(state serviceState) error {
	return ser.barrier.setState(state)
}

func (ser *Service) runWorker() {
	lg := log.For(ser)
	defer func() {
		if err := recover(); err != nil {
			lg.Error(err)
		}
		lg.Info("%s stopped", ser.worker.Name())
	}()
	lg.Info("%s started", ser.worker.Name())
	if err := ser.worker.Run(&ser.barrier); err != nil {
		lg.Error(err)
	}

}

func (ser *Service) Serve() {
	defer ser.wg.Done()

	var stopped = false
	go func() {
		<-stopCtx.Done()
		stopped = true
		ser.barrier.setStopFlag()
	}()
	for {
		if !ser.barrier.waitFor(SS_STARTING) {
			break
		}
		ser.runWorker()
		if stopped {
			break
		}
	}
}
