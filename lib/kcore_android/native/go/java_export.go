package main

import "C"

import (
	"kcore_android/bluetooth"
	"kcore_android/dffi"
	"kcore_android/jni"
	"kcore_android/mux"
	"kmux"
	"sync"
	"unsafe"

	ajni "github.com/hsfzxjy/android-jni-go"
)

var initDLLOnce sync.Once

//export Java_site_hsfzxjy_kcore_KcoreAndroidPlugin_initDLL
func Java_site_hsfzxjy_kcore_KcoreAndroidPlugin_initDLL(envp, self, act, btman, bta unsafe.Pointer) int32 {

	dffi.IncDartVersion()

	initDLLOnce.Do(func() {
		jni.AEnv.Init(
			(*ajni.JNIEnv)(envp),
			ajni.Jobject(act),
			ajni.Jobject(btman),
			ajni.Jobject(bta),
		)
		bluetooth.Init()
		mux.Init()
		go kmux.AccepterManager.Loop()
	})

	return 0
}
