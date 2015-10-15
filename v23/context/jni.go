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
	// Global reference for io.v.v23.context.VContext class.
	jVContextClass jutil.Class
	// Global reference for io.v.v23.context.CancelableVContext class.
	jCancelableVContextClass jutil.Class
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
	jVContextClass, err = jutil.JFindClass(env, "io/v/v23/context/VContext")
	if err != nil {
		return err
	}
	jCancelableVContextClass, err = jutil.JFindClass(env, "io/v/v23/context/CancelableVContext")
	if err != nil {
		return err
	}
	jCountDownLatchClass, err = jutil.JFindClass(env, "java/util/concurrent/CountDownLatch")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_v23_context_VContext_nativeCreate
func Java_io_v_v23_context_VContext_nativeCreate(jenv *C.JNIEnv, jVContext C.jclass) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	ctx, _ := context.RootContext()
	jContext, err := JavaContext(env, ctx)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jContext))
}

//export Java_io_v_v23_context_VContext_nativeDeadline
func Java_io_v_v23_context_VContext_nativeDeadline(jenv *C.JNIEnv, jVContext C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
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

//export Java_io_v_v23_context_VContext_nativeDone
func Java_io_v_v23_context_VContext_nativeDone(jenv *C.JNIEnv, jVContext C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	c := (*(*context.T)(jutil.NativePtr(goPtr))).Done()
	jCounter, err := JavaCountDownLatch(env, c)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCounter))
}

//export Java_io_v_v23_context_VContext_nativeValue
func Java_io_v_v23_context_VContext_nativeValue(jenv *C.JNIEnv, jVContext C.jobject, goPtr C.jlong, jKey C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	key, err := GoContextKey(env, jutil.Object(uintptr(unsafe.Pointer(jKey))))
	value := (*(*context.T)(jutil.NativePtr(goPtr))).Value(key)
	jValue, err := JavaContextValue(env, value)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jValue))
}

//export Java_io_v_v23_context_VContext_nativeWithCancel
func Java_io_v_v23_context_VContext_nativeWithCancel(jenv *C.JNIEnv, jVContext C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	ctx, cancelFunc := context.WithCancel((*context.T)(jutil.NativePtr(goPtr)))
	jCtx, err := JavaCancelableContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_v23_context_VContext_nativeWithDeadline
func Java_io_v_v23_context_VContext_nativeWithDeadline(jenv *C.JNIEnv, jVContext C.jobject, goPtr C.jlong, jDeadline C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	deadline, err := jutil.GoTime(env, jutil.Object(uintptr(unsafe.Pointer(jDeadline))))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx, cancelFunc := context.WithDeadline((*context.T)(jutil.NativePtr(goPtr)), deadline)
	jCtx, err := JavaCancelableContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_v23_context_VContext_nativeWithTimeout
func Java_io_v_v23_context_VContext_nativeWithTimeout(jenv *C.JNIEnv, jVContext C.jobject, goPtr C.jlong, jTimeout C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	timeout, err := jutil.GoDuration(env, jutil.Object(uintptr(unsafe.Pointer(jTimeout))))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx, cancelFunc := context.WithTimeout((*context.T)(jutil.NativePtr(goPtr)), timeout)
	jCtx, err := JavaCancelableContext(env, ctx, cancelFunc)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_v23_context_VContext_nativeWithValue
func Java_io_v_v23_context_VContext_nativeWithValue(jenv *C.JNIEnv, jVContext C.jobject, goPtr C.jlong, jKey C.jobject, jValue C.jobject) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	key, err := GoContextKey(env, jutil.Object(uintptr(unsafe.Pointer(jKey))))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	value, err := GoContextValue(env, jutil.Object(uintptr(unsafe.Pointer(jValue))))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	ctx := context.WithValue((*context.T)(jutil.NativePtr(goPtr)), key, value)
	jCtx, err := JavaContext(env, ctx)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jCtx))
}

//export Java_io_v_v23_context_VContext_nativeFinalize
func Java_io_v_v23_context_VContext_nativeFinalize(jenv *C.JNIEnv, jVContext C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_v23_context_CancelableVContext_nativeCancel
func Java_io_v_v23_context_CancelableVContext_nativeCancel(jenv *C.JNIEnv, jCancelableVContext C.jobject, goCancelPtr C.jlong) {
	if goCancelPtr != 0 {
		(*(*context.CancelFunc)(jutil.NativePtr(goCancelPtr)))()
	}
}

//export Java_io_v_v23_context_CancelableVContext_nativeFinalize
func Java_io_v_v23_context_CancelableVContext_nativeFinalize(jenv *C.JNIEnv, jCancelableVContext C.jobject, goCancelPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goCancelPtr))
}
