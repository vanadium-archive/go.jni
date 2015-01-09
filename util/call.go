// +build android

package util

import (
	"fmt"
	"reflect"
	"unsafe"
)

// #cgo LDFLAGS: -ljniwrapper
// #include <stdlib.h>
// #include "jni_wrapper.h"
//
// static jvalue* allocJValueArray(int elements) {
//   return malloc(sizeof(jvalue) * elements);
// }
//
// static setJValueArrayElement(jvalue* arr, int index, jvalue val) {
//   arr[index] = val;
// }
//
import "C"

// jArg converts a Go argument to a Java argument.  It uses the provided sign to
// validate that the argument is of a compatible type.
func jArg(env *C.JNIEnv, v interface{}, sign Sign) (C.jvalue, bool) {
	var errValue C.jvalue
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.UnsafePointer {
		rv = reflect.ValueOf(rv.Pointer()) // Convert the pointer's address to a uintptr
	}
	var ptr unsafe.Pointer
	switch rv.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		if !isSignOneOf(sign, []Sign{ByteSign, ShortSign, IntSign, LongSign}) {
			return errValue, false
		}
		jv := C.jint(rv.Int())
		ptr = unsafe.Pointer(&jv)
	case reflect.Int64:
		if sign != LongSign {
			return errValue, false
		}
		jv := C.jlong(rv.Int())
		ptr = unsafe.Pointer(&jv)
	case reflect.Uint64:
		if sign != LongSign {
			return errValue, false
		}
		jv := C.jlong(rv.Uint())
		ptr = unsafe.Pointer(&jv)
	case reflect.Uintptr:
		if isSignOneOf(sign, []Sign{ByteSign, BoolSign, CharSign, ShortSign, IntSign, FloatSign, DoubleSign}) {
			return errValue, false
		}
		jv := C.jlong(rv.Uint())
		ptr = unsafe.Pointer(&jv)
	case reflect.String:
		if sign != StringSign {
			return errValue, false
		}
		// JString allocates the strings locally, so they are freed automatically when we return to Java.
		jv := JString(env, rv.String())
		if jv == nil {
			return errValue, false
		}
		ptr = unsafe.Pointer(&jv)
	case reflect.Slice, reflect.Array:
		switch rv.Type().Elem().Kind() {
		case reflect.Uint8:
			if sign != ArraySign(ByteSign) {
				return errValue, false
			}
			bs := rv.Interface().([]byte)
			jv := JByteArray(env, bs)
			ptr = unsafe.Pointer(&jv)
		case reflect.String:
			if sign != ArraySign(StringSign) {
				return errValue, false
			}
			// TODO(bprosnitz) We should handle objects by calling jArg recursively. We need a way to get the sign of the target type or treat it as an Object for non-string types.
			strs := rv.Interface().([]string)
			jv := JStringArray(env, strs)
			ptr = unsafe.Pointer(&jv)
		default:
			return errValue, false
		}
	default:
		return errValue, false
	}
	if ptr == nil {
		return errValue, false
	}
	return *(*C.jvalue)(ptr), true
}

// jArgArray converts a slice of Go args to an array of Java args.  It uses the provided slice of
// Signs to validate that the arguments are of compatible types.
func jArgArray(env *C.JNIEnv, args []interface{}, argSigns []Sign) (jArr *C.jvalue, free func(), err error) {
	if len(argSigns) != len(args) {
		return (*C.jvalue)(nil), nil, fmt.Errorf("mismatch in number of arguments, want %d, got %d", len(argSigns), len(args))
	}
	jvalueArr := C.allocJValueArray(C.int(len(args)))
	for i, arg := range args {
		sign := argSigns[i]
		jval, ok := jArg(env, arg, sign)
		if !ok {
			return (*C.jvalue)(nil), nil, fmt.Errorf("couldn't get Java value for argument #%d [%v] of expected type %v", i, arg, sign)
		}
		C.setJValueArrayElement(jvalueArr, C.int(i), jval)
	}
	freeFunc := func() {
		C.free(unsafe.Pointer(jvalueArr))
	}
	return jvalueArr, freeFunc, nil
}

