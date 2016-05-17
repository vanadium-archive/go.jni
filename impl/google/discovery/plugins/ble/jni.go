// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package ble

import (
	"unsafe"

	"v.io/x/ref/lib/discovery/plugins/ble"

	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

var (
	contextSign     = jutil.ClassSign("io.v.v23.context.VContext")
	scanHandlerSign = jutil.ClassSign("io.v.android.impl.google.discovery.plugins.ble.Driver$ScanHandler")

	jNativeScanHandlerClass jutil.Class // io.v.android.impl.google.discovery.plugins.ble.NativeScanHandler
)

func Init(env jutil.Env) error {
	var err error
	jNativeScanHandlerClass, err = jutil.JFindClass(env, "io/v/android/impl/google/discovery/plugins/ble/NativeScanHandler")
	if err != nil {
		return err
	}
	return initDriverFactory(env)
}

//export Java_io_v_android_impl_google_discovery_plugins_ble_NativeScanHandler_nativeOnDiscovered
func Java_io_v_android_impl_google_discovery_plugins_ble_NativeScanHandler_nativeOnDiscovered(jenv *C.JNIEnv, _ C.jobject, handlerRef C.jlong, jUuid C.jstring, jCharacteristics C.jobject, jRssi C.jint) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	scanHandler := (*(*ble.ScanHandler)(jutil.GoRefValue(jutil.Ref(handlerRef))))
	uuid := jutil.GoString(env, jutil.Object(uintptr(unsafe.Pointer(jUuid))))
	csObjMap, err := jutil.GoObjectMap(env, jutil.Object(uintptr(unsafe.Pointer(jCharacteristics))))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	characteristics := make(map[string][]byte, len(csObjMap))
	for jUuid, jCharacteristic := range csObjMap {
		characteristics[jutil.GoString(env, jUuid)] = jutil.GoByteArray(env, jCharacteristic)
	}
	rssi := int(jRssi)
	scanHandler.OnDiscovered(uuid, characteristics, rssi)
}

//export Java_io_v_android_impl_google_discovery_plugins_ble_NativeScanHandler_nativeFinalize
func Java_io_v_android_impl_google_discovery_plugins_ble_NativeScanHandler_nativeFinalize(_ *C.JNIEnv, _ C.jobject, handlerRef C.jlong) {
	jutil.GoDecRef(jutil.Ref(handlerRef))
}
