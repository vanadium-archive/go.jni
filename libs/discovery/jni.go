// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package discovery

import (
	"unsafe"

	jutil "v.io/x/jni/util"
	"v.io/x/ref/lib/discovery"
)

// #include "jni.h"
import "C"

var (
	androidContextSign = jutil.ClassSign("android.content.Context")
	contextSign        = jutil.ClassSign("io.v.v23.context.VContext")
	advertisementSign  = jutil.ClassSign("io.v.x.ref.lib.discovery.Advertisement")
	uuidSign           = jutil.ClassSign("java.util.UUID")
	scanHandlerSign    = jutil.ClassSign("io.v.impl.google.lib.discovery.ScanHandler")

	// Global reference for io.v.android.libs.discovery.ble.BlePlugin
	jBlePluginClass jutil.Class
	// Global reference for io.v.x.ref.lib.discovery.Advertisement
	jAdvertisementClass jutil.Class
	// Global reference for java.util.UUID
	jUUIDClass jutil.Class
	// Global reference for io.v.android.libs.discovery.ble.NativeScanHandler
	jNativeScanHandlerClass jutil.Class
)

func Init(env jutil.Env) error {
	var err error

	jUUIDClass, err = jutil.JFindClass(env, "java/util/UUID")
	if err != nil {
		return err
	}

	jNativeScanHandlerClass, err = jutil.JFindClass(env, "io/v/android/libs/discovery/ble/NativeScanHandler")
	if err != nil {
		return err
	}

	jBlePluginClass, err = jutil.JFindClass(env, "io/v/android/libs/discovery/ble/BlePlugin")
	if err != nil {
		return err
	}

	jAdvertisementClass, err = jutil.JFindClass(env, "io/v/x/ref/lib/discovery/Advertisement")
	return err
}

//export Java_io_v_android_libs_discovery_ble_NativeScanHandler_nativeHandleUpdate
func Java_io_v_android_libs_discovery_ble_NativeScanHandler_nativeHandleUpdate(jenv *C.JNIEnv, _ C.jobject, jAdvObj C.jobject, goPtr C.jlong) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	ch := (*(*chan<- discovery.Advertisement)(jutil.NativePtr(goPtr)))

	jAdv := jutil.Object(uintptr(unsafe.Pointer(jAdvObj)))
	var adv discovery.Advertisement
	if err := jutil.GoVomCopy(env, jAdv, jAdvertisementClass, &adv); err != nil {
		jutil.JThrowV(env, err)
		return
	}
	ch <- adv
}

//export Java_io_v_android_libs_discovery_ble_NativeScanHandler_nativeFinalize
func Java_io_v_android_libs_discovery_ble_NativeScanHandler_nativeFinalize(jenv *C.JNIEnv, _ C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}
