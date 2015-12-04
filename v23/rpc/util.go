// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	jutil "v.io/x/jni/util"
)

// JavaNativeCallback creates a new Java Callback object that calls the provided Go functions
// on success/failures.
func JavaNativeCallback(env jutil.Env, success func(jResult jutil.Object), failure func(err error)) (jutil.Object, error) {
	jCallback, err := jutil.NewObject(env, jNativeCallbackClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, int64(jutil.PtrValue(&success)), int64(jutil.PtrValue(&failure)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&success)  // Un-refed when jCallback is finalized
	jutil.GoRef(&failure)  // Un-refed when jCallback is finalized
	return jCallback, nil
}