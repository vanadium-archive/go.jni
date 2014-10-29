// +build android

// Package util provides various JNI utilities shared across our JNI code.
package util

import (
	"errors"
	"fmt"
	"log"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"veyron.io/veyron/veyron2/verror"
)

// #cgo LDFLAGS: -ljniwrapper
// #include <stdlib.h>
// #include "jni_wrapper.h"
// // CGO doesn't support variadic functions so we have to hard-code these
// // functions to match the invoking code. Ugh!
// static jobject CallNewVeyronExceptionObject(JNIEnv* env, jclass class, jmethodID mid, jstring msg, jstring id) {
//   return (*env)->NewObject(env, class, mid, msg, id);
// }
// static jstring CallGetExceptionMessage(JNIEnv* env, jobject obj, jmethodID id) {
//   return (jstring)(*env)->CallObjectMethod(env, obj, id);
// }
import "C"

var (
	// Global reference for io.veyron.veyron.veyron2.ipc.VeyronException class.
	jVeyronExceptionClass C.jclass
	// Global reference for org.joda.time.DateTime class.
	jDateTimeClass C.jclass
	// Global reference for org.joda.time.Duration class.
	jDurationClass C.jclass
	// Global reference for java.lang.Throwable class.
	jThrowableClass C.jclass
	// Global reference for java.lang.System class.
	jSystemClass C.jclass
	// Global reference for java.lang.String class.
	jStringClass C.jclass
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func Init(jEnv interface{}) {
	env := getEnv(jEnv)
	jVeyronExceptionClass = JFindClassOrPrint(env, "io/veyron/veyron/veyron2/ipc/VeyronException")
	jDateTimeClass = JFindClassOrPrint(env, "org/joda/time/DateTime")
	jDurationClass = JFindClassOrPrint(env, "org/joda/time/Duration")
	jThrowableClass = JFindClassOrPrint(env, "java/lang/Throwable")
	jSystemClass = JFindClassOrPrint(env, "java/lang/System")
	jStringClass = JFindClassOrPrint(env, "java/lang/String")
}

// CamelCase converts ThisString to thisString.
func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
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
func GetEnv(javaVM interface{}) (jEnv unsafe.Pointer, free func()) {
	jVM := getJVM(javaVM)
	var env *C.JNIEnv
	if C.GetEnv(jVM, &env, C.JNI_VERSION_1_6) == C.JNI_OK {
		return unsafe.Pointer(env), func() {}
	}
	// Couldn't get env, attach the thread.
	C.AttachCurrentThread(jVM, &env, nil)
	return unsafe.Pointer(env), func() { C.DetachCurrentThread(jVM) }
}

// JString returns a Java string given the Go string.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JString(jEnv interface{}, str string) C.jstring {
	env := getEnv(jEnv)
	cString := C.CString(str)
	defer C.free(unsafe.Pointer(cString))
	return C.NewStringUTF(env, cString)
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

// JThrowV throws a new Java VeyronException corresponding to the given error.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JThrowV(jEnv interface{}, err error) {
	env := getEnv(jEnv)
	verr := verror.Convert(err)
	id, err := JMethodID(env, jVeyronExceptionClass, "<init>", FuncSign([]Sign{StringSign, StringSign}, VoidSign))
	if err != nil {
		panic(err.Error())
	}
	obj := C.jthrowable(C.CallNewVeyronExceptionObject(env, jVeyronExceptionClass, id, JString(env, verr.Error()), JString(env, string(verr.ErrorID()))))
	C.Throw(env, obj)
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
	jMsg := C.CallGetExceptionMessage(env, C.jobject(e), id)
	return errors.New(GoString(env, jMsg))
}

// JObjectField returns the value of the provided Java object's Object field.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JObjectField(jEnv, jObj interface{}, field string) C.jobject {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid := C.jfieldID(JFieldIDPtrOrDie(env, C.GetObjectClass(env, obj), field, ObjectSign))
	return C.GetObjectField(env, obj, fid)
}

