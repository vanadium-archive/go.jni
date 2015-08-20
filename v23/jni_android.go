// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package v23

import (
	"syscall"
)

// #include "jni.h"
import "C"

//export Java_io_v_android_v23_RedirectStderr_nativeStart
func Java_io_v_android_v23_RedirectStderr_nativeStart(jenv *C.JNIEnv, jRuntime C.jclass, fileno C.jint) {
	syscall.Dup2(int(fileno), syscall.Stderr)
}
