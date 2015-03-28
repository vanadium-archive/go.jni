// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

// Package util provides various JNI utilities shared across our JNI code.
package util

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"v.io/v23/vdl"
	"v.io/v23/vom"
)

// #include <stdlib.h>
// #include "jni_wrapper.h"
// static jstring CallGetExceptionMessage(JNIEnv* env, jobject obj, jmethodID id) {
//   return (jstring)(*env)->CallObjectMethod(env, obj, id);
// }
import "C"

var (
	// Global reference for io.v.v23.vom.VomUtil class.
	jVomUtilClass C.jclass
	// Global reference for io.v.v23.verror.VException class.
	jVExceptionClass C.jclass
	// Global reference for io.v.v23.verror.VException$ActionCode class.
	jActionCodeClass C.jclass
	// Global reference for io.v.v23.verror.VException$IDAction class.
	jIDActionClass C.jclass
	// Global reference for io.v.v23.vdl.VdlValue class.
	jVdlValueClass C.jclass
	// Global reference for org.joda.time.DateTime class.
	jDateTimeClass C.jclass
	// Global reference for org.joda.time.Duration class.
	jDurationClass C.jclass
	// Global reference for java.lang.Throwable class.
	jThrowableClass C.jclass
	// Global reference for java.lang.System class.
	jSystemClass C.jclass
	// Global reference for java.lang.Object class.
	jObjectClass C.jclass
	// Global reference for java.lang.String class.
	jStringClass C.jclass
	// Global reference for java.util.HashMap class.
	jHashMapClass C.jclass
	// Global reference for []byte class.
	jByteArrayClass C.jclass
	// Cached Java VM.
	jVM *C.JavaVM
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func Init(jEnv interface{}) error {
	env := getEnv(jEnv)
	class, err := JFindClass(env, "io/v/v23/vom/VomUtil")
	if err != nil {
		return err
	}
	jVomUtilClass = C.jclass(class)
	class, err = JFindClass(env, "io/v/v23/verror/VException")
	if err != nil {
		return err
	}
	jVExceptionClass = C.jclass(class)
	class, err = JFindClass(env, "io/v/v23/verror/VException$ActionCode")
	if err != nil {
		return err
	}
	jActionCodeClass = C.jclass(class)
	class, err = JFindClass(env, "io/v/v23/verror/VException$IDAction")
	if err != nil {
		return err
	}
	jIDActionClass = C.jclass(class)
	class, err = JFindClass(env, "io/v/v23/vdl/VdlValue")
	if err != nil {
		return err
	}
	jVdlValueClass = C.jclass(class)
	class, err = JFindClass(env, "org/joda/time/DateTime")
	if err != nil {
		return err
	}
	jDateTimeClass = C.jclass(class)
	class, err = JFindClass(env, "org/joda/time/Duration")
	if err != nil {
		return err
	}
	jDurationClass = C.jclass(class)
	class, err = JFindClass(env, "java/lang/Throwable")
	if err != nil {
		return err
	}
	jThrowableClass = C.jclass(class)
	class, err = JFindClass(env, "java/lang/System")
	if err != nil {
		return err
	}
	jSystemClass = C.jclass(class)
	class, err = JFindClass(env, "java/lang/Object")
	if err != nil {
		return err
	}
	jObjectClass = C.jclass(class)
	class, err = JFindClass(env, "java/lang/String")
	if err != nil {
		return err
	}
	jStringClass = C.jclass(class)
	class, err = JFindClass(env, "java/util/HashMap")
	if err != nil {
		return err
	}
	jHashMapClass = C.jclass(class)
	class, err = JFindClass(env, "[B")
	if err != nil {
		return err
	}
	jByteArrayClass = C.jclass(class)
	if status := C.GetJavaVM(env, &jVM); status != 0 {
		return fmt.Errorf("couldn't get Java VM from the (Java) environment")
	}
	return nil
}

// CamelCase converts ThisString to thisString.
func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

// UpperCamelCase converts thisString to ThisString.
func UpperCamelCase(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}

// GoString returns a Go string given the Java string.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoString(jEnv, jStr interface{}) string {
	env := getEnv(jEnv)
	str := getString(jStr)
	if str == nil {
		return ""
	}
	cString := C.GetStringUTFChars(env, str, nil)
	defer C.ReleaseStringUTFChars(env, str, cString)
	return C.GoString(cString)
}