// JBoolField returns the value of the provided Java object's boolean field.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JBoolField(jEnv, jObj interface{}, field string) bool {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid := C.jfieldID(JFieldIDPtrOrDie(env, C.GetObjectClass(env, obj), field, BoolSign))
	return C.GetBooleanField(env, obj, fid) != C.JNI_FALSE
}

// JIntField returns the value of the provided Java object's int field.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JIntField(jEnv, jObj interface{}, field string) int {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid := C.jfieldID(JFieldIDPtrOrDie(env, C.GetObjectClass(env, obj), field, IntSign))
	return int(C.GetIntField(env, obj, fid))
}

// JStringField returns the value of the provided Java object's String field,
// as a Go string.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStringField(jEnv, jObj interface{}, field string) string {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid := C.jfieldID(JFieldIDPtrOrDie(env, C.GetObjectClass(env, obj), field, StringSign))
	jStr := C.jstring(C.GetObjectField(env, obj, fid))
	return GoString(env, jStr)
}

// JStringArrayField returns the value of the provided object's String[] field,
// as a slice of Go strings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStringArrayField(jEnv, jObj interface{}, field string) []string {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid := C.jfieldID(JFieldIDPtrOrDie(env, C.GetObjectClass(env, obj), field, ArraySign(StringSign)))
	jStrArray := C.jobjectArray(C.GetObjectField(env, obj, fid))
	return GoStringArray(env, jStrArray)
}

// JByteArrayField returns the value of the provided object's byte[] field as a
// Go byte slice.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JByteArrayField(jEnv, jObj interface{}, field string) []byte {
	env := getEnv(jEnv)
	obj := getObject(jObj)
	fid := C.jfieldID(JFieldIDPtrOrDie(env, C.GetObjectClass(env, obj), field, ArraySign(ByteSign)))
	arr := C.jbyteArray(C.GetObjectField(env, obj, fid))
	if arr == nil {
		return nil
	}
	return GoByteArray(env, arr)
}

// JStaticStringField returns the value of the static String field of the
// provided Java class, as a Go string.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStaticStringField(jEnv, jClass interface{}, field string) string {
	env := getEnv(jEnv)
	class := getClass(jClass)
	fid := C.jfieldID(JStaticFieldIDPtrOrDie(env, class, field, StringSign))
	jStr := C.jstring(C.GetStaticObjectField(env, class, fid))
	return GoString(env, jStr)
}

// JStringArray converts the provided slice of Go strings into a Java array of strings.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStringArray(jEnv interface{}, strs []string) C.jobjectArray {
	if strs == nil {
		return C.jobjectArray(nil)
	}
	env := getEnv(jEnv)
	ret := C.NewObjectArray(env, C.jsize(len(strs)), jStringClass, nil)
	for i, str := range strs {
		C.SetObjectArrayElement(env, ret, C.jsize(i), C.jobject(JString(env, str)))
	}
	return ret
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
func JByteArray(jEnv interface{}, bytes []byte) C.jbyteArray {
	if bytes == nil {
		return C.jbyteArray(nil)
	}
	env := getEnv(jEnv)
	ret := C.NewByteArray(env, C.jsize(len(bytes)))
	if len(bytes) > 0 {
		C.SetByteArrayRegion(env, ret, 0, C.jsize(len(bytes)), (*C.jbyte)(unsafe.Pointer(&bytes[0])))
	}
	return ret
}

// GoByteArray converts the provided Java byte array into a Go byte slice.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoByteArray(jEnv, jArr interface{}) (ret []byte) {
	env := getEnv(jEnv)
	arr := getByteArray(jArr)
	length := int(C.GetArrayLength(env, C.jarray(arr)))
	ret = make([]byte, length)
	bytes := C.GetByteArrayElements(env, arr, nil)
	for i := 0; i < length; i++ {
		ret[i] = byte(*bytes)
		bytes = (*C.jbyte)(unsafe.Pointer(uintptr(unsafe.Pointer(bytes)) + unsafe.Sizeof(*bytes)))
	}
	return
}

