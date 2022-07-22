package bluetooth

import (
	"context"
	"kcore_android/jni"
	"kmux"
)

func newBluetoothAccepter() (kmux.Listener, error) {
	env := jni.AEnv.AttachGetJNIEnv()
	defer jni.AEnv.Detach()

	adapter := jni.AEnv.BluetoothAdapter
	serviceUUID := getServiceUUID(env)
	jsocket, err := jni.CallObjectExc(
		env, adapter, "listenUsingInsecureRfcommWithServiceRecord", JClassBluetoothServerSocket,
		"Kuroko", serviceUUID,
	)
	if err != nil {
		return nil, err
	}
	socket, err := NewBluetoothServerSocket(env, jsocket)
	if err != nil {
		return nil, err
	}
	jsocket.AsGlobal(env)
	return socket, nil
}

func bluetoothDialer(ctx context.Context, ba [6]byte) (kmux.RawTransport, error) {
	env := jni.AEnv.AttachGetJNIEnv()
	defer jni.AEnv.Detach()

	addr := kmux.Ba2Str(ba)
	device, err := jni.CallObjectExc(
		env, jni.AEnv.BluetoothAdapter, "getRemoteDevice", JClassBluetoothDevice,
		addr,
	)

	if err != nil {
		return nil, err
	}
	socket, err := jni.CallObjectExc(
		env, device, "createInsecureRfcommSocketToServiceRecord",
		JClassBluetoothSocket, getServiceUUID(env),
	)
	if err != nil {
		return nil, err
	}
	socket.AsGlobal(env)

	var rtr kmux.RawTransport
	for retry := 1; retry <= 5; retry++ {
		rtr, err = NewBluetoothSocket(ctx, env, socket, &ba)
		if err == nil {
			return rtr, nil
		}
		select {
		case <-ctx.Done():
			return nil, ErrBluetoothSocketConnectionCanceled
		default:
		}
	}
	return rtr, err
}

func Init() {
	kmux.AccepterManager.Register(kmux.TT_BLUETOOTH, newBluetoothAccepter)
	kmux.SetBluetoothDialer(bluetoothDialer)
	StartWorker(2)
}