// NewObject calls a java constructor through JNI, passing the specified args.
func NewObject(env interface{}, class interface{}, argSigns []Sign, args ...interface{}) (C.jobject, error) {
	if class == nil {
		panic("cannot call constructor of nil class")
	}
	jenv := getEnv(env)
	jclass := getClass(class)

	jcid, err := JMethodID(jenv, jclass, "<init>", FuncSign(argSigns, VoidSign))
	if err != nil {
		return nil, err
	}

	valArray, freeFunc, err := jArgArray(jenv, args, argSigns)
	if err != nil {
		return nil, err
	}
	defer freeFunc()

	ret := C.NewObjectA(jenv, jclass, jcid, valArray)
	err = JExceptionMsg(env)
	return ret, err
}

// setupMethodCall performs the shared preparation operations between various Java method invocation functions.
func setupMethodCall(env interface{}, object interface{}, name string, argSigns []Sign, retSign Sign, args []interface{}) (jenv *C.JNIEnv, jobject C.jobject, jmid C.jmethodID, jvalArray *C.jvalue, freeFunc func(), err error) {
	jenv = getEnv(env)
	jobject = getObject(object)
	jclass := C.GetObjectClass(jenv, jobject)

	jmid, err = JMethodID(jenv, jclass, name, FuncSign(argSigns, retSign))
	if err != nil {
		return
	}
	jvalArray, freeFunc, err = jArgArray(jenv, args, argSigns)
	return
}

// CallObjectMethod calls a Java method that returns a java object.
func CallObjectMethod(env interface{}, object interface{}, name string, argSigns []Sign, retSign Sign, args ...interface{}) (C.jobject, error) {
	switch retSign {
	case ByteSign, CharSign, ShortSign, LongSign, FloatSign, DoubleSign, BoolSign, IntSign, VoidSign:
		panic(fmt.Sprintf("Illegal call to CallObjectMethod on method with return sign %s", retSign))
	}
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, retSign, args)
	if err != nil {
		return nil, err
	}
	defer freeFunc()
	ret := C.CallObjectMethodA(jenv, jobject, jmid, valArray)
	return ret, JExceptionMsg(env)
}

// CallStringMethod calls a Java method that returns a string.
func CallStringMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (string, error) {
	jstr, err := CallObjectMethod(env, object, name, argSigns, StringSign, args...)
	return GoString(env, jstr), err
}

// CallByteArrayMethod calls a Java method that returns a byte array.
func CallByteArrayMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) ([]byte, error) {
	jArr, err := CallObjectMethod(env, object, name, argSigns, ArraySign(ByteSign), args...)
	if err != nil {
		return nil, err
	}
	return GoByteArray(env, jArr), nil
}

// CallObjectArrayMethod calls a Java method that returns an object array.
func CallObjectArrayMethod(env interface{}, object interface{}, name string, argSigns []Sign, retElemSign Sign, args ...interface{}) ([]C.jobject, error) {
	jenv := getEnv(env)
	jArr, err := CallObjectMethod(env, object, name, argSigns, ArraySign(retElemSign), args...)
	if err != nil {
		return nil, err
	}
	if jArr == nil {
		return nil, nil
	}
	ret := make([]C.jobject, int(C.GetArrayLength(jenv, C.jarray(jArr))))
	for i, _ := range ret {
		ret[i] = C.jobject(C.GetObjectArrayElement(jenv, C.jobjectArray(jArr), C.jsize(i)))
	}
	return ret, nil
}

// CallStringArrayMethod calls a Java method that returns an string array.
func CallStringArrayMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) ([]string, error) {
	objarr, err := CallObjectArrayMethod(env, object, name, argSigns, StringSign, args...)
	if err != nil {
		return nil, err
	}
	strs := make([]string, len(objarr))
	for i, obj := range objarr {
		strs[i] = GoString(env, obj)
	}
	return strs, nil
}

// CallMapMethod calls a Java method that returns a Map.
func CallMapMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (map[C.jobject]C.jobject, error) {
	jMap, err := CallObjectMethod(env, object, name, argSigns, MapSign, args...)
	if err != nil {
		return nil, err
	}
	if jMap == nil {
		return nil, nil
	}
	jEntrySet, err := CallObjectMethod(env, jMap, "entrySet", nil, SetSign)
	if err != nil {
		return nil, err
	}
	jIter, err := CallObjectMethod(env, jEntrySet, "iterator", nil, IteratorSign)
	if err != nil {
		return nil, err
	}
	ret := make(map[C.jobject]C.jobject)
	for {
		if hasNext, err := CallBooleanMethod(env, jIter, "hasNext", nil); err != nil {
			return nil, err
		} else if !hasNext {
			break
		}
		jEntry, err := CallObjectMethod(env, jIter, "next", nil, ObjectSign)
		if err != nil {
			return nil, err
		}
		jKey, err := CallObjectMethod(env, jEntry, "getKey", nil, ObjectSign)
		if err != nil {
			return nil, err
		}
		jValue, err := CallObjectMethod(env, jEntry, "getValue", nil, ObjectSign)
		if err != nil {
			return nil, err
		}
		ret[jKey] = jValue
	}
	return ret, nil
}

