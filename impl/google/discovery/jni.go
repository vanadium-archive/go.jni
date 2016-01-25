// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package discovery

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"v.io/v23/context"
	"v.io/v23/discovery"
	"v.io/v23/security"
	"v.io/v23/verror"
	idiscovery "v.io/x/ref/lib/discovery"

	jchannel "v.io/x/jni/impl/google/channel"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
// #include <stdlib.h>
import "C"

var (
	updateSign  = jutil.ClassSign("io.v.v23.discovery.Update")
	contextSign = jutil.ClassSign("io.v.v23.context.VContext")

	// Global reference for java.util.UUID class.
	jUUIDClass jutil.Class

	// Global reference io.v.impl.google.lib.discovery.ScanHandler
	jScanHandlerClass jutil.Class

	// Global reference io.v.v23.discovery.Service
	jServiceClass jutil.Class

	// Global reference io.v.v23.security.BlessingPattern
	jBlessingPatternClass jutil.Class

	// Global reference io.v.v23.discovery.Update
	jUpdateClass jutil.Class

	// Global reference io.v.impl.google.lib.discovery.VDiscoveryImpl
	jVDiscoveryImplClass jutil.Class
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
func Init(env jutil.Env) error {
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	var err error
	jUUIDClass, err = jutil.JFindClass(env, "java/util/UUID")
	if err != nil {
		return err
	}
	jScanHandlerClass, err = jutil.JFindClass(env, "io/v/impl/google/lib/discovery/ScanHandler")
	if err != nil {
		return err
	}
	jServiceClass, err = jutil.JFindClass(env, "io/v/v23/discovery/Service")
	if err != nil {
		return err
	}
	jBlessingPatternClass, err = jutil.JFindClass(env, "io/v/v23/security/BlessingPattern")
	if err != nil {
		return err
	}
	jUpdateClass, err = jutil.JFindClass(env, "io/v/v23/discovery/Update")
	if err != nil {
		return err
	}

	jVDiscoveryImplClass, err = jutil.JFindClass(env, "io/v/impl/google/lib/discovery/VDiscoveryImpl")
	if err != nil {
		return err
	}

	return nil
}

func convertStringtoUUID(jenv *C.JNIEnv, _ C.jclass, jName C.jstring, generator func(string) idiscovery.Uuid) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	name := jutil.GoString(env, jutil.Object(uintptr(unsafe.Pointer(jName))))
	uuid := generator(name)
	buf := bytes.NewBuffer(uuid)
	var high, low int64
	binary.Read(buf, binary.BigEndian, &high)
	binary.Read(buf, binary.BigEndian, &low)
	jUUID, err := jutil.NewObject(env, jUUIDClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, high, low)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jUUID))
}

//export Java_io_v_impl_google_lib_discovery_UUIDUtil_UUIDForInterfaceName
func Java_io_v_impl_google_lib_discovery_UUIDUtil_UUIDForInterfaceName(jenv *C.JNIEnv, jclass C.jclass, jName C.jstring) C.jobject {
	return convertStringtoUUID(jenv, jclass, jName, idiscovery.NewServiceUUID)
}

//export Java_io_v_impl_google_lib_discovery_UUIDUtil_UUIDForAttributeKey
func Java_io_v_impl_google_lib_discovery_UUIDUtil_UUIDForAttributeKey(jenv *C.JNIEnv, jclass C.jclass, jName C.jstring) C.jobject {
	converter := func(s string) idiscovery.Uuid {
		return idiscovery.NewAttributeUUID(s)
	}
	return convertStringtoUUID(jenv, jclass, jName, converter)
}

func doAdvertise(ctx *context.T, ds discovery.T, trigger *idiscovery.Trigger, service discovery.Service, visibility []security.BlessingPattern, jService jutil.Object, jDoneCallback jutil.Object) (jutil.Object, error) {
	// Blocking call, so don't call GetEnv() beforehand.
	done, err := ds.Advertise(ctx, &service, visibility)
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	defer jutil.DeleteGlobalRef(env, jService)
	if err != nil {
		jutil.DeleteGlobalRef(env, jDoneCallback)
		return jutil.NullObject, err
	}
	// Copy back service.InstanceId to jService since it's the only field that would be updated.
	if err = jutil.CallVoidMethod(env, jService, "setInstanceId", []jutil.Sign{jutil.StringSign}, service.InstanceId); err != nil {
		jutil.DeleteGlobalRef(env, jDoneCallback)
		return jutil.NullObject, err
	}
	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		return jutil.NullObject, err
	}
	listenableFutureSign := jutil.ClassSign("com.google.common.util.concurrent.ListenableFuture")
	jDoneFuture, err := jutil.CallObjectMethod(env, jDoneCallback, "getFuture", []jutil.Sign{contextSign}, listenableFutureSign, jContext)
	if err != nil {
		jutil.DeleteGlobalRef(env, jDoneCallback)
		return jutil.NullObject, err
	}
	trigger.Add(func() {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.CallbackOnSuccess(env, jDoneCallback, jutil.NullObject)
		jutil.DeleteGlobalRef(env, jDoneCallback)
	}, done)
	// Must grab a global reference as we free up the env and all local references that come along
	// with it.
	return jutil.NewGlobalRef(env, jDoneFuture), nil // Un-refed in DoAsyncCall
}

