// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package jni

import (
	"os"
	"unsafe"

	"v.io/x/lib/vlog"

	jgoogle "v.io/x/jni/impl/google"
	jutil "v.io/x/jni/util"
	jv23 "v.io/x/jni/v23"
)

// #include "jni.h"
import "C"

//export Java_io_v_v23_V_nativeInit
func Java_io_v_v23_V_nativeInit(jenv *C.JNIEnv, jVRuntimeClass C.jclass) {
	env := jutil.WrapEnv(uintptr(unsafe.Pointer(jenv)))
	// Ignore all args except for the first one.
	// NOTE(spetrovic): in the future, we could accept all arguments that are
	// actually defined in Go.  We'd have to manually check.
	if len(os.Args) > 1 {
		os.Args = os.Args[:1]
	}
	// Send all logging to stderr, so that the output is visible in android.
	// Note that if this flag is removed, the process will likely crash on
	// android as android requires that all logs are written into the app's
	// local directory.
	vlog.Log.Configure(vlog.LogToStderr(true))
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
}
