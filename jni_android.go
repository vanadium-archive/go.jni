// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package jni

import (
	"unsafe"

	"v.io/x/lib/vlog"
	"v.io/x/ref/lib/discovery/factory"
	_ "v.io/x/ref/runtime/factories/android"

	jdiscovery "v.io/x/jni/libs/discovery"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

//export Java_io_v_android_v23_V_nativeInitAndroid
func Java_io_v_android_v23_V_nativeInitAndroid(jenv *C.JNIEnv, jVClass C.jclass, jAndroidContext C.jobject, jOptions C.jobject) {
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

	// Setup discovery.
	if err := jdiscovery.Init(env); err != nil {
		jutil.JThrowV(env, err)
		return
	}
	factory.SetBleFactory(jdiscovery.NewBleCreator(env, jutil.Object(uintptr(unsafe.Pointer(jAndroidContext)))))
}
