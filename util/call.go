// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package util

import (
	"fmt"
	"unsafe"
)

// #include <stdlib.h>
// #include "jni_wrapper.h"
//
// static jvalue* allocJValueArray(int elements) {
//   return malloc(sizeof(jvalue) * elements);
// }
//
// static void setJValueArrayElement(jvalue* arr, int index, jvalue val) {
//   arr[index] = val;
// }
//
import "C"

// jArgArray converts a slice of Go args to an array of Java args.  It uses the provided slice of
// Signs to validate that the arguments are of compatible types.
func jArgArray(env *C.JNIEnv, args []interface{}, argSigns []Sign) (jArr *C.jvalue, free func(), err error) {
	if len(argSigns) != len(args) {
		return (*C.jvalue)(nil), nil, fmt.Errorf("mismatch in number of arguments, want %d, got %d", len(argSigns), len(args))
	}
	jvalueArr := C.allocJValueArray(C.int(len(args)))
	for i, arg := range args {
		sign := argSigns[i]
		jVal, ok := jValue(env, arg, sign)
		if !ok {
			return (*C.jvalue)(nil), nil, fmt.Errorf("couldn't get Java value for argument #%d [%v] of expected type %v", i, arg, sign)
		}
		C.setJValueArrayElement(jvalueArr, C.int(i), jVal)
	}
	freeFunc := func() {
		C.free(unsafe.Pointer(jvalueArr))
	}
	return jvalueArr, freeFunc, nil
}

// NewObject calls a java constructor through JNI, passing the specified args.
func NewObject(env interface{}, class interface{}, argSigns []Sign, args ...interface{}) (unsafe.Pointer, error) {
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

	ret := C.NewObjectA(jenv, jclass, C.jmethodID(jcid), valArray)
	err = JExceptionMsg(env)
	return unsafe.Pointer(ret), err
}

// setupMethodCall performs the shared preparation operations between various
// Java method invocation functions.
func setupMethodCall(env interface{}, object interface{}, name string, argSigns []Sign, retSign Sign, args ...interface{}) (jenv *C.JNIEnv, jobject C.jobject, jmid C.jmethodID, jvalArray *C.jvalue, freeFunc func(), err error) {
	jenv = getEnv(env)
	jobject = getObject(object)
	jclass := C.GetObjectClass(jenv, jobject)

	var id unsafe.Pointer
	id, err = JMethodID(jenv, jclass, name, FuncSign(argSigns, retSign))
	if err != nil {
		return
	}
	jmid = C.jmethodID(id)
	jvalArray, freeFunc, err = jArgArray(jenv, args, argSigns)
	if err != nil {
		err = fmt.Errorf("error creating arguments for method %s: %v", name, err)
	}
	return

}

// setupStaticMethodCall performs the shared preparation operations between
// various Java static method invocation functions.
func setupStaticMethodCall(env interface{}, class interface{}, name string, argSigns []Sign, retSign Sign, args ...interface{}) (jenv *C.JNIEnv, jclass C.jclass, jmid C.jmethodID, jvalArray *C.jvalue, freeFunc func(), err error) {
	jenv = getEnv(env)
	jclass = getClass(class)

	var id unsafe.Pointer
	id, err = JStaticMethodID(env, jclass, name, FuncSign(argSigns, retSign))
	if err != nil {
		return
	}
	jmid = C.jmethodID(id)
	jvalArray, freeFunc, err = jArgArray(jenv, args, argSigns)
	if err != nil {
		err = fmt.Errorf("error creating arguments for method %s: %v", name, err)
	}
	return
}

// CallObjectMethod calls a Java method that returns a java object.
func CallObjectMethod(env interface{}, object interface{}, name string, argSigns []Sign, retSign Sign, args ...interface{}) (unsafe.Pointer, error) {
	switch retSign {
	case ByteSign, CharSign, ShortSign, LongSign, FloatSign, DoubleSign, BoolSign, IntSign, VoidSign:
		panic(fmt.Sprintf("Illegal call to CallObjectMethod on method with return sign %s", retSign))
	}
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, retSign, args...)
	if err != nil {
		return nil, err
	}
	defer freeFunc()
	ret := C.CallObjectMethodA(jenv, jobject, jmid, valArray)
	return unsafe.Pointer(ret), JExceptionMsg(env)
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
func CallObjectArrayMethod(env interface{}, object interface{}, name string, argSigns []Sign, retElemSign Sign, args ...interface{}) ([]unsafe.Pointer, error) {
	jenv := getEnv(env)
	jArr, err := CallObjectMethod(env, object, name, argSigns, ArraySign(retElemSign), args...)
	if err != nil {
		return nil, err
	}
	if jArr == nil {
		return nil, nil
	}
	ret := make([]unsafe.Pointer, int(C.GetArrayLength(jenv, C.jarray(jArr))))
	for i, _ := range ret {
		ret[i] = unsafe.Pointer(C.GetObjectArrayElement(jenv, C.jobjectArray(jArr), C.jsize(i)))
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
func CallMapMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (map[unsafe.Pointer]unsafe.Pointer, error) {
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
	ret := make(map[unsafe.Pointer]unsafe.Pointer)
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
		jVal, err := CallObjectMethod(env, jEntry, "getValue", nil, ObjectSign)
		if err != nil {
			return nil, err
		}
		ret[jKey] = jVal
	}
	return ret, nil
}

// CallMultimapMethod calls a Java method that returns a Multimap.
func CallMultimapMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (map[unsafe.Pointer][]unsafe.Pointer, error) {
	jMultimap, err := CallObjectMethod(env, object, name, argSigns, MultimapSign, args...)
	if err != nil {
		return nil, err
	}
	if jMultimap == nil {
		return nil, nil
	}
	jEntrySet, err := CallObjectMethod(env, jMultimap, "entrySet", nil, SetSign)
	if err != nil {
		return nil, err
	}
	jIter, err := CallObjectMethod(env, jEntrySet, "iterator", nil, IteratorSign)
	if err != nil {
		return nil, err
	}
	ret := make(map[unsafe.Pointer][]unsafe.Pointer)
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
		jVal, err := CallObjectMethod(env, jEntry, "getValue", nil, ObjectSign)
		if err != nil {
			return nil, err
		}
		ret[jKey] = append(ret[jKey], jVal)
	}
	return ret, nil
}

