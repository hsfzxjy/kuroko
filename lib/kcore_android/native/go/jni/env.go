package jni

import (
	"runtime"
	"sync/atomic"

	ajni "github.com/hsfzxjy/android-jni-go"
)

type AndroidEnv struct {
	VM            *ajni.JavaVM
	JNIEnvVersion int32

	Activity         *Object
	BluetoothManager *Object
	BluetoothAdapter *Object

	inited int32
}

func newAndroidEnv() *AndroidEnv {
	return &AndroidEnv{
		inited: 0,
	}
}

func (a *AndroidEnv) Init(env *ajni.JNIEnv, act, btman, bta ajni.Jobject) {
	if !atomic.CompareAndSwapInt32(&a.inited, 0, 1) {
		return
	}
	var vm *ajni.JavaVM
	ajni.JNIEnvGetJavaVM(env, &vm)
	a.VM = vm

	a.JNIEnvVersion = int32(ajni.JNIEnvGetVersion(env))

	a.Activity = Wrap(ajni.JNIEnvNewGlobalRef(env, act), nil, 0)
	a.BluetoothManager = Wrap(ajni.JNIEnvNewGlobalRef(env, btman), nil, 0)
	a.BluetoothAdapter = Wrap(ajni.JNIEnvNewGlobalRef(env, bta), nil, 0)

	atomic.StoreInt32(&a.inited, 2)
}

func (a *AndroidEnv) ensureInited() {
	for atomic.LoadInt32(&a.inited) != 2 {
		// since the condition is less likely to fail. we spin here
		runtime.Gosched()
	}
}

func (a *AndroidEnv) GetJNIEnv() *ajni.JNIEnv {
	a.ensureInited()

	var env *ajni.JNIEnv
	ajni.JNIGetEnv(a.VM, &env, a.JNIEnvVersion)
	return env
}

func (a *AndroidEnv) AttachGetJNIEnv() *ajni.JNIEnv {
	a.ensureInited()

	var env *ajni.JNIEnv
	runtime.LockOSThread()
	ajni.JNIAttachCurrentThread(a.VM, &env, nil)
	return env
}

func (a *AndroidEnv) Detach() {
	a.ensureInited()

	ajni.JNIDetachCurrentThread(a.VM)
	runtime.UnlockOSThread()
}

var AEnv = newAndroidEnv()
