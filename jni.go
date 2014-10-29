// +build android

package main

import (
	"flag"
	"unsafe"

	"veyron.io/jni/util"
	"veyron.io/jni/veyron/runtimes/google"
	"veyron.io/jni/veyron2"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env *C.JNIEnv) {}

//export JNI_OnLoad
func JNI_OnLoad(jVM *C.JavaVM, reserved unsafe.Pointer) C.jint {
	jEnv, freeFunc := util.GetEnv(jVM)
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	Init(env)
	util.Init(env)
	veyron2.Init(env)
	google.Init(env)
	return C.JNI_VERSION_1_6
}

func main() {
	// Send all logging to stderr, so that the output is visible in Android.  Note that if this
	// flag is removed, the process will likely crash as android requires that all logs are written
	// into a specific directory.
	flag.Set("logtostderr", "true")
}
