// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package plugins

import (
	"unsafe"

	idiscovery "v.io/x/ref/lib/discovery"

	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

var (
	contextSign     = jutil.ClassSign("io.v.v23.context.VContext")
	adInfoSign      = jutil.ClassSign("io.v.x.ref.lib.discovery.AdInfo")
	scanHandlerSign = jutil.ClassSign("io.v.impl.google.lib.discovery.Plugin$ScanHandler")

	jAdInfoClass            jutil.Class // io.v.x.ref.lib.discovery.AdInfo
	jNativeScanHandlerClass jutil.Class // io.v.android.impl.google.discovery.plugins.NativeScanHandler
)

func Init(env jutil.Env) error {
	var err error
	jAdInfoClass, err = jutil.JFindClass(env, "io/v/x/ref/lib/discovery/AdInfo")
	if err != nil {
		return err
	}
	jNativeScanHandlerClass, err = jutil.JFindClass(env, "io/v/android/impl/google/discovery/plugins/NativeScanHandler")
	if err != nil {
		return err
	}

	return initPluginFactories(env)
}

//export Java_io_v_android_impl_google_discovery_plugins_NativeScanHandler_nativeHandleUpdate
func Java_io_v_android_impl_google_discovery_plugins_NativeScanHandler_nativeHandleUpdate(jenv *C.JNIEnv, _ C.jobject, chPtr C.jlong, jAdInfoObj C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	ch := (*(*chan<- *idiscovery.AdInfo)(jutil.NativePtr(chPtr)))

	jAdInfo := jutil.Object(uintptr(unsafe.Pointer(jAdInfoObj)))

	var adInfo idiscovery.AdInfo
	if err := jutil.GoVomCopy(env, jAdInfo, jAdInfoClass, &adInfo); err != nil {
		jutil.JThrowV(env, err)
		return
	}
	ch <- &adInfo
}

//export Java_io_v_android_impl_google_discovery_plugins_NativeScanHandler_nativeFinalize
func Java_io_v_android_impl_google_discovery_plugins_NativeScanHandler_nativeFinalize(jenv *C.JNIEnv, _ C.jobject, chPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(chPtr))
}
