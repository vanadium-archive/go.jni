// +build android

package channel

import (
	jutil "v.io/jni/util"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

var (
	// Global reference for io.v.core.veyron.runtimes.google.InputChannel class.
	jInputChannelImplClass C.jclass
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
	jInputChannelImplClass = C.jclass(jutil.JFindClassOrPrint(jEnv, "io/v/core/veyron/runtimes/google/InputChannel"))
	jEOFExceptionClass = C.jclass(jutil.JFindClassOrPrint(jEnv, "java/io/EOFException"))
}

//export Java_io_v_core_veyron_runtimes_google_InputChannel_nativeAvailable
func Java_io_v_core_veyron_runtimes_google_InputChannel_nativeAvailable(env *C.JNIEnv, jInputChannel C.jobject, goChanPtr C.jlong) C.jboolean {
	ch := *(*chan C.jobject)(jutil.Ptr(goChanPtr))
	if len(ch) > 0 {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_core_veyron_runtimes_google_InputChannel_nativeReadValue
func Java_io_v_core_veyron_runtimes_google_InputChannel_nativeReadValue(env *C.JNIEnv, jInputChannel C.jobject, goChanPtr C.jlong) C.jobject {
	ch := *(*chan C.jobject)(jutil.Ptr(goChanPtr))
	jObj, ok := <-ch
	if !ok {
		jutil.JThrow(env, jEOFExceptionClass, "Channel closed.")
		return nil
	}
	return jObj
}

//export Java_io_v_core_veyron_runtimes_google_InputChannel_nativeFinalize
func Java_io_v_core_veyron_runtimes_google_InputChannel_nativeFinalize(env *C.JNIEnv, jInputChannel C.jobject, goChanPtr C.jlong, goSourceChanPtr C.jlong) {
	jutil.GoUnref(*(*chan C.jobject)(jutil.Ptr(goChanPtr)))
	jutil.GoUnref(*(*chan C.jobject)(jutil.Ptr(goSourceChanPtr)))
}
