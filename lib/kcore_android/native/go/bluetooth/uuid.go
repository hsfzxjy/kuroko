package bluetooth

import (
	"kcore_android/jni"
	"sync"

	ajni "github.com/hsfzxjy/android-jni-go"
)

var serviceUUID *jni.Object
var serviceUUIDOnce sync.Once

const uuid = "94F45378-7D6D-437D-973B-FBA39E49D4EE\x00"

func getServiceUUID(env *ajni.JNIEnv) *jni.Object {
	serviceUUIDOnce.Do(func() {
		juuidString := ajni.JNIEnvNewStringUTF(env, uuid)
		JClassUUID := ajni.JNIEnvFindClass(env, "java/util/UUID\x00")
		JClassUUID = ajni.Jclass(ajni.JNIEnvNewGlobalRef(env, ajni.Jobject(JClassUUID)))
		JfromString := ajni.JNIEnvGetStaticMethodID(
			env, JClassUUID, "fromString\x00",
			"(Ljava/lang/String;)Ljava/util/UUID;\x00",
		)
		jserviceUUID := ajni.JNIEnvCallStaticObjectMethod(
			env, JClassUUID, JfromString,
			[]ajni.Jvalue{ajni.JobjectV(ajni.Jobject(juuidString))},
		)
		serviceUUID = jni.Wrap(
			jserviceUUID,
			jni.ClassName("java/util/UUID"),
			JClassUUID,
		)
		serviceUUID.AsGlobal(env)
	})
	return serviceUUID
}
