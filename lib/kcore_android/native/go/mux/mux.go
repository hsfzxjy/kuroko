package mux

/*

#include <stdlib.h>

typedef struct {
	uint32_t counter[3];
	uint8_t* ReadBuf;
	uint8_t* WriteBuf;
} SessionExtra;

static SessionExtra* NewSessionExtra() {
	SessionExtra *extra = (SessionExtra*)calloc(1, sizeof(SessionExtra));
	extra->ReadBuf = (uint8_t*)calloc(1024, 1);
	extra->WriteBuf = (uint8_t*)calloc(1024, 1);
	return extra;
}

static void FreeSessionExtra(SessionExtra* extra) {
	free(extra->ReadBuf);
	free(extra->WriteBuf);
	free(extra);
}
*/
import "C"
import (
	"kmux"
	"sync/atomic"
	"time"

	"github.com/hsfzxjy/smux"
)

type SessionExtra C.SessionExtra

const (
	readFlagIndex  = 0
	writeFlagIndex = 1
	refIndex       = 2
)

func (extra *SessionExtra) SetReading() bool {
	return atomic.CompareAndSwapUint32((*uint32)(&extra.counter[readFlagIndex]), 0, 1)
}

func (extra *SessionExtra) ClearReading() {
	if !atomic.CompareAndSwapUint32((*uint32)(&extra.counter[readFlagIndex]), 1, 0) {
		panic("invalid state")
	}
}

func (extra *SessionExtra) SetWriting() bool {
	return atomic.CompareAndSwapUint32((*uint32)(&extra.counter[writeFlagIndex]), 0, 1)
}

func (extra *SessionExtra) ClearWriting() {
	if !atomic.CompareAndSwapUint32((*uint32)(&extra.counter[writeFlagIndex]), 1, 0) {
		panic("invalid state")
	}
}

func (extra *SessionExtra) IncRefCnt() {
	atomic.AddUint32((*uint32)(&extra.counter[refIndex]), 1)
}

func (extra *SessionExtra) DecRefCnt() {
	refcnt := atomic.AddUint32((*uint32)(&extra.counter[refIndex]), 0)
	if refcnt == 0 &&
		atomic.LoadUint32((*uint32)(&extra.counter[readFlagIndex])) == 0 &&
		atomic.LoadUint32((*uint32)(&extra.counter[writeFlagIndex])) == 0 {
		C.FreeSessionExtra((*C.SessionExtra)(extra))
	}
}

func newSessionExtra(sess kmux.Session) any {
	extra := (*SessionExtra)(C.NewSessionExtra())
	extra.IncRefCnt()

	go func() {
		<-sess.GetDieCh()
		extra.DecRefCnt()
	}()

	return extra
}

func Init() {
	kmux.SessionNewExtraFunc = newSessionExtra
	kmux.SmuxClientConfig = smux.DefaultConfig()
	kmux.SmuxClientConfig.MaxIdleInterval = 3 * time.Minute
}
