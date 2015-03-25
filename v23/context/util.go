// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package context

import (
	"fmt"
	"log"
	"runtime"
	"unsafe"

	"v.io/v23/context"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

type goContextKey string

type goContextValue struct {
	jObj C.jobject
}

// JavaContext converts the provided Go Context into a Java Context.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaContext(jEnv interface{}, ctx *context.T, cancel context.CancelFunc) (unsafe.Pointer, error) {
	cancelPtr := int64(0)
	if cancel != nil {
		cancelPtr = int64(jutil.PtrValue(&cancel))
	}
	jCtx, err := jutil.NewObject(jEnv, jVContextImplClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, int64(jutil.PtrValue(ctx)), cancelPtr)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(ctx) // Un-refed when the Java context object is finalized.
	if cancel != nil {
		jutil.GoRef(&cancel) // Un-refed when the Java context object is finalized.
	}
	return jCtx, err
}

// GoContext converts the provided Java Context into a Go context.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoContext(jEnv, jContext interface{}) (*context.T, error) {
	if jutil.IsNull(jContext) {
		return nil, nil
	}
	goCtxPtr, err := jutil.CallLongMethod(jEnv, jContext, "nativePtr", nil)
	if err != nil {
		return nil, err
	}
	return (*context.T)(jutil.Ptr(goCtxPtr)), nil
}

// JavaCountDownLatch creates a Java CountDownLatch object with an initial count
// of one that counts down (to zero) the moment the value is sent on the
// provided Go channel or if the channel gets closed.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaCountDownLatch(jEnv interface{}, c <-chan struct{}) (unsafe.Pointer, error) {
	if c == nil {
		return nil, nil
	}
	jLatchObj, err := jutil.NewObject(jEnv, jCountDownLatchClass, []jutil.Sign{jutil.IntSign}, int(1))
	if err != nil {
		return nil, err
	}
	// Reference Java CountDownLatch; it will be de-referenced when the goroutine below exits.
	jLatch := C.jobject(jutil.NewGlobalRef(jEnv, jLatchObj))
	go func() {
		<-c
		javaEnv, freeFunc := jutil.GetEnv()
		jenv := (*C.JNIEnv)(javaEnv)
		defer freeFunc()
		if err := jutil.CallVoidMethod(jenv, jLatch, "countDown", nil); err != nil {
			log.Printf("Error decrementing CountDownLatch: %v", err)
		}
		jutil.DeleteGlobalRef(jenv, jLatch)
	}()
	return unsafe.Pointer(jLatch), nil
}

// GoContextKey creates a Go Context key given the Java Context key.  The
// returned key guarantees that the two Java keys will be equal iff (1) they
// belong to the same class, and (2) they have the same hashCode().
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoContextKey(jEnv, jKey interface{}) (interface{}, error) {
	// Create a lookup key we use to map Java context keys to Go context keys.
	hashCode, err := jutil.CallIntMethod(jEnv, jKey, "hashCode", nil)
	if err != nil {
		return nil, err
	}
	jClass, err := jutil.CallObjectMethod(jEnv, jKey, "getClass", nil, classSign)
	if err != nil {
		return nil, err
	}
	className, err := jutil.CallStringMethod(jEnv, jClass, "getName", nil)
	if err != nil {
		return nil, err
	}
	return goContextKey(fmt.Sprintf("%s:%d", className, hashCode)), nil
}

// GoContextValue returns the Go Context value given the Java Context value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoContextValue(jEnv, jValueObj interface{}) (interface{}, error) {
	// Reference Java object; it will be de-referenced when the Go wrapper
	// object created below is garbage-collected (via the finalizer we setup
	// just below.)
	jValue := C.jobject(jutil.NewGlobalRef(jEnv, jValueObj))
	value := &goContextValue{
		jObj: jValue,
	}
	runtime.SetFinalizer(value, func(value *goContextValue) {
		javaEnv, freeFunc := jutil.GetEnv()
		jenv := (*C.JNIEnv)(javaEnv)
		defer freeFunc()
		jutil.DeleteGlobalRef(jenv, value.jObj)
	})
	return value, nil
}

// JavaContextValue returns the Java Context value given the Go Context value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaContextValue(jEnv interface{}, value interface{}) (unsafe.Pointer, error) {
	if value == nil {
		return nil, nil
	}
	val, ok := value.(*goContextValue)
	if !ok {
		return nil, fmt.Errorf("Invalid type %T for value %v, wanted goContextValue", value, value)
	}
	return unsafe.Pointer(val.jObj), nil
}
