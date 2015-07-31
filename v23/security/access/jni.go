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
	jAccessListClass jutil.Class
)

func Init(env jutil.Env) error {
	var err error
	jAccessListClass, err = jutil.JFindClass(env, "io/v/v23/security/access/AccessList")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_v23_security_access_AccessList_nativeCreate
func Java_io_v_v23_security_access_AccessList_nativeCreate(jenv *C.JNIEnv, jAccessList C.jobject) C.jlong {
	env := jutil.WrapEnv(jenv)
	acl, err := GoAccessList(env, jutil.WrapObject(jAccessList))
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	jutil.GoRef(&acl) // Un-refed when the AccessList object is finalized
	return C.jlong(jutil.PtrValue(&acl))
}

//export Java_io_v_v23_security_access_AccessList_nativeIncludes
func Java_io_v_v23_security_access_AccessList_nativeIncludes(jenv *C.JNIEnv, jAccessList C.jobject, goPtr C.jlong, jBlessings C.jobjectArray) C.jboolean {
	env := jutil.WrapEnv(jenv)
	blessings, err := jutil.GoStringArray(env, jutil.WrapObject(jBlessings))
	if err != nil {
		jutil.JThrowV(env, err)
		return C.JNI_FALSE
	}
	ok := (*(*access.AccessList)(jutil.NativePtr(goPtr))).Includes(blessings...)
	if ok {
		return C.JNI_TRUE
	}
	return C.JNI_FALSE
}

//export Java_io_v_v23_security_access_AccessList_nativeAuthorize
func Java_io_v_v23_security_access_AccessList_nativeAuthorize(jenv *C.JNIEnv, jAccessList C.jobject, goPtr C.jlong, jCtx C.jobject, jCall C.jobject) {
	env := jutil.WrapEnv(jenv)
	ctx, err := jcontext.GoContext(env, jutil.WrapObject(jCtx))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	call, err := jsecurity.GoCall(env, jutil.WrapObject(jCall))
	if err != nil {
		jutil.JThrowV(env, err)
	}
	if err := (*(*access.AccessList)(jutil.NativePtr(goPtr))).Authorize(ctx, call); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_v23_security_access_AccessList_nativeFinalize
func Java_io_v_v23_security_access_AccessList_nativeFinalize(jenv *C.JNIEnv, jAccessList C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}
