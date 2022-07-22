package jni

import ajni "github.com/hsfzxjy/android-jni-go"

type (
	Jboolean      = ajni.Jboolean
	Jbyte         = ajni.Jbyte
	Jchar         = ajni.Jchar
	Jshort        = ajni.Jshort
	Jint          = ajni.Jint
	Jlong         = ajni.Jlong
	Jfloat        = ajni.Jfloat
	Jdouble       = ajni.Jdouble
	Jstring       = ajni.Jstring
	JobjectArray  = ajni.JobjectArray
	JbooleanArray = ajni.JbooleanArray
	JbyteArray    = ajni.JbyteArray
	JcharArray    = ajni.JcharArray
	JshortArray   = ajni.JshortArray
	JintArray     = ajni.JintArray
	JlongArray    = ajni.JlongArray
	JfloatArray   = ajni.JfloatArray
	JdoubleArray  = ajni.JdoubleArray
	Jvoid         struct{}
)

type SimpleReturnType interface {
	Jvoid |
		Jboolean | Jbyte | Jchar | Jint | Jlong | Jshort | Jfloat | Jdouble |
		Jstring |
		JobjectArray | JbooleanArray | JbyteArray | JcharArray | JshortArray | JintArray | JlongArray | JfloatArray | JdoubleArray
}
