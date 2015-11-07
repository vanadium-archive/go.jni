// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package jni

import (
	"os"
	"unsafe"

	"v.io/v23/context"
	"v.io/x/lib/vlog"

	jcontext "v.io/x/jni/v23/context"
	jgoogle "v.io/x/jni/impl/google"
	jutil "v.io/x/jni/util"
	jv23 "v.io/x/jni/v23"
)

// #include "jni.h"
import "C"

//export Java_io_v_v23_V_nativeInitLib
func Java_io_v_v23_V_nativeInitLib(jenv *C.JNIEnv, jVClass C.jclass) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	// Ignore all args except for the first one.
	if len(os.Args) > 1 {
		os.Args = os.Args[:1]
	}
	// Send all vlog logs to stderr during the init so that we don't crash on android trying
	// to create a log file.  These settings will be overwritten in nativeInitLogging below.
	vlog.Log.Configure(vlog.OverridePriorConfiguration(true), vlog.LogToStderr(true))

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

//export Java_io_v_v23_V_nativeInitLogging
func Java_io_v_v23_V_nativeInitLogging(jenv *C.JNIEnv, jVClass C.jclass, jContext C.jobject, jOptions C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jCtx := jutil.Object(uintptr(unsafe.Pointer(jContext)))
	jOpts := jutil.Object(uintptr(unsafe.Pointer(jOptions)))

	level, err := jutil.GetIntOption(env, jOpts, "io.v.v23.LOG_LEVEL")
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	logToStderr, err := jutil.GetBooleanOption(env, jOpts, "io.v.v23.LOG_TO_STDERR")
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	logger := vlog.NewLogger("jlog")
	logger.Configure(vlog.OverridePriorConfiguration(true), vlog.Level(level), vlog.LogToStderr(logToStderr))
	// Configure the vlog package to use the new logger.
	vlog.Log = logger
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

func main() {
}
