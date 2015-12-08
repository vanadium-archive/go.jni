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

// GoDispatcher creates a new rpc.Dispatcher given the Java Dispatcher object and the Java
// thread executor on which to run server methods.
func GoDispatcher(env jutil.Env, jDispatcher jutil.Object, jExecutor jutil.Object) (rpc.Dispatcher, error) {
	// Reference Java dispatcher and executor; they will be de-referenced when the go
	// dispatcher created below is garbage-collected (through the finalizer
	// callback we setup below).
	jDispatcher = jutil.NewGlobalRef(env, jDispatcher)
	jExecutor = jutil.NewGlobalRef(env, jExecutor)
	d := &dispatcher{
		jDispatcher: jDispatcher,
		jExecutor:   jExecutor,
	}
	runtime.SetFinalizer(d, func(d *dispatcher) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, d.jDispatcher)
		jutil.DeleteGlobalRef(env, d.jExecutor)
	})

	return d, nil
}

type dispatcher struct {
	jDispatcher, jExecutor jutil.Object
}

func (d *dispatcher) Lookup(ctx *context.T, suffix string) (interface{}, security.Authorizer, error) {
	// Get Java environment.
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	result, err := jutil.CallStaticLongArrayMethod(env, jServerRPCHelperClass, "lookup", []jutil.Sign{dispatcherSign, jutil.StringSign, executorSign}, d.jDispatcher, suffix, d.jExecutor)
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
	invoker := *(*rpc.Invoker)(jutil.NativePtr(result[0]))
	jutil.GoUnref(jutil.NativePtr(result[0]))
	authorizer := security.Authorizer(nil)
	if result[1] != 0 {
		authorizer = *(*security.Authorizer)(jutil.NativePtr(result[1]))
		jutil.GoUnref(jutil.NativePtr(result[1]))
	}
	return invoker, authorizer, nil
}
