package jni

import (
	"unsafe"

	ajni "github.com/hsfzxjy/android-jni-go"
)

type SigBuilder interface {
	WriteRune(rune) (int, error)
	WriteString(string) (int, error)
}

type dummyBuilder struct{}

func (dummyBuilder) WriteRune(rune) (int, error)     { return 0, nil }
func (dummyBuilder) WriteString(string) (int, error) { return 0, nil }

func asJarg(
	val any,
	sigBuilder SigBuilder,
	env *ajni.JNIEnv,
	finalizers *[]func(),
) (ret ajni.Jvalue) {
	var value any = val
	switch val := val.(type) {
	case bool:
		sigBuilder.WriteRune('Z')
		return ajni.JbooleanV(val)
	case Jboolean:
		sigBuilder.WriteRune('Z')
	case byte:
		sigBuilder.WriteRune('B')
	case Jbyte:
		sigBuilder.WriteRune('B')
	case uint16:
		sigBuilder.WriteRune('C')
	case Jchar:
		sigBuilder.WriteRune('C')
	case int16:
		sigBuilder.WriteRune('S')
	case Jshort:
		sigBuilder.WriteRune('S')
	case int32:
		sigBuilder.WriteRune('I')
	case Jint:
		sigBuilder.WriteRune('I')
	case float32:
		sigBuilder.WriteRune('F')
	case Jfloat:
		sigBuilder.WriteRune('F')
	case float64:
		sigBuilder.WriteRune('D')
	case Jdouble:
		sigBuilder.WriteRune('D')
	case int64:
		sigBuilder.WriteRune('J')
	case Jlong:
		sigBuilder.WriteRune('J')
	case Jvoid:
		sigBuilder.WriteRune('V')
		return ajni.Jvalue{}
	case ajni.JbooleanArray:
		sigBuilder.WriteString("[Z")
	case ajni.JbyteArray:
		sigBuilder.WriteString("[B")
	case ajni.JcharArray:
		sigBuilder.WriteString("[C")
	case ajni.JshortArray:
		sigBuilder.WriteString("[S")
	case ajni.JintArray:
		sigBuilder.WriteString("[I")
	case ajni.JfloatArray:
		sigBuilder.WriteString("[F")
	case ajni.JdoubleArray:
		sigBuilder.WriteString("[D")
	case ajni.JlongArray:
		sigBuilder.WriteString("[J")
	case string:
		sigBuilder.WriteString("Ljava/lang/String;")
		var str = val
		if str[len(str)-1] != '\x00' {
			str = str + "\x00"
		}
		jstr := ajni.JNIEnvNewStringUTF(env, str)
		value = jstr
		*finalizers = append(*finalizers, func() {
			ajni.JNIEnvDeleteLocalRef(env, transmute[ajni.Jobject](jstr))
		})
	case ajni.Jstring:
		sigBuilder.WriteString("Ljava/lang/String;")
	case *Object:
		obj := val
		value = obj.ptr

		sigBuilder.WriteRune('L')
		sigBuilder.WriteString(obj.classname.Name())
		sigBuilder.WriteRune(';')
	default:
		panic("unsupported type")
	}

	// NOTE: very unsafe!! This requires Go interface ABI to be stable
	var datap = (*struct {
		_    unsafe.Pointer
		data unsafe.Pointer
	})(unsafe.Pointer(&value)).data
	if datap != nil {
		ret = *(*ajni.Jvalue)(datap)
	}
	return
}

func transmute[R any, S any](val S) R {
	return *(*R)(unsafe.Pointer(&val))
}