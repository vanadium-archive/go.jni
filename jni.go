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

//export Java_io_v_v23_V_nativeInitLib
func Java_io_v_v23_V_nativeInitLib(jenv *C.JNIEnv, jVClass C.jclass) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	// Ignore all args except for the first one.
	if len(os.Args) > 1 {
		os.Args = os.Args[:1]
	}
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
	jOpts := jutil.Object(uintptr(unsafe.Pointer(jOptions)))

	dir, toStderr, level, vmodule, err := loggingOpts(env, jOpts)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	vlog.Log.Configure(vlog.OverridePriorConfiguration(true), dir, toStderr, level, vmodule)
	return jContext
}

func loggingOpts(env jutil.Env, jOpts jutil.Object) (dir vlog.LogDir, toStderr vlog.LogToStderr, level vlog.Level, vmodule vlog.ModuleSpec, err error) {
	var d string
	d, err = jutil.GetStringOption(env, jOpts, "io.v.v23.LOG_DIR")
	if err != nil {
		return
	}
	dir = vlog.LogDir(d)
	var s bool
	s, err = jutil.GetBooleanOption(env, jOpts, "io.v.v23.LOG_TO_STDERR")
	if err != nil {
		return
	}
	toStderr = vlog.LogToStderr(s)
	var l int
	l, err = jutil.GetIntOption(env, jOpts, "io.v.v23.LOG_VLEVEL")
	if err != nil {
		return
	}
	level = vlog.Level(l)
	var m string
	m, err = jutil.GetStringOption(env, jOpts, "io.v.v23.LOG_VMODULE")
	if err != nil {
		return
	}
	err = vmodule.Set(m)
	return
}

func main() {
}
