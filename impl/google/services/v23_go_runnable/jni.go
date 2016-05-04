// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package v23_go_runnable

import (
	"unsafe"
	"v.io/v23/context"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

// V23GoRunnableFunc is the go code that is run by java/android callers of
// V23GoRunnable.run().
// Users must edit this function and rebuild the java lib/android-lib.
// Then android apps may run this Go code.
func V23GoRunnableFunc(ctx *context.T) error {
	ctx.Infof("Running V23GoRunnableFunc")
	return nil
}

//export Java_io_v_android_util_V23GoRunnable_nativeGoContextCall
func Java_io_v_android_util_V23GoRunnable_nativeGoContextCall(jenv *C.JNIEnv, jV23GoRunner C.jobject, jContext C.jobject) {
	env := jutil.Env(uintptr(unsafe.Pointer(jenv)))
	ctx, _, err := jcontext.GoContext(env, jutil.Object(uintptr(unsafe.Pointer(jContext))))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := V23GoRunnableFunc(ctx); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}