//export Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_nativeAdvertise
func Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_nativeAdvertise(jenv *C.JNIEnv, jDiscovery C.jobject, goDiscoveryPtr C.jlong, goTriggerPtr C.jlong, jContext C.jobject, jServiceObj C.jobject, jVisibilityObj C.jobject, jStartCallbackObj C.jobject, jDoneCallbackObj C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	jService := jutil.Object(uintptr(unsafe.Pointer(jServiceObj)))
	jVisibility := jutil.Object(uintptr(unsafe.Pointer(jVisibilityObj)))
	jStartCallback := jutil.Object(uintptr(unsafe.Pointer(jStartCallbackObj)))
	jDoneCallback := jutil.Object(uintptr(unsafe.Pointer(jDoneCallbackObj)))
	ctx, _, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		jutil.CallbackOnFailure(env, jStartCallback, err)
		return
	}
	var service discovery.Service
	if err := jutil.GoVomCopy(env, jService, jServiceClass, &service); err != nil {
		jutil.CallbackOnFailure(env, jStartCallback, err)
		return
	}
	varr, err := jutil.GoObjectList(env, jVisibility)
	if err != nil {
		jutil.CallbackOnFailure(env, jStartCallback, err)
		return
	}
	visibility := make([]security.BlessingPattern, len(varr))
	for i, jPattern := range varr {
		if err := jutil.GoVomCopy(env, jPattern, jBlessingPatternClass, &visibility[i]); err != nil {
			jutil.CallbackOnFailure(env, jStartCallback, err)
			return
		}
	}
	ds := *(*discovery.T)(jutil.NativePtr(goDiscoveryPtr))
	trigger := (*idiscovery.Trigger)(jutil.NativePtr(goTriggerPtr))
	jService = jutil.NewGlobalRef(env, jService)           // Un-refed in doAdvertise
	jDoneCallback = jutil.NewGlobalRef(env, jDoneCallback) // Un-refed in doAdvertise
	jutil.DoAsyncCall(env, jStartCallback, func() (jutil.Object, error) {
		return doAdvertise(ctx, ds, trigger, service, visibility, jService, jDoneCallback)
	})
}

//export Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_nativeScan
func Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_nativeScan(jenv *C.JNIEnv, jDiscovery C.jobject, goDiscoveryPtr C.jlong, jContext C.jobject, jQuery C.jstring) C.jobject {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	ctx, cancel, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	query := jutil.GoString(env, jutil.Object(uintptr(unsafe.Pointer(jQuery))))
	ds := *(*discovery.T)(jutil.NativePtr(goDiscoveryPtr))

	var scanChannel <-chan discovery.Update
	var scanError error
	scanDone := make(chan bool)
	go func() {
		scanChannel, scanError = ds.Scan(ctx, query)
		close(scanDone)
	}()
	jChannel, err := jchannel.JavaInputChannel(env, ctx, cancel, func() (jutil.Object, error) {
		// A few blocking calls below - don't call GetEnv() before they complete.
		<-scanDone
		if scanError != nil {
			return jutil.NullObject, scanError
		}
		update, ok := <-scanChannel
		if !ok {
			return jutil.NullObject, verror.NewErrEndOfFile(ctx)
		}
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jUpdate, err := jutil.JVomCopy(env, update, jUpdateClass)
		if err != nil {
			return jutil.NullObject, err
		}
		// Must grab a global reference as we free up the env and all local references that come
		// along with it.
		return jutil.NewGlobalRef(env, jUpdate), nil // Un-refed by InputChannelImpl_nativeRecv
	})
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jChannel))
}

//export Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_nativeFinalize
func Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_nativeFinalize(jenv *C.JNIEnv, jDiscovery C.jobject, goDiscoveryPtr C.jlong, goTriggerPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goDiscoveryPtr))
	jutil.GoUnref(jutil.NativePtr(goTriggerPtr))
}
