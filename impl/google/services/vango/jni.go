// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package vango

import (
	"fmt"
	"unsafe"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

//export Java_io_v_android_util_Vango_nativeGoContextCall
func Java_io_v_android_util_Vango_nativeGoContextCall(jenv *C.JNIEnv, jVango C.jobject, jContext C.jobject, jKey C.jstring) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	key := jutil.GoString(env, jutil.Object(uintptr(unsafe.Pointer(jKey))))
	ctx, _, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	f, ok := vangoFuncs[key]
	if !ok {
		jutil.JThrowV(env, fmt.Errorf("vangoFunc key %q doesn't exist", key))
		return
	}
	if err := f(ctx); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}