// GetEnv returns the Java environment for the running thread, creating a new
// one if it doesn't already exist.  This method also returns a function which
// must be invoked when the returned environment is no longer needed. The
// returned environment can only be used by the thread that invoked this method,
// and the function must be invoked by the same thread as well.
func GetEnv() (jEnv unsafe.Pointer, free func()) {
	// Lock the goroutine to the current OS thread.  This is necessary as
	// *C.JNIEnv must not be shared across threads.  The scenario that can break
	// this requrement is:
	//   - goroutine A executing on thread X, obtaining a *C.JNIEnv pointer P.
	//   - goroutine A gets re-scheduled on thread Y, maintaining the pointer P.
	//   - goroutine B starts executing on thread X, obtaining pointer P.
	//
	// By locking the goroutines to their OS thread while they hold the pointer
	// to *C.JNIEnv, the above scenario can never occur.
	runtime.LockOSThread()
	var env *C.JNIEnv
	if C.GetEnv(jVM, &env, C.JNI_VERSION_1_6) == C.JNI_OK {
		return unsafe.Pointer(env), func() {
			runtime.UnlockOSThread()
		}
	}
	// Couldn't get env, attach the thread.
	C.AttachCurrentThread(jVM, &env, nil)
	return unsafe.Pointer(env), func() {
		C.DetachCurrentThread(jVM)
		runtime.UnlockOSThread()
	}
}

// IsInstanceOf returns true iff the provided Java object is an instance of the
// provided Java class.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func IsInstanceOf(jEnv, jObj, jClass interface{}) bool {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	class := getClass(jClass)
	return C.IsInstanceOf(env, obj, class) == C.JNI_TRUE
}

// IsJavaObject returns true iff the provided value is of type C.jobject.
// Note that this test is somewhat flawed: if the object is created by a package
// other than this one, the type will be C.jobject local to that package, which
// will fail the test below.  (This is because CGO creates package-local types.)
// So this test should be trusted only if we are sure that the passed-in object
// value would have been created by this package.
func IsJavaObject(value interface{}) bool {
	_, ok := value.(C.jobject)
	return ok
}

// JString returns a Java string given the Go string.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JString(jEnv interface{}, str string) unsafe.Pointer {
	env := getEnv(jEnv)
	cString := C.CString(str)
	defer C.free(unsafe.Pointer(cString))
	return unsafe.Pointer(C.NewStringUTF(env, cString))
}

// JThrow throws a new Java exception of the provided type with the given message.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JThrow(jEnv, jClass interface{}, msg string) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	s := C.CString(msg)
	defer C.free(unsafe.Pointer(s))
	C.ThrowNew(env, class, s)
}

// JThrowV throws a new Java VException corresponding to the given error.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JThrowV(jEnv interface{}, err error) {
	jObj, errNew := JVException(jEnv, err)
	if errNew != nil {
		log.Printf("Couldn't throw exception %q: %v", err, errNew)
		return
	}
	if jObj == nil {
		log.Printf("Couldn't throw exception %q: got null Java exception", err)
		return
	}
	env := getEnv(jEnv)
	C.Throw(env, C.jthrowable(jObj))
}

// JVException returns the Java VException given the Go error.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JVException(jEnv interface{}, err error) (unsafe.Pointer, error) {
	if err == nil {
		return nil, nil
	}
	data, err := vom.Encode(err)
	if err != nil {
		return nil, err
	}
	return JVomDecode(jEnv, data, jVExceptionClass)
}

// JExceptionMsg returns the exception message if an exception occurred, or
// nil otherwise.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JExceptionMsg(jEnv interface{}) error {
	env := getEnv(jEnv)
	e := C.ExceptionOccurred(env)
	if e == nil { // no exception
		return nil
	}
	C.ExceptionClear(env)
	id, err := JMethodID(env, jThrowableClass, "getMessage", FuncSign(nil, StringSign))
	if err != nil {
		panic(err.Error())
	}
	jMsg := C.CallGetExceptionMessage(env, C.jobject(e), C.jmethodID(id))
	return errors.New(GoString(env, jMsg))
}

