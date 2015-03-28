// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package security

import (
	"runtime"

	"v.io/v23/context"
	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// GoAuthorizer converts the given Java authorizer into a Go authorizer.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoAuthorizer(jEnv, jAuthObj interface{}) (security.Authorizer, error) {
	if jutil.IsNull(jAuthObj) {
		return nil, nil
	}
	// Reference Java dispatcher; it will be de-referenced when the go
	// dispatcher created below is garbage-collected (through the finalizer
	// callback we setup below).
	jAuth := C.jobject(jutil.NewGlobalRef(jEnv, jAuthObj))
	a := &authorizer{
		jAuth: jAuth,
	}
	runtime.SetFinalizer(a, func(a *authorizer) {
		jEnv, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		jutil.DeleteGlobalRef(env, a.jAuth)
	})
	return a, nil
}

type authorizer struct {
	jAuth C.jobject
}

func (a *authorizer) Authorize(ctx *context.T) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jCtx, err := JavaContext(env, ctx, nil)
	if err != nil {
		return err
	}

	// Run Java Authorizer.
	contextSign := jutil.ClassSign("io.v.v23.context.VContext")
	return jutil.CallVoidMethod(env, a.jAuth, "authorize", []jutil.Sign{contextSign}, jCtx)
}
