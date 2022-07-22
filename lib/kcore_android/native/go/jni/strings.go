package jni

import (
	ajni "github.com/hsfzxjy/android-jni-go"
)

type BorrowedJString struct {
	env  *ajni.JNIEnv
	jstr Jstring
	Str  string
}

func ToGoString(env *ajni.JNIEnv, jstr Jstring) *BorrowedJString {
	ret := new(BorrowedJString)
	if env == nil {
		env = AEnv.GetJNIEnv()
	}
	ret.jstr = jstr
	ret.env = env

	var isCopy byte
	ret.Str = ajni.JNIEnvGetStringUTFChars(env, jstr, &isCopy)
	return ret
}

func (bjs *BorrowedJString) Release() {
	ajni.JNIEnvReleaseStringUTFChars(bjs.env, bjs.jstr, bjs.Str)
}
