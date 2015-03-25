// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package android

import "syscall"

// #include "jni.h"
// #include <stdlib.h>
import "C"

// Init initializes the JNI code with the given Java environment.  This method
// must be invoked before any other method in this package and must be called
// from the main Java thread (e.g., On_Load()).
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	return nil
}

//export Java_io_v_v23_android_RedirectStderr_nativeStart
func Java_io_v_v23_android_RedirectStderr_nativeStart(env *C.JNIEnv, jRuntime C.jclass, fileno C.jint) {
	syscall.Dup2(int(fileno), syscall.Stderr)
}
