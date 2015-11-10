// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package jni

import (
	"unsafe"

	"v.io/v23/context"
	"v.io/x/lib/vlog"
	_ "v.io/x/ref/runtime/factories/android"
	"v.io/x/ref/lib/discovery/factory"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	jdiscovery "v.io/x/jni/libs/discovery"
)

// #include "jni.h"
import "C"

//export Java_io_v_android_v23_V_nativeInit
func Java_io_v_android_v23_V_nativeInit(jenv *C.JNIEnv, jVClass C.jclass, jContext C.jobject, jAndroidContext C.jobject, jOptions C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jCtx := jutil.Object(uintptr(unsafe.Pointer(jContext)))
	jOpts := jutil.Object(uintptr(unsafe.Pointer(jOptions)))

	if err := jdiscovery.Init(env); err != nil {
		jutil.JThrowV(env, err)
		return nil
	}

	factory.SetBleFactory(jdiscovery.NewBleCreator(env, jutil.Object(uintptr(unsafe.Pointer(jAndroidContext)))))

	_, _, level, vmodule, err := loggingOpts(env, jOpts)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	// Setup vlog logger.  We force the logToStderr option so that we don't crash on android trying
	// to create a log file.
	vlog.Log.Configure(vlog.OverridePriorConfiguration(true), vlog.LogToStderr(true), level, vmodule)

	// Setup the android logger.
	logger := NewLogger("alog", int(level))
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
