package bluez

import (
	"kuroko-linux/internal/exc"
	"sync"

	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/profile"
)

type cachedGlobal[T any] struct {
	l      sync.Mutex
	data   *T
	getter func() (*T, error)
}

func (g *cachedGlobal[T]) Get() (ret *T, err error) {
	g.l.Lock()
	defer g.l.Unlock()

	if g.data == nil {
		var err error
		ret, err = g.getter()
		err = exc.Wrap(err)
		if err == nil {
			g.data = ret
		}
	} else {
		ret = g.data
	}

	return
}

func (g *cachedGlobal[T]) MustGet() *T {
	if data, err := g.Get(); err != nil {
		panic(exc.Wrap(err))
	} else {
		return data
	}
}

func newCachedGlobal[T any](
	getter func() (*T, error),
) *cachedGlobal[T] {
	return &cachedGlobal[T]{
		getter: getter,
		l:      sync.Mutex{},
		data:   nil,
	}
}

var Adapter1 = newCachedGlobal(adapter.GetDefaultAdapter)
var ProfileManager1 = newCachedGlobal(profile.NewProfileManager1)