// JFieldIDPtrOrDie returns the Java field ID for the given object
// (i.e., non-static) field.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JFieldIDPtrOrDie(jEnv, jClass interface{}, name string, sign Sign) unsafe.Pointer {
	env := getEnv(jEnv)
	class := getClass(jClass)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSign := C.CString(string(sign))
	defer C.free(unsafe.Pointer(cSign))
	ptr := unsafe.Pointer(C.GetFieldID(env, class, cName, cSign))
	if err := JExceptionMsg(env); err != nil || ptr == nil {
		panic(fmt.Sprintf("couldn't find field %s: %v", name, err))
	}
	return ptr
}

// JStaticFieldIDPtrOrDie returns the Java field ID for the given static field.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStaticFieldIDPtrOrDie(jEnv, jClass interface{}, name string, sign Sign) unsafe.Pointer {
	env := getEnv(jEnv)
	class := getClass(jClass)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSign := C.CString(string(sign))
	defer C.free(unsafe.Pointer(cSign))
	ptr := unsafe.Pointer(C.GetStaticFieldID(env, class, cName, cSign))
	if err := JExceptionMsg(env); err != nil || ptr == nil {
		panic(fmt.Sprintf("couldn't find field %s: %v", name, err))
	}
	return ptr
}

// JMethodID returns the Java method ID for the given instance (non-static)
// method, or an error if the method couldn't be found.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JMethodID(jEnv, jClass interface{}, name string, signature Sign) (C.jmethodID, error) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSignature := C.CString(string(signature))
	defer C.free(unsafe.Pointer(cSignature))
	mid := C.GetMethodID(env, class, cName, cSignature)
	if err := JExceptionMsg(env); err != nil || mid == C.jmethodID(nil) {
		return C.jmethodID(nil), fmt.Errorf("couldn't find method %s with signature %v.", name)
	}
	return mid, nil
}

// JStaticMethodID returns the Java method ID for the given static method, or an
// error if the method couldn't be found.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JStaticMethodID(jEnv, jClass interface{}, name string, signature Sign) (C.jmethodID, error) {
	env := getEnv(jEnv)
	class := getClass(jClass)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cSignature := C.CString(string(signature))
	defer C.free(unsafe.Pointer(cSignature))
	mid := C.GetStaticMethodID(env, class, cName, cSignature)
	if err := JExceptionMsg(env); err != nil || mid == C.jmethodID(nil) {
		return C.jmethodID(nil), fmt.Errorf("couldn't find method %s with a given signature.", name)
	}
	return mid, nil
}

// JFindClass returns the global reference to the Java class with the
// given pathname, or an error if the class cannot be found.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JFindClass(jEnv interface{}, name string) (C.jclass, error) {
	env := getEnv(jEnv)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	class := C.FindClass(env, cName)
	if err := JExceptionMsg(env); err != nil || class == nil {
		return nil, fmt.Errorf("couldn't find class %s: %v", name, err)
	}
	return C.jclass(C.NewGlobalRef(env, C.jobject(class))), nil
}

// JFindClassOrPrint returns the global reference to the Java class with the
// given pathname.  If the class cannot be found, it prints an error message and
// return nil.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JFindClassOrPrint(jEnv interface{}, name string) C.jclass {
	ret, err := JFindClass(jEnv, name)
	if err != nil {
		log.Printf("Couldn't find class %q: %v", name, err)
		return nil
	}
	return ret
}

// GoTime converts the provided Java DateTime object into a Go time.Time value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoTime(jEnv, jTime interface{}) (time.Time, error) {
	if jTime == nil {
		return time.Time{}, fmt.Errorf("Nil Java time.")
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
func JTime(jEnv interface{}, t time.Time) (C.jobject, error) {
	millis := t.UnixNano() / 1000000
	jTime, err := NewObject(jEnv, jDateTimeClass, []Sign{LongSign}, millis)
	return C.jobject(jTime), err
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
func JDuration(jEnv interface{}, d time.Duration) (C.jobject, error) {
	millis := d.Nanoseconds() / 1000000
	jDuration, err := NewObject(jEnv, jDurationClass, []Sign{LongSign}, int64(millis))
	return C.jobject(jDuration), err
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
