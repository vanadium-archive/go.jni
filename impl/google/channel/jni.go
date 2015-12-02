// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package channel

import (
	"unsafe"

	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

var (
	// Global reference for io.v.impl.google.channel.InputChannelImpl class.
	jInputChannelImplClass jutil.Class
	// Global reference for io.v.impl.google.channel.OutputChannelImpl class.
	jOutputChannelImplClass jutil.Class
)

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
func Init(env jutil.Env) error {
	var err error
	jInputChannelImplClass, err = jutil.JFindClass(env, "io/v/impl/google/channel/InputChannelImpl")
	if err != nil {
		return err
	}
	jOutputChannelImplClass, err = jutil.JFindClass(env, "io/v/impl/google/channel/OutputChannelImpl")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_impl_google_channel_InputChannelImpl_nativeRecv
func Java_io_v_impl_google_channel_InputChannelImpl_nativeRecv(jenv *C.JNIEnv, jInputChannelImpl C.jobject, goRecvPtr C.jlong, jCallbackObj C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	recv := *(*func() (jutil.Object, error))(jutil.NativePtr(goRecvPtr))
	jCallback := jutil.Object(uintptr(unsafe.Pointer(jCallbackObj)))
	jutil.DoAsyncCall(env, jCallback, recv)
}

//export Java_io_v_impl_google_channel_InputChannelImpl_nativeFinalize
func Java_io_v_impl_google_channel_InputChannelImpl_nativeFinalize(jenv *C.JNIEnv, jInputChannelImpl C.jobject, goRecvPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goRecvPtr))
}

//export Java_io_v_impl_google_channel_OutputChannelImpl_nativeSend
func Java_io_v_impl_google_channel_OutputChannelImpl_nativeSend(jenv *C.JNIEnv, jOutputChannelClass C.jclass, goConvertPtr C.jlong, goSendPtr C.jlong, jItemObj C.jobject, jCallbackObj C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	convert := *(*func(jutil.Object) (interface{}, error))(jutil.NativePtr(goConvertPtr))
	send := *(*func(interface{}) error)(jutil.NativePtr(goSendPtr))
	jItem := jutil.Object(uintptr(unsafe.Pointer(jItemObj)))
	jCallback := jutil.Object(uintptr(unsafe.Pointer(jCallbackObj)))
	// NOTE(spetrovic): Conversion must be done outside of DoAsyncCall as it references a Java
	// object.
	item, err := convert(jItem)
	if err != nil {
		jutil.CallbackOnFailure(env, jCallback, err)
		return
	}
	jutil.DoAsyncCall(env, jCallback, func() (jutil.Object, error) {
		return jutil.NullObject, send(item)
	})
}

//export Java_io_v_impl_google_channel_OutputChannelImpl_nativeClose
func Java_io_v_impl_google_channel_OutputChannelImpl_nativeClose(jenv *C.JNIEnv, jOutputChannelClass C.jclass, goClosePtr C.jlong, jCallbackObj C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	close := *(*func() error)(jutil.NativePtr(goClosePtr))
	jCallback := jutil.Object(uintptr(unsafe.Pointer(jCallbackObj)))
	jutil.DoAsyncCall(env, jCallback, func() (jutil.Object, error) {
		return jutil.NullObject, close()
	})
}

//export Java_io_v_impl_google_channel_OutputChannelImpl_nativeFinalize
func Java_io_v_impl_google_channel_OutputChannelImpl_nativeFinalize(jenv *C.JNIEnv, jOutputChannelClass C.jclass, goConvertPtr C.jlong, goSendPtr C.jlong, goClosePtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goConvertPtr))
	jutil.GoUnref(jutil.NativePtr(goSendPtr))
	jutil.GoUnref(jutil.NativePtr(goClosePtr))
}
