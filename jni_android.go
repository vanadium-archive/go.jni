// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package jni

import (
	"unsafe"

	"v.io/v23/context"
	"v.io/x/lib/vlog"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

//export Java_io_v_android_v23_V_nativeInitLogging
func Java_io_v_android_v23_V_nativeInitLogging(jenv *C.JNIEnv, jVClass C.jclass, jContext C.jobject, jOptions C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jCtx := jutil.Object(uintptr(unsafe.Pointer(jContext)))
	jOpts := jutil.Object(uintptr(unsafe.Pointer(jOptions)))

	level, err := jutil.GetIntOption(env, jOpts, "io.v.v23.LOG_LEVEL")
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	// Send all vlog logs to stderr so that we don't crash on android trying to create a log file.
	vlog.Log.Configure(vlog.OverridePriorConfiguration(true), vlog.LogToStderr(true), vlog.Level(level))
	// Setup the android logger.
	logger := NewLogger("alog", level)
	// Attach the new logger to the context.
	ctx, err := jcontext.GoContext(env, jCtx)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	newCtx := context.WithLogger(ctx, logger)
	jNewCtx, err := jcontext.JavaContext(env, newCtx)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jNewCtx))
}