// CallBooleanMethod calls a Java method that returns a boolean.
func CallBooleanMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (bool, error) {
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, BoolSign, args)
	if err != nil {
		return false, err
	}
	defer freeFunc()
	ret := C.CallBooleanMethodA(jenv, jobject, jmid, valArray) != C.JNI_OK
	return ret, JExceptionMsg(env)
}

// CallIntMethod calls a Java method that returns an int.
func CallIntMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (int, error) {
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, IntSign, args)
	if err != nil {
		return 0, err
	}
	defer freeFunc()
	ret := int(C.CallIntMethodA(jenv, jobject, jmid, valArray))
	return ret, JExceptionMsg(env)
}

// CallLongMethod calls a Java method that returns an int64.
func CallLongMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (int64, error) {
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, LongSign, args)
	if err != nil {
		return 0, err
	}
	defer freeFunc()
	ret := int64(C.CallLongMethodA(jenv, jobject, jmid, valArray))
	return ret, JExceptionMsg(env)
}

// CallVoidMethod calls a Java method that "returns" void.
func CallVoidMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) error {
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, VoidSign, args)
	if err != nil {
		return err
	}
	C.CallVoidMethodA(jenv, jobject, jmid, valArray)
	freeFunc()
	return JExceptionMsg(env)
}

// CallStaticObjectMethod calls a static Java method that returns a Java object.
func CallStaticObjectMethod(env interface{}, class interface{}, name string, argSigns []Sign, retSign Sign, args ...interface{}) (C.jobject, error) {
	switch retSign {
	case ByteSign, CharSign, ShortSign, LongSign, FloatSign, DoubleSign, BoolSign, IntSign, VoidSign:
		panic(fmt.Sprintf("Illegal call to CallObjectMethod on method with return sign %s", retSign))
	}
	jenv := getEnv(env)
	jclass := getClass(class)

	jmid, err := JStaticMethodID(jenv, jclass, name, FuncSign(argSigns, retSign))
	if err != nil {
		return nil, err
	}
	jvalArray, freeFunc, err := jArgArray(jenv, args, argSigns)
	if err != nil {
		return nil, err
	}
	defer freeFunc()
	ret := C.CallStaticObjectMethodA(jenv, jclass, jmid, jvalArray)
	return ret, JExceptionMsg(env)
}

// CallStaticStringMethod calls a static Java method that returns a string.
func CallStaticStringMethod(env interface{}, class interface{}, name string, argSigns []Sign, args ...interface{}) (string, error) {
	jString, err := CallStaticObjectMethod(env, class, name, argSigns, StringSign, args...)
	return GoString(env, jString), err
}

// CallStaticByteArrayMethod calls a static Java method that returns a byte array.
func CallStaticByteArrayMethod(env interface{}, class interface{}, name string, argSigns []Sign, args ...interface{}) ([]byte, error) {
	jArr, err := CallStaticObjectMethod(env, class, name, argSigns, ArraySign(ByteSign), args...)
	if err != nil {
		return nil, err
	}
	return GoByteArray(env, jArr), nil
}

// CallStaticIntMethod calls a static Java method that returns an int.
func CallStaticIntMethod(env interface{}, class interface{}, name string, argSigns []Sign, args ...interface{}) (int, error) {
	jenv := getEnv(env)
	jclass := getClass(class)

	jmid, err := JStaticMethodID(jenv, jclass, name, FuncSign(argSigns, IntSign))
	if err != nil {
		return 0, err
	}
	jvalArray, freeFunc, err := jArgArray(jenv, args, argSigns)
	if err != nil {
		return -1, err
	}
	defer freeFunc()
	ret := int(C.CallStaticIntMethodA(jenv, jclass, jmid, jvalArray))
	return ret, JExceptionMsg(env)
}

func handleError(name string, err error) {
	if err != nil {
		panic(fmt.Sprintf("error while calling jni method %q: %v", name, err))
	}
}

func isSignOneOf(sign Sign, set []Sign) bool {
	for _, s := range set {
		if sign == s {
			return true
		}
	}
	return false
}
