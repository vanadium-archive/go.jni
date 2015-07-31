// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package context

import (
	"fmt"
	"log"
	"runtime"

	"v.io/v23/context"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

type goContextKey string

type goContextValue struct {
	jObj jutil.Object
}

// JavaContext converts the provided Go Context into a Java Context.
func JavaContext(env jutil.Env, ctx *context.T, cancel context.CancelFunc) (jutil.Object, error) {
	cancelPtr := int64(0)
	if cancel != nil {
		cancelPtr = int64(jutil.PtrValue(&cancel))
	}
	jCtx, err := jutil.NewObject(env, jVContextImplClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, int64(jutil.PtrValue(ctx)), cancelPtr)
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(ctx) // Un-refed when the Java context object is finalized.
	if cancel != nil {
		jutil.GoRef(&cancel) // Un-refed when the Java context object is finalized.
	}
	return jCtx, err
}

// GoContext converts the provided Java Context into a Go context.
func GoContext(env jutil.Env, jContext jutil.Object) (*context.T, error) {
	if jContext.IsNull() {
		return nil, nil
	}
	goCtxPtr, err := jutil.CallLongMethod(env, jContext, "nativePtr", nil)
	if err != nil {
		return nil, err
	}
	return (*context.T)(jutil.NativePtr(goCtxPtr)), nil
}

// JavaCountDownLatch creates a Java CountDownLatch object with an initial count
// of one that counts down (to zero) the moment the value is sent on the
// provided Go channel or if the channel gets closed.
func JavaCountDownLatch(env jutil.Env, c <-chan struct{}) (jutil.Object, error) {
	if c == nil {
		return jutil.NullObject, nil
	}
	jLatch, err := jutil.NewObject(env, jCountDownLatchClass, []jutil.Sign{jutil.IntSign}, int(1))
	if err != nil {
		return jutil.NullObject, err
	}
	// Reference Java CountDownLatch; it will be de-referenced when the goroutine below exits.
	jLatchGlobal := jutil.NewGlobalRef(env, jLatch)
	go func() {
		<-c
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		if err := jutil.CallVoidMethod(env, jLatchGlobal, "countDown", nil); err != nil {
			log.Printf("Error decrementing CountDownLatch: %v", err)
		}
		jutil.DeleteGlobalRef(env, jLatchGlobal)
	}()
	return jLatch, nil
}

// GoContextKey creates a Go Context key given the Java Context key.  The
// returned key guarantees that the two Java keys will be equal iff (1) they
// belong to the same class, and (2) they have the same hashCode().
func GoContextKey(env jutil.Env, jKey jutil.Object) (interface{}, error) {
	// Create a lookup key we use to map Java context keys to Go context keys.
	hashCode, err := jutil.CallIntMethod(env, jKey, "hashCode", nil)
	if err != nil {
		return nil, err
	}
	jClass, err := jutil.CallObjectMethod(env, jKey, "getClass", nil, classSign)
	if err != nil {
		return nil, err
	}
	className, err := jutil.CallStringMethod(env, jClass, "getName", nil)
	if err != nil {
		return nil, err
	}
	return goContextKey(fmt.Sprintf("%s:%d", className, hashCode)), nil
}

// GoContextValue returns the Go Context value given the Java Context value.
func GoContextValue(env jutil.Env, jValue jutil.Object) (interface{}, error) {
	// Reference Java object; it will be de-referenced when the Go wrapper
	// object created below is garbage-collected (via the finalizer we setup
	// just below.)
	jValue = jutil.NewGlobalRef(env, jValue)
	value := &goContextValue{
		jObj: jValue,
	}
	runtime.SetFinalizer(value, func(value *goContextValue) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, value.jObj)
	})
	return value, nil
}

// JavaContextValue returns the Java Context value given the Go Context value.
func JavaContextValue(env jutil.Env, value interface{}) (jutil.Object, error) {
	if value == nil {
		return jutil.NullObject, nil
	}
	val, ok := value.(*goContextValue)
	if !ok {
		return jutil.NullObject, fmt.Errorf("Invalid type %T for value %v, wanted goContextValue", value, value)
	}
	return val.jObj, nil
}
