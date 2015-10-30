// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package channel

import (
	"fmt"
	"unsafe"

	"v.io/v23/verror"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

var (
	// Global reference for io.v.impl.google.channel.ChannelIterable class.
	jChannelIterableClass jutil.Class
	// Global reference for io.v.impl.google.channel.OutputChannelImpl class.
	jOutputChannelImplClass jutil.Class
	// Global reference for java.io.EOFException class.
	jEOFExceptionClass jutil.Class
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env jutil.Env) error {
	var err error
	jChannelIterableClass, err = jutil.JFindClass(env, "io/v/impl/google/channel/ChannelIterable")
	if err != nil {
		return err
	}
	jOutputChannelImplClass, err = jutil.JFindClass(env, "io/v/impl/google/channel/OutputChannelImpl")
	if err != nil {
		return err
	}
	jEOFExceptionClass, err = jutil.JFindClass(env, "java/io/EOFException")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_impl_google_channel_ChannelIterable_nativeReadValue
func Java_io_v_impl_google_channel_ChannelIterable_nativeReadValue(jenv *C.JNIEnv, jChannelIterable C.jobject, goChanPtr C.jlong, goConvertPtr C.jlong) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	c := *(*chan interface{})(jutil.NativePtr(goChanPtr))
	convert := *(*func(jutil.Env, interface{})(jutil.Object, error))(jutil.NativePtr(goConvertPtr))
	val, ok := <-c
	if !ok {  // channel closed
		jutil.JThrow(env, jEOFExceptionClass, "Reached end of input channel.")
		return nil
	}
	jVal, err := convert(env, val)
	if err == nil {
		return C.jobject(unsafe.Pointer(jVal))
	}
	if verr, ok := err.(verror.E); ok && verr.ID == verror.ErrCanceled.ID {  // EOF
		jutil.JThrow(env, jEOFExceptionClass, "User canceled the operation.")
		return nil
	}
	jutil.JThrowV(env, err)
	return nil
}

//export Java_io_v_impl_google_channel_ChannelIterable_nativeFinalize
func Java_io_v_impl_google_channel_ChannelIterable_nativeFinalize(jenv *C.JNIEnv, jChannelIterable C.jobject, goChanPtr C.jlong, goConvertPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goChanPtr))
	jutil.GoUnref(jutil.NativePtr(goConvertPtr))
}

//export Java_io_v_impl_google_channel_OutputChannelImpl_nativeWriteValue
func Java_io_v_impl_google_channel_OutputChannelImpl_nativeWriteValue(jenv *C.JNIEnv, jOutputChannelClass C.jclass, goChanPtr C.jlong, jObject C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	outCh := *(*outputChannel)(jutil.NativePtr(goChanPtr))
	// The other side of the channel is responsible for deleting this
	// global reference.
	if err := outCh.ReadFunc(jutil.NewGlobalRef(env, jutil.Object(uintptr(unsafe.Pointer(jObject))))); err != nil {
		jutil.JThrowV(env, fmt.Errorf("Exception while writing to OutputChannel: %+v", err))
	}
}

//export Java_io_v_impl_google_channel_OutputChannelImpl_nativeClose
func Java_io_v_impl_google_channel_OutputChannelImpl_nativeClose(jenv *C.JNIEnv, jOutputChannelClass C.jclass, goChanPtr C.jlong) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	outCh := *(*outputChannel)(jutil.NativePtr(goChanPtr))

	if err := outCh.CloseFunc(); err != nil {
		jutil.JThrowV(env, fmt.Errorf("Exception while closing OutputChannel: %+v", err))
	}
}

//export Java_io_v_impl_google_channel_OutputChannelImpl_nativeFinalize
func Java_io_v_impl_google_channel_OutputChannelImpl_nativeFinalize(jenv *C.JNIEnv, jOutputChannelClass C.jclass, goChanPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goChanPtr))
}
