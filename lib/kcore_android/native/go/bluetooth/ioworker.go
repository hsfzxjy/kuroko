package bluetooth

import (
	"kcore_android/jni"
	"kcore_android/magic"
	"time"
	"unsafe"

	ajni "github.com/hsfzxjy/android-jni-go"
)

type ioRequestType int

const (
	iortRead ioRequestType = iota
	iortWrite
	iortConnect
)

type ioRequest struct {
	typ  ioRequestType
	buf  []byte
	sock *BluetoothSocket
	resp chan<- ioResult
}

type ioResult struct {
	n   int
	err error
}

type ioWorker int

var iowReqCh chan ioRequest

func StartWorker(n int) {
	iowReqCh = make(chan ioRequest)
	for i := 0; i < n; i++ {
		go ioWorker(0).Loop()
	}
}

const _BUFFER_SIZE int = 2048

func (w ioWorker) Loop() {
	env := jni.AEnv.AttachGetJNIEnv()
	defer jni.AEnv.Detach()

	jbuffer := ajni.JNIEnvNewByteArray(env, int32(_BUFFER_SIZE))

LOOP:
	for {
		req := <-iowReqCh
		switch req.typ {
		case iortConnect:
			_, err := jni.CallExc[jni.Jvoid](
				env, req.sock.jsocket, "connect",
			)
			if err != nil {
				goto EXCEPT
			}
			err = req.sock.initStreams(env)
			if err != nil {
				goto EXCEPT
			}
			time.Sleep(100 * time.Millisecond)
			req.resp <- ioResult{}
			continue LOOP

		EXCEPT:
			req.resp <- ioResult{err: err}
			continue LOOP

		case iortRead:
			data, l := magic.UnpackSlice(req.buf)
			if l > _BUFFER_SIZE {
				l = _BUFFER_SIZE
			}
			n, err := jni.CallExc[jni.Jint](
				env, req.sock.jinput, "read",
				jbuffer, jni.Jint(0), jni.Jint(l),
			)
			if err != nil {
				req.resp <- ioResult{err: err}
				continue LOOP
			}
			ajni.JNIEnvGetByteArrayRegion(env, jbuffer, 0, int32(n), (*byte)(data))
			req.resp <- ioResult{n: int(n)}
		case iortWrite:
			data, l := magic.UnpackSlice(req.buf)
			var start int = 0
			for start < l {
				var length = l - start
				if length > _BUFFER_SIZE {
					length = _BUFFER_SIZE
				}
				ajni.JNIEnvSetByteArrayRegion(
					env, jbuffer,
					int32(start), int32(length),
					(*byte)(unsafe.Add(data, length)),
				)
				_, err := jni.CallExc[jni.Jvoid](
					env, req.sock.joutput, "write",
					jbuffer, jni.Jint(start), jni.Jint(length),
				)
				if err != nil {
					req.resp <- ioResult{err: err}
				}
				start += length
			}
			req.resp <- ioResult{n: l}
		}
	}
}