// JObjectField returns the value of the provided Java object's Object field, or
// error if the field value couldn't be retrieved.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JObjectField(jEnv, jObj interface{}, field string) (unsafe.Pointer, error) {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid, err := JFieldIDPtr(env, C.GetObjectClass(env, obj), field, ObjectSign)
	if err != nil {
		return nil, err
	}
	return unsafe.Pointer(C.GetObjectField(env, obj, C.jfieldID(fid))), nil
}

// JBoolField returns the value of the provided Java object's boolean field, or
// error if the field value couldn't be retrieved.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JBoolField(jEnv, jObj interface{}, field string) (bool, error) {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid, err := JFieldIDPtr(env, C.GetObjectClass(env, obj), field, BoolSign)
	if err != nil {
		return false, err
	}
	return C.GetBooleanField(env, obj, C.jfieldID(fid)) != C.JNI_FALSE, nil
}

// JIntField returns the value of the provided Java object's int field, or
// error if the field value couldn't be retrieved.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JIntField(jEnv, jObj interface{}, field string) (int, error) {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid, err := JFieldIDPtr(env, C.GetObjectClass(env, obj), field, IntSign)
	if err != nil {
		return -1, err
	}
	return int(C.GetIntField(env, obj, C.jfieldID(fid))), nil
}

// JStringField returns the value of the provided Java object's String field, or
// error if the field value couldn't be retrieved.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStringField(jEnv, jObj interface{}, field string) (string, error) {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid, err := JFieldIDPtr(env, C.GetObjectClass(env, obj), field, StringSign)
	if err != nil {
		return "", err
	}
	jStr := C.jstring(C.GetObjectField(env, obj, C.jfieldID(fid)))
	return GoString(env, jStr), nil
}

// JStringArrayField returns the value of the provided object's String[] field,
// or error if the field value couldn't be retrieved.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStringArrayField(jEnv, jObj interface{}, field string) ([]string, error) {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid, err := JFieldIDPtr(env, C.GetObjectClass(env, obj), field, ArraySign(StringSign))
	if err != nil {
		return nil, err
	}
	jStrArray := C.jobjectArray(C.GetObjectField(env, obj, C.jfieldID(fid)))
	return GoStringArray(env, jStrArray), nil
}

// JByteArrayField returns the value of the provided object's byte[] field, or
// error if the field value couldn't be retrieved.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JByteArrayField(jEnv, jObj interface{}, field string) ([]byte, error) {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid, err := JFieldIDPtr(env, C.GetObjectClass(env, obj), field, ArraySign(ByteSign))
	if err != nil {
		return nil, err
	}
	arr := C.jbyteArray(C.GetObjectField(env, obj, C.jfieldID(fid)))
	if arr == nil {
		return nil, nil
	}
	return GoByteArray(env, arr), nil
}

// JByteArrayArrayField returns the value of the provided object's byte[][]
// field, or error if the field value couldn't be retrieved.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JByteArrayArrayField(jEnv, jObj interface{}, field string) ([][]byte, error) {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid, err := JFieldIDPtr(env, C.GetObjectClass(env, obj), field, ArraySign(ArraySign(ByteSign)))
	if err != nil {
		return nil, err
	}
	arr := C.jobjectArray(C.GetObjectField(env, obj, C.jfieldID(fid)))
	if arr == nil {
		return nil, nil
	}
	return GoByteArrayArray(env, arr), nil
}

func JStaticObjectField(jEnv, jClass interface{}, field string, fieldTypeSign Sign) (C.jobject, error) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	fid, err := JStaticFieldIDPtr(env, class, field, fieldTypeSign)
	if err != nil {
		return nil, err
	}
	return C.jobject(C.GetStaticObjectField(env, class, C.jfieldID(fid))), nil
}

// JStaticStringField returns the value of the static String field of the
// provided Java class, or error if the field value couldn't be retrieved.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStaticStringField(jEnv, jClass interface{}, field string) (string, error) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	fid, err := JStaticFieldIDPtr(env, class, field, StringSign)
	if err != nil {
		return "", err
	}
	jStr := C.jstring(C.GetStaticObjectField(env, class, C.jfieldID(fid)))
	return GoString(env, jStr), nil
}

