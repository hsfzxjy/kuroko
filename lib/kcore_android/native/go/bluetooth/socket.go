package bluetooth

import (
	"context"
	"errors"
	"kcore_android/jni"
	"kmux"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	ajni "github.com/hsfzxjy/android-jni-go"
)

type BluetoothSocket struct {
	jsocket  *jni.Object
	jinput   *jni.Object
	joutput  *jni.Object
	tk       *kmux.TransportKey
	closedCh chan struct{}

	dieOnce     sync.Once
	connectOnce sync.Once

	ioCounter int32
}

var ErrBluetoothSocketClosed = errors.New("bluetooth socket was closed")
var ErrBluetoothSocketConnectionTimeout = errors.New("bluetooth socket costs too lang to connect")
var ErrBluetoothSocketConnectionCanceled = errors.New("bluetooth socket connection canceled by user")

func NewBluetoothSocket(
	ctx context.Context,
	env *ajni.JNIEnv,
	jsocket *jni.Object,
	ba *[6]byte,
) (*BluetoothSocket, error) {
	var err error
	ret := new(BluetoothSocket)
	ret.jsocket = jsocket

	if ba == nil {
		if err = ret.initStreams(env); err != nil {
			return nil, err
		}

		device := jni.CallObject(
			env, jsocket,
			"getRemoteDevice", JClassBluetoothDevice,
		)
		jaddr := jni.Call[jni.Jstring](env, device, "getAddress")
		addr := jni.ToGoString(env, jaddr)
		defer addr.Release()
		tmp := kmux.Str2Ba(addr.Str)
		ba = &tmp

		// socket was connected, consumes the Once
		ret.connectOnce.Do(func() {})
	}
	tk := new(kmux.TransportKey)
	tk[0] = byte(kmux.TT_BLUETOOTH)
	copy(tk[1:7], ba[:])
	ret.tk = tk

	ret.closedCh = make(chan struct{})
	go func() {
		<-ret.closedCh
		ret.clearResources()
	}()

	err = ret.ensureConnected(ctx)

	return ret, err
}

func (s *BluetoothSocket) IsClosed() bool {
	select {
	case <-s.closedCh:
		return true
	default:
		return false
	}
}

func (s *BluetoothSocket) ClosedCh() <-chan struct{}        { return s.closedCh }
func (s *BluetoothSocket) TransportKey() *kmux.TransportKey { return s.tk }

func (s *BluetoothSocket) clearResources() error {
	env := jni.AEnv.AttachGetJNIEnv()
	defer jni.AEnv.Detach()

	for atomic.LoadInt32(&s.ioCounter) != 0 {
		runtime.Gosched()
		// spinning wait for ongoing IO
	}

	if s.joutput != nil {
		defer s.joutput.ReleaseGlobal(env)
		jni.CallExc[jni.Jvoid](env, s.joutput, "flush")
		time.Sleep(100 * time.Millisecond)
		jni.CallExc[jni.Jvoid](env, s.joutput, "close")
	}

	if s.jinput != nil {
		defer s.jinput.ReleaseGlobal(env)
		jni.CallExc[jni.Jvoid](env, s.jinput, "close")
	}

	defer s.jsocket.ReleaseGlobal(env)
	jni.CallExc[jni.Jvoid](env, s.jsocket, "close")
	return nil
}

func (s *BluetoothSocket) Close() error {
	s.dieOnce.Do(func() { close(s.closedCh) })
	return nil
}

func (s *BluetoothSocket) initStreams(env *ajni.JNIEnv) error {
	if s.IsClosed() {
		return ErrBluetoothSocketClosed
	}
	var err error
	s.jinput, err = jni.CallObjectExcNonNull(env, s.jsocket, "getInputStream", JClassInputStream)
	if err != nil {
		return err
	}
	s.joutput, err = jni.CallObjectExcNonNull(env, s.jsocket, "getOutputStream", JClassOutputStream)
	if err != nil {
		return err
	}
	s.jinput.AsGlobal(env)
	s.joutput.AsGlobal(env)
	return nil
}

func (s *BluetoothSocket) ensureConnected(ctx context.Context) error {
	var err error
	s.connectOnce.Do(func() {
		respCh := make(chan ioResult)
		iowReqCh <- ioRequest{iortConnect, nil, s, respCh}
		select {
		case resp := <-respCh:
			err = resp.err
			close(respCh)
			if err != nil {
				s.Close()
			}
		case <-time.After(10 * time.Second):
			go func() { <-respCh; close(respCh) }()
			s.Close()
			err = ErrBluetoothSocketConnectionTimeout
		case <-ctx.Done():
			go func() { <-respCh; close(respCh) }()
			s.Close()
			err = ErrBluetoothSocketConnectionCanceled
		}
	})
	return err
}

func (s *BluetoothSocket) Read(buf []byte) (int, error) {
	atomic.AddInt32(&s.ioCounter, 1)
	defer atomic.AddInt32(&s.ioCounter, -1)

	if s.IsClosed() {
		return -1, ErrBluetoothSocketClosed
	}

	respCh := make(chan ioResult)
	defer close(respCh)

	req := ioRequest{iortRead, buf, s, respCh}
	iowReqCh <- req
	resp := <-respCh
	if resp.err != nil {
		s.Close()
	}

	return resp.n, resp.err
}

func (s *BluetoothSocket) Write(buf []byte) (int, error) {
	atomic.AddInt32(&s.ioCounter, 1)
	defer atomic.AddInt32(&s.ioCounter, -1)

	if s.IsClosed() {
		return -1, ErrBluetoothSocketClosed
	}

	respCh := make(chan ioResult)
	defer close(respCh)

	req := ioRequest{iortWrite, buf, s, respCh}
	iowReqCh <- req
	resp := <-respCh
	if resp.err != nil {
		s.Close()
	}

	return resp.n, resp.err
}
