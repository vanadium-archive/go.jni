// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package rpc

import (
	"fmt"
	"runtime"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
	jsecurity "v.io/x/jni/v23/security"
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

func (d *dispatcher) Lookup(suffix string) (interface{}, security.Authorizer, error) {
	// Get Java environment.
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	// Call Java dispatcher's lookup() method.
	serviceObjectWithAuthorizerSign := jutil.ClassSign("io.v.v23.rpc.ServiceObjectWithAuthorizer")
	tempJObj, err := jutil.CallObjectMethod(env, d.jDispatcher, "lookup", []jutil.Sign{jutil.StringSign}, serviceObjectWithAuthorizerSign, suffix)
	jObj := C.jobject(tempJObj)
	if err != nil {
		return nil, nil, fmt.Errorf("error invoking Java dispatcher's lookup() method: %v", err)
	}
	if jObj == nil {
		// Lookup returned null, which means that the dispatcher isn't handling the object -
		// this is not an error.
		return nil, nil, nil
	}

	// Extract the Java service object and Authorizer.
	jServiceObj, err := jutil.CallObjectMethod(env, jObj, "getServiceObject", nil, jutil.ObjectSign)
	if err != nil {
		return nil, nil, err
	}
	if jServiceObj == nil {
		return nil, nil, fmt.Errorf("null service object returned by Java's ServiceObjectWithAuthorizer")
	}
	authSign := jutil.ClassSign("io.v.v23.security.Authorizer")
	jAuth, err := jutil.CallObjectMethod(env, jObj, "getAuthorizer", nil, authSign)
	if err != nil {
		return nil, nil, err
	}

	// Create Go Invoker and Authorizer.
	i, err := goInvoker((*C.JNIEnv)(env), C.jobject(jServiceObj))
	if err != nil {
		return nil, nil, err
	}
	a, err := jsecurity.GoAuthorizer(env, jAuth)
	if err != nil {
		return nil, nil, err
	}
	return i, a, nil
}
