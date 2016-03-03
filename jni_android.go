// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package jni

import (
	"unsafe"

	"v.io/x/lib/vlog"
	_ "v.io/x/ref/runtime/factories/android"

	jdplugins "v.io/x/jni/impl/google/discovery/plugins"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

//export Java_io_v_android_v23_V_nativeInitGlobalAndroid
func Java_io_v_android_v23_V_nativeInitGlobalAndroid(jenv *C.JNIEnv, jVClass C.jclass, jOptions C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jOpts := jutil.Object(uintptr(unsafe.Pointer(jOptions)))

	// Setup logging.
	_, _, level, vmodule, err := loggingOpts(env, jOpts)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	// Disable any logging to STDERR.
	// This assumes that vlog.Log is the underlying logging system for.
	vlog.Log.Configure(vlog.OverridePriorConfiguration(true), vlog.LogToStderr(false), vlog.AlsoLogToStderr(false), level, vmodule)

	// Setup discovery plugins.
	if err := jdplugins.Init(env); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}