// JObjectArray converts the provided slice of C.jobject pointers into a Java
// array of the provided element type.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JObjectArray(jEnv interface{}, arr []interface{}, jElemClass interface{}) unsafe.Pointer {
	if arr == nil {
		return nil
	}
	env := getEnv(jEnv)
	jElementClass := getClass(jElemClass)
	ret := C.NewObjectArray(env, C.jsize(len(arr)), jElementClass, nil)
	for i, elem := range arr {
		jElem := getObject(elem)
		C.SetObjectArrayElement(env, ret, C.jsize(i), jElem)
	}
	return unsafe.Pointer(ret)
}

// GoObjectArray converts a Java Object array to a Go slice of C.jobject
// pointers.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoObjectArray(jEnv, jObjArray interface{}) []unsafe.Pointer {
	env := getEnv(jEnv)
	jArr := getObjectArray(jObjArray)
	if jArr == nil {
		return nil
	}
	length := C.GetArrayLength(env, C.jarray(jArr))
	ret := make([]unsafe.Pointer, int(length))
	for i := 0; i < int(length); i++ {
		ret[i] = unsafe.Pointer(C.GetObjectArrayElement(env, jArr, C.jsize(i)))
	}
	return ret
}

// JStringArray converts the provided slice of Go strings into a Java array of
// strings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStringArray(jEnv interface{}, strs []string) unsafe.Pointer {
	if strs == nil {
		return nil
	}
	env := getEnv(jEnv)
	ret := C.NewObjectArray(env, C.jsize(len(strs)), jStringClass, nil)
	for i, str := range strs {
		C.SetObjectArrayElement(env, ret, C.jsize(i), C.jobject(JString(env, str)))
	}
	return unsafe.Pointer(ret)
}

// GoStringArray converts a Java string array to a Go string array.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoStringArray(jEnv, jStrArray interface{}) []string {
	env := getEnv(jEnv)
	jArr := getObjectArray(jStrArray)
	if jArr == nil {
		return nil
	}
	length := C.GetArrayLength(env, C.jarray(jArr))
	ret := make([]string, int(length))
	for i := 0; i < int(length); i++ {
		ret[i] = GoString(env, C.jstring(C.GetObjectArrayElement(env, jArr, C.jsize(i))))
	}
	return ret
}

// JByteArray converts the provided Go byte slice into a Java byte array.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JByteArray(jEnv interface{}, bytes []byte) unsafe.Pointer {
	if bytes == nil {
		return nil
	}
	env := getEnv(jEnv)
	ret := C.NewByteArray(env, C.jsize(len(bytes)))
	if len(bytes) > 0 {
		C.SetByteArrayRegion(env, ret, 0, C.jsize(len(bytes)), (*C.jbyte)(unsafe.Pointer(&bytes[0])))
	}
	return unsafe.Pointer(ret)
}

// GoByteArray converts the provided Java byte array into a Go byte slice.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoByteArray(jEnv, jArr interface{}) (ret []byte) {
	env := getEnv(jEnv)
	arr := getByteArray(jArr)
	if arr == nil {
		return
	}
	length := int(C.GetArrayLength(env, C.jarray(arr)))
	ret = make([]byte, length)
	bytes := C.GetByteArrayElements(env, arr, nil)
	defer C.ReleaseByteArrayElements(env, arr, bytes, C.JNI_ABORT)
	ptr := bytes
	for i := 0; i < length; i++ {
		ret[i] = byte(*ptr)
		ptr = (*C.jbyte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + unsafe.Sizeof(*ptr)))
	}
	return
}

// JByteArrayArray converts the provided [][]byte value into a Java array of
// byte arrays.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JByteArrayArray(jEnv interface{}, arr [][]byte) unsafe.Pointer {
	if arr == nil {
		return nil
	}
	env := getEnv(jEnv)
	ret := C.NewObjectArray(env, C.jsize(len(arr)), jByteArrayClass, nil)
	for i, elem := range arr {
		jElem := JByteArray(env, elem)
		C.SetObjectArrayElement(env, ret, C.jsize(i), C.jobject(jElem))
	}
	return unsafe.Pointer(ret)
}

