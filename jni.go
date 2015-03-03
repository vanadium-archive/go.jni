// +build android

package main

import (
	"flag"

	"golang.org/x/mobile/app"

	jgoogle "v.io/x/jni/core/veyron/runtimes/google"
	jutil "v.io/x/jni/util"
	jv23 "v.io/x/jni/v23"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

//export Java_io_v_v23_V_nativeInit
func Java_io_v_v23_V_nativeInit(env *C.JNIEnv, jVRuntimeClass C.jclass) {
	if err := jutil.Init(env); err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := jv23.Init(env); err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := jgoogle.Init(env); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

func main() {
	// Send all logging to stderr, so that the output is visible in Android.  Note that if this
	// flag is removed, the process will likely crash as android requires that all logs are written
	// into a specific directory.
	flag.Set("logtostderr", "true")
	app.Run(app.Callbacks{})
}
