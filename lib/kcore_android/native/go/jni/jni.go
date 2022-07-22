package jni

import (
	"fmt"
	"strings"
	"sync"

	ajni "github.com/hsfzxjy/android-jni-go"
)

type hasName interface {
	Name() string
}

type ClassName string

func (cn ClassName) Name() string { return string(cn) }

type Object struct {
	ptr       ajni.Jobject
	classname hasName

	class   ajni.Jclass
	methods map[string]map[string]ajni.JmethodID

	lock sync.Mutex
}

func (o *Object) Ptr() uintptr { return uintptr(o.ptr) }
func (o *Object) Class() uintptr {
	if o.class == ajni.Jclass(0) {
		return 0
	} else {
		return uintptr(o.class)
	}
}

func Wrap(ptr ajni.Jobject, classname hasName, class ajni.Jclass) *Object {
	return &Object{
		ptr:       ptr,
		classname: classname,
		class:     class,
	}
}

func (o *Object) AsGlobal(env *ajni.JNIEnv) {
	o.lock.Lock()
	defer o.lock.Unlock()

	oldptr := o.ptr
	o.ptr = ajni.JNIEnvNewGlobalRef(env, o.ptr)
	ajni.JNIEnvDeleteLocalRef(env, oldptr)
}

func (o *Object) ReleaseGlobal(env *ajni.JNIEnv) {
	o.lock.Lock()
	defer o.lock.Unlock()

	ajni.JNIEnvDeleteGlobalRef(env, o.ptr)
	if o.class != 0 {
		ajni.JNIEnvDeleteGlobalRef(env, ajni.Jobject(o.class))
	}
}

func prepareCall[R any](
	env *ajni.JNIEnv,
	obj *Object,
	method string,
	dummyRet R,
	args []any,
) ([]ajni.Jvalue, ajni.JmethodID, []func()) {
	if obj.methods == nil {
		obj.methods = make(map[string]map[string]ajni.JmethodID)
	}

	var methodID ajni.JmethodID
	var overloadMap map[string]ajni.JmethodID
	overloadMap, methodHit := obj.methods[method]

	var jargs = make([]ajni.Jvalue, len(args))
	var finalizers []func()

	if methodHit && len(overloadMap) == 1 {
		for _, value := range overloadMap {
			methodID = value
		}
		for i, arg := range args {
			jargs[i] = asJarg(arg, dummyBuilder{}, env, &finalizers)
		}
	} else {
		var builder = strings.Builder{}

		builder.WriteRune('(')
		for i, arg := range args {
			jargs[i] = asJarg(arg, &builder, env, &finalizers)
		}
		builder.WriteRune(')')
		asJarg(dummyRet, &builder, env, &finalizers)
		builder.WriteRune('\x00')

		var funcSig = builder.String()

		if obj.class == 0 {
			obj.class = ajni.JNIEnvGetObjectClass(env, obj.ptr)
			obj.class = ajni.Jclass(ajni.JNIEnvNewGlobalRef(env, (ajni.Jobject)(obj.class)))
		}

		var overloadHit bool = false
		if methodHit {
			methodID, overloadHit = overloadMap[funcSig]
		}

		if !overloadHit {
			methodID = ajni.JNIEnvGetMethodID(env, obj.class, method+"\x00", funcSig)
			if !methodHit {
				obj.methods[method] = make(map[string]ajni.JmethodID)
			}
			obj.methods[method][funcSig] = methodID
		}
	}
	return jargs, methodID, finalizers
}

func CallObjectExcNonNull(
	env *ajni.JNIEnv,
	obj *Object,
	method string,
	clsname hasName,
	args ...any,
) (*Object, error) {
	ret, err := CallObjectExc(env, obj, method, clsname, args...)
	if err != nil && ret.ptr == 0 {
		return nil, fmt.Errorf(
			"jni call returns nil [clsname=%s, method=%s]",
			obj.classname.Name(), method,
		)
	}
	return ret, err
}

func CallObjectExc(
	env *ajni.JNIEnv,
	obj *Object,
	method string,
	clsname hasName,
	args ...any,
) (*Object, error) {
	ret := CallObject(env, obj, method, clsname, args...)
	if ajni.JNIEnvExceptionCheck(env) == 1 {
		ajni.JNIEnvExceptionDescribe(env)
		ajni.JNIEnvExceptionClear(env)
		return nil, fmt.Errorf(
			"java exception caught [clsname=%s, method=%s]",
			obj.classname.Name(), method,
		)
	}
	return ret, nil
}

func CallObject(
	env *ajni.JNIEnv,
	obj *Object,
	method string,
	clsname hasName,
	args ...any,
) *Object {
	ret := new(Object)
	ret.classname = clsname

	if env == nil {
		env = AEnv.GetJNIEnv()
	}
	jargs, methodID, finalizers := prepareCall(env, obj, method, ret, args)
	for _, finalizer := range finalizers {
		defer finalizer()
	}

	ret.ptr = ajni.JNIEnvCallObjectMethod(env, obj.ptr, methodID, jargs)
	return ret
}

func CallExc[R SimpleReturnType](env *ajni.JNIEnv, obj *Object, method string, args ...any) (R, error) {
	ret := Call[R](env, obj, method, args...)
	if ajni.JNIEnvExceptionCheck(env) == 1 {
		ajni.JNIEnvExceptionDescribe(env)
		ajni.JNIEnvExceptionClear(env)
		return ret, fmt.Errorf("java exception caught [clsname=%s, method=%s]", obj.classname.Name(), method)
	}
	return ret, nil
}

func Call[R SimpleReturnType](env *ajni.JNIEnv, obj *Object, method string, args ...any) R {
	obj.lock.Lock()
	defer obj.lock.Unlock()

	var d R
	var dummy any = d

	if env == nil {
		env = AEnv.GetJNIEnv()
	}
	jargs, methodID, finalizers := prepareCall(env, obj, method, d, args)
	for _, finalizer := range finalizers {
		defer finalizer()
	}

	switch dummy.(type) {
	case Jboolean:
		return transmute[R](ajni.JNIEnvCallBooleanMethod(env, obj.ptr, methodID, jargs))
	case Jbyte:
		return transmute[R](ajni.JNIEnvCallByteMethod(env, obj.ptr, methodID, jargs))
	case Jchar:
		return transmute[R](ajni.JNIEnvCallCharMethod(env, obj.ptr, methodID, jargs))
	case Jshort:
		return transmute[R](ajni.JNIEnvCallShortMethod(env, obj.ptr, methodID, jargs))
	case Jint:
		return transmute[R](ajni.JNIEnvCallIntMethod(env, obj.ptr, methodID, jargs))
	case Jlong:
		return transmute[R](ajni.JNIEnvCallLongMethod(env, obj.ptr, methodID, jargs))
	case Jfloat:
		return transmute[R](ajni.JNIEnvCallFloatMethod(env, obj.ptr, methodID, jargs))
	case Jdouble:
		return transmute[R](ajni.JNIEnvCallDoubleMethod(env, obj.ptr, methodID, jargs))
	case Jvoid:
		ajni.JNIEnvCallVoidMethod(env, obj.ptr, methodID, jargs)
		return d
	default:
		x := ajni.JNIEnvCallObjectMethod(env, obj.ptr, methodID, jargs)
		ret := transmute[R](x)
		return ret
	}
}