// GoByteArrayArray converts the provided Java array of byte arrays into a Go
// [][]byte value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoByteArrayArray(jEnv, jArr interface{}) (ret [][]byte) {
	env := getEnv(jEnv)
	arr := getObjectArray(jArr)
	if arr == nil {
		return
	}
	length := int(C.GetArrayLength(env, C.jarray(arr)))
	ret = make([][]byte, length)
	for i := 0; i < length; i++ {
		ret[i] = GoByteArray(env, C.GetObjectArrayElement(env, arr, C.jsize(i)))
	}
	return
}

// JVDLValueArray converts the provided Go slice of *vdl.Value values into a
// Java array of VdlValue objects.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JVDLValueArray(jEnv interface{}, arr []*vdl.Value) (unsafe.Pointer, error) {
	valarr := make([]interface{}, len(arr))
	for i, val := range arr {
		var err error
		if valarr[i], err = JVomCopy(jEnv, val, jVdlValueClass); err != nil {
			return nil, err
		}
	}
	return JObjectArray(jEnv, valarr, jVdlValueClass), nil
}

// GoVDLValueArray converts the provided Java array of VdlValue objects into a
// Go slice of *vdl.Value values.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoVDLValueArray(jEnv, jArr interface{}) ([]*vdl.Value, error) {
	valarr := GoObjectArray(jEnv, jArr)
	vals := make([]*vdl.Value, len(valarr))
	for i, jVal := range valarr {
		var err error
		// THIS GUY MESSES UP.
		if vals[i], err = GoVomCopyValue(jEnv, jVal); err != nil {
			return nil, err
		}
	}
	return vals, nil
}

// JObjectMap converts the provided Go map of Java objects into a Java
// object map.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JObjectMap(jEnv interface{}, goMap map[interface{}]interface{}) (unsafe.Pointer, error) {
	env := getEnv(jEnv)
	jMap, err := NewObject(env, jHashMapClass, nil)
	if err != nil {
		return nil, err
	}
	for jKey, jVal := range goMap {
		if _, err := CallObjectMethod(env, jMap, "put", []Sign{ObjectSign, ObjectSign}, ObjectSign, getObject(jKey), getObject(jVal)); err != nil {
			return nil, err
		}
	}
	return unsafe.Pointer(jMap), nil
}

// GoObjectMap converts the provided Java object map into a Go map of Java
// objects.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoObjectMap(jEnv, jObjMap interface{}) (map[unsafe.Pointer]unsafe.Pointer, error) {
	env := getEnv(jEnv)
	jMap := getObject(jObjMap)
	jKeySet, err := CallObjectMethod(env, jMap, "keySet", nil, SetSign)
	if err != nil {
		return nil, err
	}
	keysArr, err := CallObjectArrayMethod(env, jKeySet, "toArray", nil, ObjectSign)
	if err != nil {
		return nil, err
	}
	ret := make(map[unsafe.Pointer]unsafe.Pointer)
	for _, jKey := range keysArr {
		jVal, err := CallObjectMethod(env, jMap, "get", []Sign{ObjectSign}, ObjectSign, jKey)
		if err != nil {
			return nil, err
		}
		ret[jKey] = jVal
	}
	return ret, nil
}

// JFieldIDPtr returns the Java field ID for the given object (i.e., non-static)
// field, or an error if the field couldn't be found.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JFieldIDPtr(jEnv, jClass interface{}, name string, sign Sign) (unsafe.Pointer, error) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSign := C.CString(string(sign))
	defer C.free(unsafe.Pointer(cSign))
	ptr := unsafe.Pointer(C.GetFieldID(env, class, cName, cSign))
	if err := JExceptionMsg(env); err != nil || ptr == nil {
		return nil, fmt.Errorf("couldn't find field %s: %v", name, err)
	}
	return ptr, nil
}

// JStaticFieldIDPtr returns the Java field ID for the given static field,
// or an error if the field couldn't be found.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStaticFieldIDPtr(jEnv, jClass interface{}, name string, sign Sign) (unsafe.Pointer, error) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSign := C.CString(string(sign))
	defer C.free(unsafe.Pointer(cSign))
	ptr := unsafe.Pointer(C.GetStaticFieldID(env, class, cName, cSign))
	if err := JExceptionMsg(env); err != nil || ptr == nil {
		return nil, fmt.Errorf("couldn't find field %s: %v", name, err)
	}
	return ptr, nil
}

