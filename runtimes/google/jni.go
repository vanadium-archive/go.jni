// +build android

package main

import (
	"encoding/json"
	"flag"
	"unsafe"

	"veyron.io/jni/runtimes/google/ipc"
	"veyron.io/jni/runtimes/google/naming"
	"veyron.io/jni/runtimes/google/security"
	"veyron.io/jni/runtimes/google/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	// Global reference for java.io.EOFException class.
	jEOFExceptionClass C.jclass
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env *C.JNIEnv) {
	jEOFExceptionClass = C.jclass(util.JFindClassPtrOrDie(env, "java/io/EOFException"))
}

//export JNI_OnLoad
func JNI_OnLoad(jVM *C.JavaVM, reserved unsafe.Pointer) C.jint {
	envPtr, freeFunc := util.GetEnv(jVM)
	env := (*C.JNIEnv)(envPtr)
	defer freeFunc()

	Init(env)
	util.Init(env)
	ipc.Init(env)
	security.Init(env)
	naming.Init(env)
	return C.JNI_VERSION_1_6
}

//export Java_io_veyron_veyron_veyron_runtimes_google_InputChannel_nativeAvailable
func Java_io_veyron_veyron_veyron_runtimes_google_InputChannel_nativeAvailable(env *C.JNIEnv, jInputChannel C.jobject, goChanPtr C.jlong) C.jboolean {
	ch := *(*chan interface{})(util.Ptr(goChanPtr))
	if len(ch) > 0 {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_veyron_veyron_veyron_runtimes_google_InputChannel_nativeReadValue
func Java_io_veyron_veyron_veyron_runtimes_google_InputChannel_nativeReadValue(env *C.JNIEnv, jInputChannel C.jobject, goChanPtr C.jlong) C.jstring {
	ch := *(*chan interface{})(util.Ptr(goChanPtr))
	val, ok := <-ch
	if !ok {
		util.JThrow(env, jEOFExceptionClass, "Channel closed.")
		return nil
	}
	bytes, err := json.Marshal(val)
	if err != nil {
		util.JThrowV(env, err)
		return nil
	}
	return C.jstring(util.JStringPtr(env, string(bytes)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_InputChannel_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_InputChannel_nativeFinalize(env *C.JNIEnv, jInputChannel C.jobject, goChanPtr C.jlong) {
	util.GoUnref((*chan interface{})(util.Ptr(goChanPtr)))
}

func main() {
	// Send all logging to stderr, so that the output is visible in Android.  Note that if this
	// flag is removed, the process will likely crash as android requires that all logs are written
	// into a specific directory.
	flag.Set("logtostderr", "true")
}
