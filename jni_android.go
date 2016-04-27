// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package jni

import (
	"fmt"
	"unsafe"

	"v.io/v23/context"
	"v.io/v23/namespace"
	"v.io/v23/naming"

	"v.io/x/lib/vlog"
	"v.io/x/ref/runtime/factories/android"

	jdplugins "v.io/x/jni/impl/google/discovery/plugins"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

var (
	// Global reference for io.v.android.v23.V class.
	jVClass jutil.Class
)

func Init(env jutil.Env) error {
	var err error
	jVClass, err = jutil.JFindClass(env, "io/v/android/v23/V")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_android_v23_V_nativeInitGlobalAndroid
func Java_io_v_android_v23_V_nativeInitGlobalAndroid(jenv *C.JNIEnv, _ C.jclass, jOptions C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jOpts := jutil.Object(uintptr(unsafe.Pointer(jOptions)))

	if err := Init(env); err != nil {
		jutil.JThrowV(env, err)
		return
	}

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

	// Setup namespace.
	android.SetNamespaceFactory(func(ctx *context.T, ns namespace.T, _ ...string) (namespace.T, error) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jContext, err := jcontext.JavaContext(env, ctx, nil)
		if err != nil {
			return nil, err
		}
		contextSign := jutil.ClassSign("io.v.v23.context.VContext")
		wakeupMountRoot, err := jutil.CallStaticStringMethod(env, jVClass, "getWakeupMountRoot", []jutil.Sign{contextSign}, jContext)
		if err != nil {
			return nil, err
		}
		if wakeupMountRoot == "" {
			return ns, nil
		}
		if !naming.Rooted(wakeupMountRoot) {
			return nil, fmt.Errorf("wakeup mount root %s must be ... rooted.", wakeupMountRoot)
		}
		return &wakeupNamespace{
			wakeupMountRoot: wakeupMountRoot,
			ns:              ns,
		}, nil
	})
}