// JMethodID returns the Java method ID for the given instance (non-static)
// method, or an error if the method couldn't be found.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JMethodID(jEnv, jClass interface{}, name string, signature Sign) (unsafe.Pointer, error) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSignature := C.CString(string(signature))
	defer C.free(unsafe.Pointer(cSignature))
	mid := C.GetMethodID(env, class, cName, cSignature)
	if err := JExceptionMsg(env); err != nil || mid == C.jmethodID(nil) {
		return nil, fmt.Errorf("couldn't find method %q with signature %v.", name, signature)
	}
	return unsafe.Pointer(mid), nil
}

// JStaticMethodID returns the Java method ID for the given static method, or an
// error if the method couldn't be found.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStaticMethodID(jEnv, jClass interface{}, name string, signature Sign) (unsafe.Pointer, error) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSignature := C.CString(string(signature))
	defer C.free(unsafe.Pointer(cSignature))
	mid := C.GetStaticMethodID(env, class, cName, cSignature)
	if err := JExceptionMsg(env); err != nil || mid == C.jmethodID(nil) {
		return nil, fmt.Errorf("couldn't find method %s with a given signature.", name)
	}
	return unsafe.Pointer(mid), nil
}

// JFindClass returns the global reference to the Java class with the
// given pathname, or an error if the class cannot be found.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JFindClass(jEnv interface{}, name string) (unsafe.Pointer, error) {
	env := getEnv(jEnv)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	class := C.FindClass(env, cName)
	if err := JExceptionMsg(env); err != nil || class == nil {
		return nil, fmt.Errorf("couldn't find class %s: %v", name, err)
	}
	return unsafe.Pointer(C.NewGlobalRef(env, C.jobject(class))), nil
}

// GoTime converts the provided Java DateTime object into a Go time.Time value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoTime(jEnv, jTimeObj interface{}) (time.Time, error) {
	jTime := getObject(jTimeObj)
	if jTime == nil {
		return time.Time{}, nil
	}
	millis, err := CallLongMethod(jEnv, jTime, "getMillis", nil)
	if err != nil {
		return time.Time{}, err
	}
	sec := millis / 1000
	nsec := (millis % 1000) * 1000000
	return time.Unix(sec, nsec), nil
}

// JTime converts the provided Go time.Time value into a Java DateTime
// object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JTime(jEnv interface{}, t time.Time) (unsafe.Pointer, error) {
	millis := t.UnixNano() / 1000000
	return NewObject(jEnv, jDateTimeClass, []Sign{LongSign}, millis)
}

// GoDuration converts the provided Java Duration object into a Go time.Duration
// value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoDuration(jEnv interface{}, jDuration interface{}) (time.Duration, error) {
	millis, err := CallLongMethod(jEnv, jDuration, "getMillis", nil)
	if err != nil {
		return 0, err
	}
	return time.Duration(millis) * time.Millisecond, nil
}

// JDuration converts the provided Go time.Duration value into a Java
// Duration object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JDuration(jEnv interface{}, d time.Duration) (unsafe.Pointer, error) {
	millis := d.Nanoseconds() / 1000000
	return NewObject(jEnv, jDurationClass, []Sign{LongSign}, int64(millis))
}

// Various functions that cast CGO types from various other packages into this
// package's types.
func getEnv(jEnv interface{}) *C.JNIEnv {
	return (*C.JNIEnv)(unsafe.Pointer(PtrValue(jEnv)))
}
func getJVM(jVM interface{}) *C.JavaVM {
	return (*C.JavaVM)(unsafe.Pointer(PtrValue(jVM)))
}
func getByteArray(jByteArray interface{}) C.jbyteArray {
	return C.jbyteArray(unsafe.Pointer(PtrValue(jByteArray)))
}
func getObject(jObj interface{}) C.jobject {
	return C.jobject(unsafe.Pointer(PtrValue(jObj)))
}
func getClass(jClass interface{}) C.jclass {
	return C.jclass(unsafe.Pointer(PtrValue(jClass)))
}
func getString(jString interface{}) C.jstring {
	return C.jstring(unsafe.Pointer(PtrValue(jString)))
}
func getObjectArray(jArray interface{}) C.jobjectArray {
	return C.jobjectArray(unsafe.Pointer(PtrValue(jArray)))
}