// CallBooleanMethod calls a Java method that returns a boolean.
func CallBooleanMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (bool, error) {
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, BoolSign, args...)
	if err != nil {
		return false, err
	}
	defer freeFunc()
	ret := C.CallBooleanMethodA(jenv, jobject, jmid, valArray) != C.JNI_OK
	return ret, JExceptionMsg(env)
}

// CallIntMethod calls a Java method that returns an int.
func CallIntMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (int, error) {
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, IntSign, args...)
	if err != nil {
		return 0, err
	}
	defer freeFunc()
	ret := int(C.CallIntMethodA(jenv, jobject, jmid, valArray))
	return ret, JExceptionMsg(env)
}

// CallLongMethod calls a Java method that returns an int64.
func CallLongMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) (int64, error) {
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, LongSign, args...)
	if err != nil {
		return 0, err
	}
	defer freeFunc()
	ret := int64(C.CallLongMethodA(jenv, jobject, jmid, valArray))
	return ret, JExceptionMsg(env)
}

// CallVoidMethod calls a Java method that doesn't return anything.
func CallVoidMethod(env interface{}, object interface{}, name string, argSigns []Sign, args ...interface{}) error {
	jenv, jobject, jmid, valArray, freeFunc, err := setupMethodCall(env, object, name, argSigns, VoidSign, args...)
	if err != nil {
		return err
	}
	defer freeFunc()
	C.CallVoidMethodA(jenv, jobject, jmid, valArray)
	return JExceptionMsg(env)
}

// CallStaticObjectMethod calls a static Java method that returns a Java object.
func CallStaticObjectMethod(env interface{}, class interface{}, name string, argSigns []Sign, retSign Sign, args ...interface{}) (unsafe.Pointer, error) {
	switch retSign {
	case ByteSign, CharSign, ShortSign, LongSign, FloatSign, DoubleSign, BoolSign, IntSign, VoidSign:
		panic(fmt.Sprintf("Illegal call to CallStaticObjectMethod on method with return sign %s", retSign))
	}
	jenv, jclass, jmid, jvalArray, freeFunc, err := setupStaticMethodCall(env, class, name, argSigns, retSign, args...)
	if err != nil {
		return nil, err
	}
	defer freeFunc()
	ret := C.CallStaticObjectMethodA(jenv, jclass, jmid, jvalArray)
	return unsafe.Pointer(ret), JExceptionMsg(env)
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

// CallStaticLongArrayMethod calls a static Java method that returns a array of long.
func CallStaticLongArrayMethod(env interface{}, class interface{}, name string, argSigns []Sign, args ...interface{}) ([]int64, error) {
	jArr, err := CallStaticObjectMethod(env, class, name, argSigns, ArraySign(LongSign), args...)
	if err != nil {
		return nil, err
	}
	return GoLongArray(env, jArr), nil
}

// CallStaticIntMethod calls a static Java method that returns an int.
func CallStaticIntMethod(env interface{}, class interface{}, name string, argSigns []Sign, args ...interface{}) (int, error) {
	jenv, jclass, jmid, jvalArray, freeFunc, err := setupStaticMethodCall(env, class, name, argSigns, IntSign, args...)
	if err != nil {
		return 0, err
	}
	defer freeFunc()
	ret := int(C.CallStaticIntMethodA(jenv, jclass, jmid, jvalArray))
	return ret, JExceptionMsg(env)
}

// CallStaticVoidMethod calls a static Java method doesn't return anything.
func CallStaticVoidMethod(env interface{}, class interface{}, name string, argSigns []Sign, args ...interface{}) error {
	jenv, jclass, jmid, jvalArray, freeFunc, err := setupStaticMethodCall(env, class, name, argSigns, VoidSign, args...)
	if err != nil {
		return err
	}
	defer freeFunc()
	C.CallStaticVoidMethodA(jenv, jclass, jmid, jvalArray)
	return JExceptionMsg(env)
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
