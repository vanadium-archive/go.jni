// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package context

import (
	"unsafe"

	"v.io/v23/context"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
// #include <stdlib.h>
import "C"

var (
	classSign = jutil.ClassSign("java.lang.Class")
	// Global reference for io.v.v23.context.VContextImpl class.
	jVContextImplClass jutil.Class
	// Global reference for java.jutil.concurrent.CountDownLatch class.
	jCountDownLatchClass jutil.Class
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
func Init(env jutil.Env) error {
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	var err error
	jVContextImplClass, err = jutil.JFindClass(env, "io/v/v23/context/VContextImpl")
	if err != nil {
		return err
	}
	jCountDownLatchClass, err = jutil.JFindClass(env, "java/util/concurrent/CountDownLatch")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_v23_context_VContextImpl_nativeCreate
func Java_io_v_v23_context_VContextImpl_nativeCreate(jenv *C.JNIEnv, jVContextImpl C.jclass) C.jobject {
	env := jutil.WrapEnv(jenv)
	ctx, cancel := context.RootContext()
	jContext, err := JavaContext(env, ctx, cancel)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jContext))
}

//export Java_io_v_v23_context_VContextImpl_nativeDeadline
func Java_io_v_v23_context_VContextImpl_nativeDeadline(jenv *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	d, ok := (*(*context.T)(jutil.NativePtr(goPtr))).Deadline()
	if !ok {
		return nil
	}
	jDeadline, err := jutil.JTime(env, d)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jDeadline))
}

//export Java_io_v_v23_context_VContextImpl_nativeDone
func Java_io_v_v23_context_VContextImpl_nativeDone(jenv *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	c := (*(*context.T)(jutil.NativePtr(goPtr))).Done()
	jCounter, err := JavaCountDownLatch(env, c)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCounter))
}

//export Java_io_v_v23_context_VContextImpl_nativeValue
func Java_io_v_v23_context_VContextImpl_nativeValue(jenv *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, jKey C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	key, err := GoContextKey(env, jutil.WrapObject(jKey))
	value := (*(*context.T)(jutil.NativePtr(goPtr))).Value(key)
	jValue, err := JavaContextValue(env, value)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jValue))
}

//export Java_io_v_v23_context_VContextImpl_nativeWithCancel
func Java_io_v_v23_context_VContextImpl_nativeWithCancel(jenv *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	ctx, cancelFunc := context.WithCancel((*context.T)(jutil.NativePtr(goPtr)))
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_v23_context_VContextImpl_nativeWithDeadline
func Java_io_v_v23_context_VContextImpl_nativeWithDeadline(jenv *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, jDeadline C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	deadline, err := jutil.GoTime(env, jutil.WrapObject(jDeadline))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx, cancelFunc := context.WithDeadline((*context.T)(jutil.NativePtr(goPtr)), deadline)
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_v23_context_VContextImpl_nativeWithTimeout
func Java_io_v_v23_context_VContextImpl_nativeWithTimeout(jenv *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, jTimeout C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	timeout, err := jutil.GoDuration(env, jutil.WrapObject(jTimeout))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx, cancelFunc := context.WithTimeout((*context.T)(jutil.NativePtr(goPtr)), timeout)
	jCtx, err := JavaContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_v23_context_VContextImpl_nativeWithValue
func Java_io_v_v23_context_VContextImpl_nativeWithValue(jenv *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, jKey C.jobject, jValue C.jobject) C.jobject {
	env := jutil.WrapEnv(jenv)
	key, err := GoContextKey(env, jutil.WrapObject(jKey))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	value, err := GoContextValue(env, jutil.WrapObject(jValue))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx := context.WithValue((*context.T)(jutil.NativePtr(goPtr)), key, value)
	jCtx, err := JavaContext(env, ctx, nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_v23_context_VContextImpl_nativeCancel
func Java_io_v_v23_context_VContextImpl_nativeCancel(jenv *C.JNIEnv, jVContextImpl C.jobject, goCancelPtr C.jlong) {
	if goCancelPtr != 0 {
		(*(*context.CancelFunc)(jutil.NativePtr(goCancelPtr)))()
	}
}

//export Java_io_v_v23_context_VContextImpl_nativeFinalize
func Java_io_v_v23_context_VContextImpl_nativeFinalize(jenv *C.JNIEnv, jVContextImpl C.jobject, goPtr C.jlong, goCancelPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
	if goCancelPtr != 0 {
		jutil.GoUnref(jutil.NativePtr(goCancelPtr))
	}
}
