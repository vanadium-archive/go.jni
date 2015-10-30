// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package discovery

import (
	"bytes"
	"encoding/binary"
	"unsafe"

	jutil "v.io/x/jni/util"
	"v.io/x/ref/lib/discovery"
)

// #include "jni.h"
// #include <stdlib.h>
import "C"

var (
	classSign = jutil.ClassSign("java.lang.Class")
	// Global reference for java.util.UUID class.
	jUUIDClass jutil.Class
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
func Init(env jutil.Env) error {
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	var err error
	jUUIDClass, err = jutil.JFindClass(env, "java/util/UUID")

	return err
}

func convertStringtoUUID(jenv *C.JNIEnv, _ C.jclass, jName C.jstring, generator func(string) discovery.Uuid) C.jobject {
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
	return convertStringtoUUID(jenv, jclass, jName, discovery.NewServiceUUID)
}

//export Java_io_v_impl_google_lib_discovery_UUIDUtil_UUIDForAttributeKey
func Java_io_v_impl_google_lib_discovery_UUIDUtil_UUIDForAttributeKey(jenv *C.JNIEnv, jclass C.jclass, jName C.jstring) C.jobject {
	converter := func(s string) discovery.Uuid {
		return discovery.NewAttributeUUID(s)
	}
	return convertStringtoUUID(jenv, jclass, jName, converter)
}
