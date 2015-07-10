// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"fmt"
	"runtime"

	"v.io/v23/context"
	"v.io/v23/rpc"
	"v.io/v23/security"

	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

func goDispatcher(env *C.JNIEnv, jDispatcher C.jobject) (*dispatcher, error) {
	// Reference Java dispatcher; it will be de-referenced when the go
	// dispatcher created below is garbage-collected (through the finalizer
	// callback we setup below).
	jDispatcher = C.jobject(jutil.NewGlobalRef(env, jDispatcher))
	d := &dispatcher{
		jDispatcher: jDispatcher,
	}
	runtime.SetFinalizer(d, func(d *dispatcher) {
		jEnv, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		jutil.DeleteGlobalRef(env, d.jDispatcher)
	})

	return d, nil
}

type dispatcher struct {
	jDispatcher C.jobject
}

func (d *dispatcher) Lookup(ctx *context.T, suffix string) (interface{}, security.Authorizer, error) {
	// Get Java environment.
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	dispatcherSign := jutil.ClassSign("io.v.v23.rpc.Dispatcher")
	result, err := jutil.CallStaticLongArrayMethod(env, jUtilClass, "lookup", []jutil.Sign{dispatcherSign, jutil.StringSign}, d.jDispatcher, suffix)
	if err != nil {
		return nil, nil, fmt.Errorf("error invoking Java dispatcher's lookup() method: %v", err)
	}
	if result == nil {
		// Lookup returned null, which means that the dispatcher isn't handling the object -
		// this is not an error.
		return nil, nil, nil
	}
	if len(result) != 2 {
		return nil, nil, fmt.Errorf("lookup returned %d elems, want 2", len(result))
	}
	invoker := *(*rpc.Invoker)(jutil.Ptr(result[0]))
	jutil.GoUnref(jutil.Ptr(result[0]))
	authorizer := security.Authorizer(nil)
	if result[1] != 0 {
		authorizer = *(*security.Authorizer)(jutil.Ptr(result[1]))
		jutil.GoUnref(jutil.Ptr(result[1]))
	}
	return invoker, authorizer, nil
}
