package bluetooth

import (
	"context"
	"errors"
	"kcore_android/jni"
	"kmux"
	"sync"

	ajni "github.com/hsfzxjy/android-jni-go"
)

type acceptResult struct {
	sock *BluetoothSocket
	err  error
}

type BluetoothServerSocket struct {
	jsocket  *jni.Object
	dieOnce  sync.Once
	acceptCh chan acceptResult
}

func NewBluetoothServerSocket(env *ajni.JNIEnv, jsocket *jni.Object) (*BluetoothServerSocket, error) {
	ret := new(BluetoothServerSocket)
	ret.jsocket = jsocket
	ret.acceptCh = make(chan acceptResult)

	go ret.loop()
	return ret, nil
}

func (s *BluetoothServerSocket) Accept() (kmux.RawTransport, error) {
	res, ok := <-s.acceptCh
	if !ok {
		return nil, errors.New("Accept(): server socket was closed")
	}
	if res.err != nil {
		return nil, res.err
	}
	return res.sock, nil
}

func (s *BluetoothServerSocket) Close() error {
	s.dieOnce.Do(func() {
		env := jni.AEnv.AttachGetJNIEnv()
		defer jni.AEnv.Detach()

		defer s.jsocket.ReleaseGlobal(env)
		jni.CallExc[jni.Jvoid](env, s.jsocket, "close")
	})
	return nil
}

func (s *BluetoothServerSocket) loop() {
	env := jni.AEnv.AttachGetJNIEnv()
	defer jni.AEnv.Detach()

	var (
		err   error
		conn  *BluetoothSocket
		jconn *jni.Object
	)

	for {
		jconn, err = jni.CallObjectExc(
			env, s.jsocket, "accept", JClassBluetoothSocket,
		)
		if err != nil {
			goto EXCEPT
		}
		conn, err = NewBluetoothSocket(context.Background(), env, jconn, nil)
		if err != nil {
			goto EXCEPT
		}
		jconn.AsGlobal(env)
		s.acceptCh <- acceptResult{conn, nil}
		continue
	EXCEPT:
		s.acceptCh <- acceptResult{nil, err}
		close(s.acceptCh)
		return
	}
}
