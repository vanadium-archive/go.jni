// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"unsafe"

	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

var (
	// Global reference for io.v.v23.rpc.NativeCallback class.
	jNativeCallbackClass jutil.Class
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env jutil.Env) error {
	var err error
	jNativeCallbackClass, err = jutil.JFindClass(env, "io/v/v23/rpc/NativeCallback")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_v23_rpc_NativeCallback_nativeOnSuccess
func Java_io_v_v23_rpc_NativeCallback_nativeOnSuccess(jenv *C.JNIEnv, jNativeCallback C.jobject, goSuccessPtr C.jlong, jResultObj C.jobject) {
	jResult := jutil.Object(uintptr(unsafe.Pointer(jResultObj)))
	(*(*func (jutil.Object))(jutil.NativePtr(goSuccessPtr)))(jResult)
}

//export Java_io_v_v23_rpc_NativeCallback_nativeOnFailure
func Java_io_v_v23_rpc_NativeCallback_nativeOnFailure(jenv *C.JNIEnv, jNativeCallback C.jobject, goFailurePtr C.jlong, jVException C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	err := jutil.GoError(env, jutil.Object(uintptr(unsafe.Pointer(jVException))))
	(*(*func (error))(jutil.NativePtr(goFailurePtr)))(err)
}

//export Java_io_v_v23_rpc_NativeCallback_nativeFinalize
func Java_io_v_v23_rpc_NativeCallback_nativeFinalize(jenv *C.JNIEnv, jNativeCallback C.jobject, goSuccessPtr C.jlong, goFailurePtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goSuccessPtr))
	jutil.GoUnref(jutil.NativePtr(goFailurePtr))
}
