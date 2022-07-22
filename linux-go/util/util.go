package util

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func EnsureCondition(ctx context.Context, dur time.Duration, cond func() bool) bool {
	ticker := time.NewTicker(dur)
	defer ticker.Stop()

LOOP:
	for {
		if cond() {
			break LOOP
		}
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
		}
	}
	return true
}

func OnInterrupt(handler func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-ch
	signal.Stop(ch)
	handler()
}

type StreamResult[T any] struct {
	C   chan T
	Err error
}

func NewStreamResult[T any]() *StreamResult[T] {
	return &StreamResult[T]{
		C:   make(chan T),
		Err: nil,
	}
}

func FormatList(format string, args []any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

func ArgsToMap(args []any) map[string]any {
	var key string
	var ok bool
	var ret = make(map[string]any)
	for idx, arg := range args {
		if idx%2 == 0 {
			key, ok = arg.(string)
			if !ok {
				panic(fmt.Sprintf("%d-th arg must be a string", idx))
			}
		} else {
			ret[key] = arg
		}
	}
	return ret
}

func CopyMergeMap[K comparable, V any](src map[K]V, update map[K]V) map[K]V {
	ret := make(map[K]V)
	for k, v := range src {
		ret[k] = v
	}
	for k, v := range update {
		ret[k] = v
	}
	return ret
}

func MergeMap[K comparable, V any](dst map[K]V, src map[K]V) map[K]V {
	if dst == nil {
		dst = make(map[K]V)
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
