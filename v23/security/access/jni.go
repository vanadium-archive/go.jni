// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package access

import (
	"v.io/v23/security/access"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	jsecurity "v.io/x/jni/v23/security"
)

// #include "jni.h"
import "C"

var (
	// Global reference for io.v.v23.security.access.AccessList class.
	jAccessListClass C.jclass
)

func Init(jEnv interface{}) error {
	class, err := jutil.JFindClass(jEnv, "io/v/v23/security/access/AccessList")
	if err != nil {
		return err
	}
	jAccessListClass = C.jclass(class)
	return nil
}

//export Java_io_v_v23_security_access_AccessList_nativeCreate
func Java_io_v_v23_security_access_AccessList_nativeCreate(env *C.JNIEnv, jAccessList C.jobject) C.jlong {
	acl, err := GoAccessList(env, jAccessList)
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	jutil.GoRef(&acl) // Un-refed when the AccessList object is finalized
	return C.jlong(jutil.PtrValue(&acl))
}

//export Java_io_v_v23_security_access_AccessList_nativeIncludes
func Java_io_v_v23_security_access_AccessList_nativeIncludes(env *C.JNIEnv, jAccessList C.jobject, goPtr C.jlong, jBlessings C.jobjectArray) C.jboolean {
	blessings := jutil.GoStringArray(env, jBlessings)
	ok := (*(*access.AccessList)(jutil.Ptr(goPtr))).Includes(blessings...)
	if ok {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_v23_security_access_AccessList_nativeAuthorize
func Java_io_v_v23_security_access_AccessList_nativeAuthorize(env *C.JNIEnv, jAccessList C.jobject, goPtr C.jlong, jCtx C.jobject, jCall C.jobject) {
	ctx, err := jcontext.GoContext(env, jCtx)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	call, err := jsecurity.GoCall(env, jCall)
	if err != nil {
		jutil.JThrowV(env, err)
	}
	if err := (*(*access.AccessList)(jutil.Ptr(goPtr))).Authorize(ctx, call); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_v23_security_access_AccessList_nativeFinalize
func Java_io_v_v23_security_access_AccessList_nativeFinalize(env *C.JNIEnv, jAccessList C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goPtr))
}
