// +build android

package main

import (
	"flag"

	"golang.org/x/mobile/app"

	jutil "v.io/jni/util"
	jgoogle "v.io/jni/veyron/runtimes/google"
	jveyron2 "v.io/jni/veyron2"
	_ "v.io/core/veyron/profiles/roaming"
	"v.io/core/veyron2/vom2"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env *C.JNIEnv) {}

//export Java_io_veyron_veyron_veyron2_VRuntime_nativeInit
func Java_io_veyron_veyron_veyron2_VRuntime_nativeInit(env *C.JNIEnv, jVRuntimeClass C.jclass) {
	Init(env)
	jutil.Init(env)
	jveyron2.Init(env)
	jgoogle.Init(env)
}

func main() {
	// Explicitly enable VOM2 encoding.
	vom2.SetEnabled(true)
	// Send all logging to stderr, so that the output is visible in Android.  Note that if this
	// flag is removed, the process will likely crash as android requires that all logs are written
	// into a specific directory.
	flag.Set("logtostderr", "true")
	app.Run(app.Callbacks{})
}
