// +build android

package google

import (
	"encoding/json"
	"unsafe"

	"veyron.io/jni/util"
	"veyron.io/jni/veyron/runtimes/google/android"
	"veyron.io/jni/veyron/runtimes/google/ipc"
	"veyron.io/jni/veyron/runtimes/google/naming"
	"veyron.io/jni/veyron/runtimes/google/security"
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
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) {
	env := (*C.JNIEnv)(unsafe.Pointer(util.PtrValue(jEnv)))
	jEOFExceptionClass = C.jclass(util.JFindClassPtrOrDie(env, "java/io/EOFException"))

	android.Init(env)
	ipc.Init(env)
	naming.Init(env)
	security.Init(env)
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
	return C.jstring(util.JString(env, string(bytes)))
}

//export Java_io_veyron_veyron_veyron_runtimes_google_InputChannel_nativeFinalize
func Java_io_veyron_veyron_veyron_runtimes_google_InputChannel_nativeFinalize(env *C.JNIEnv, jInputChannel C.jobject, goChanPtr C.jlong) {
	util.GoUnref((*chan interface{})(util.Ptr(goChanPtr)))
}
