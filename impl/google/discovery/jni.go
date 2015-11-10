// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package discovery

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	"v.io/v23/discovery"
	"v.io/v23/security"
	idiscovery "v.io/x/ref/lib/discovery"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
// #include <stdlib.h>
import "C"

var (
	updateSign = jutil.ClassSign("io.v.v23.discovery.Update")


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


//export Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_nativeFinalize
func Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_nativeFinalize(jenv *C.JNIEnv, _ C.jobject, discovery C.jlong, trigger C.jlong) {
	jutil.GoUnref(jutil.NativePtr(discovery))
	jutil.GoUnref(jutil.NativePtr(trigger))
}

//export Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_advertise
func Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_advertise(jenv *C.JNIEnv, jDiscoveryObj C.jobject, jContext C.jobject, jServiceObject C.jobject, jPerms C.jobject, jCallback C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	ctx, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		return
	}
	jService := jutil.Object(uintptr(unsafe.Pointer(jServiceObject)))
	var service discovery.Service
	if err := jutil.GoVomCopy(env, jService, jServiceClass, &service); err != nil {
		jutil.JThrowV(env, err)
		return
	}

	permsArray, err := jutil.GoObjectList(env, jutil.Object(uintptr(unsafe.Pointer(jPerms))))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	permsSlice := make([]security.BlessingPattern, len(permsArray))
	for i, jPattern := range permsArray {
		if err := jutil.GoVomCopy(env, jPattern, jBlessingPatternClass, &permsSlice[i]); err != nil {
			jutil.JThrowV(env, err)
			return
		}
	}
	jDiscovery := jutil.Object(uintptr(unsafe.Pointer(jDiscoveryObj)))
	discoveryPtr, err := jutil.JLongField(env, jDiscovery, "nativeDiscovery")
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}

	ds := *(*discovery.T)(jutil.NativePtr(discoveryPtr))

	triggerPtr, err := jutil.JLongField(env, jDiscovery, "nativeTrigger")

	if err != nil {
		jutil.JThrowV(env, err)
		return
	}

	trigger := (*idiscovery.Trigger)(jutil.NativePtr(triggerPtr))

	done, err := ds.Advertise(ctx, service, permsSlice)

	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	cb := jutil.Object(uintptr(unsafe.Pointer(jCallback)))
	cb = jutil.NewGlobalRef(env, cb)
	trigger.Add(func() {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		// TODO(bjornick): What should we do on errors?
		if err := jutil.CallVoidMethod(env, cb, "done", nil); err != nil {
			ctx.Errorf("failed to call done:", err)
		}
		jutil.DeleteGlobalRef(env, cb)
	}, done)
}

//export Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_scan
func Java_io_v_impl_google_lib_discovery_VDiscoveryImpl_scan(jenv *C.JNIEnv, jDiscoveryObj C.jobject, jContext C.jobject, jQuery C.jstring, jCallback C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	ctx, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		return
	}
	query := jutil.GoString(env, jutil.Object(uintptr(unsafe.Pointer(jQuery))))

	jDiscovery := jutil.Object(uintptr(unsafe.Pointer(jDiscoveryObj)))
	discoveryPtr, err := jutil.JLongField(env, jDiscovery, "nativeDiscovery")
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	ds := *(*discovery.T)(jutil.NativePtr(discoveryPtr))

	updates, err := ds.Scan(ctx, query)

	jutil.NewGlobalRef(env, jutil.Object(uintptr(unsafe.Pointer(jCallback))))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	cb := jutil.Object(uintptr(unsafe.Pointer(jCallback)))
	cb = jutil.NewGlobalRef(env, cb)
	go func() {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		for v := range updates {
			jUpdate, err := jutil.JVomCopy(env, v, jUpdateClass)
			if err != nil {
				ctx.Errorf("Failed to convert update: %v", err)
				continue
			}
			err = jutil.CallVoidMethod(env, cb, "handleUpdate", []jutil.Sign{updateSign}, jUpdate)
			if err != nil {
				ctx.Errorf("Failed to call handler: %v", err)
				continue
			}
		}
		jutil.DeleteGlobalRef(env, cb)
	}()
}
