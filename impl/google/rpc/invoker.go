// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"fmt"
	"runtime"
	"unsafe"

	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/rpc"
	"v.io/v23/vdl"
	"v.io/v23/vdlroot/signature"

	jchannel "v.io/x/jni/impl/google/channel"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

func goInvoker(env *C.JNIEnv, jObj C.jobject) (rpc.Invoker, error) {
	// See if the Java object is an invoker.
	var jInvoker C.jobject
	if jutil.IsInstanceOf(env, jObj, jInvokerClass) {
		jInvoker = jObj
	} else {
		// Create a new Java ReflectInvoker object.
		jReflectInvoker, err := jutil.NewObject(env, jReflectInvokerClass, []jutil.Sign{jutil.ObjectSign}, jObj)
		if err != nil {
			return nil, fmt.Errorf("error creating Java ReflectInvoker object: %v", err)
		}
		jInvoker = C.jobject(jReflectInvoker)
	}

	// Reference Java invoker; it will be de-referenced when the go invoker
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jInvokerRef := C.jobject(jutil.NewGlobalRef(env, jInvoker))
	i := &invoker{
		jInvoker: jInvokerRef,
	}
	runtime.SetFinalizer(i, func(i *invoker) {
		jEnv, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(jEnv)
		defer freeFunc()
		jutil.DeleteGlobalRef(env, i.jInvoker)
	})
	return i, nil
}

type invoker struct {
	jInvoker C.jobject
}

func (i *invoker) Prepare(method string, numArgs int) (argptrs []interface{}, tags []*vdl.Value, err error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	// Have all input arguments be decoded into *vdl.Value.
	argptrs = make([]interface{}, numArgs)
	for i := 0; i < numArgs; i++ {
		value := new(vdl.Value)
		argptrs[i] = &value
	}
	// Get the method tags.
	jTags, err := jutil.CallObjectMethod(env, i.jInvoker, "getMethodTags", []jutil.Sign{jutil.StringSign}, jutil.ArraySign(jutil.VdlValueSign), jutil.CamelCase(method))
	if err != nil {
		return nil, nil, err
	}
	tags, err = jutil.GoVDLValueArray(env, jTags)
	if err != nil {
		return nil, nil, err
	}
	return
}

func (i *invoker) Invoke(ctx *context.T, call rpc.StreamServerCall, method string, argptrs []interface{}) (results []interface{}, err error) {
	jEnv, freeFunc := jutil.GetEnv()
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		return nil, err
	}
	jStreamServerCall, err := javaStreamServerCall(env, call)
	if err != nil {
		return nil, err
	}
	// Convert Go arguments to Java.
	jArgs, err := i.prepareArgs(env, method, argptrs)
	if err != nil {
		return nil, err
	}
	// Invoke the method.
	resultarr, err := jutil.CallObjectArrayMethod(env, i.jInvoker, "invoke", []jutil.Sign{contextSign, streamServerCallSign, jutil.StringSign, jutil.ArraySign(jutil.ObjectSign)}, jutil.ObjectSign, jContext, jStreamServerCall, jutil.CamelCase(method), jArgs)
	if err != nil {
		return nil, err
	}
	// Convert Java results into Go.
	return i.prepareResults(env, method, resultarr)
}

func (i *invoker) Signature(ctx *context.T, call rpc.ServerCall) ([]signature.Interface, error) {
	jEnv, freeFunc := jutil.GetEnv()
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		return nil, err
	}
	jServerCall, err := JavaServerCall(env, call)
	if err != nil {
		return nil, err
	}
	ifaceSign := jutil.ClassSign("io.v.v23.vdlroot.signature.Interface")
	interfacesArr, err := jutil.CallObjectArrayMethod(env, i.jInvoker, "getSignature", []jutil.Sign{contextSign, serverCallSign}, ifaceSign, jContext, jServerCall)
	if err != nil {
		return nil, err
	}

	result := make([]signature.Interface, len(interfacesArr))
	for i, jInterface := range interfacesArr {
		err = jutil.GoVomCopy(jEnv, jInterface, jInterfaceClass, &result[i])
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (i *invoker) MethodSignature(ctx *context.T, call rpc.ServerCall, method string) (signature.Method, error) {
	jEnv, freeFunc := jutil.GetEnv()
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		return signature.Method{}, err
	}
	jServerCall, err := JavaServerCall(env, call)
	if err != nil {
		return signature.Method{}, err
	}
	methodSign := jutil.ClassSign("io.v.v23.vdlroot.signature.Method")
	jMethod, err := jutil.CallObjectMethod(env, i.jInvoker, "getMethodSignature", []jutil.Sign{contextSign, serverCallSign, jutil.StringSign}, methodSign, jContext, jServerCall, method)
	if err != nil {
		return signature.Method{}, err
	}

	var result signature.Method
	err = jutil.GoVomCopy(jEnv, jMethod, jMethodClass, &result)
	if err != nil {
		return signature.Method{}, err
	}
	return result, nil
}

func (i *invoker) Globber() *rpc.GlobState {
	return &rpc.GlobState{AllGlobber: javaGlobber{i}}
}

// prepareArgs converts the provided arguments pointers into a Java Object array.
func (i *invoker) prepareArgs(env *C.JNIEnv, method string, argptrs []interface{}) (unsafe.Pointer, error) {
	args := make([]interface{}, len(argptrs))
	for i, argptr := range argptrs {
		args[i] = interface{}(jutil.DerefOrDie(argptr))
	}
	// Get Java argument types.
	typesarr, err := jutil.CallObjectArrayMethod(env, i.jInvoker, "getArgumentTypes", []jutil.Sign{jutil.StringSign}, jutil.TypeSign, jutil.CamelCase(method))
	if err != nil {
		return nil, err
	}
	if len(typesarr) != len(args) {
		return nil, fmt.Errorf("wrong number of arguments for method %s, want %d, have %d", method, len(typesarr), len(args))
	}
	arr := make([]interface{}, len(args))
	for i, arg := range args {
		var err error
		if arr[i], err = jutil.JVomCopy(env, arg, typesarr[i]); err != nil {
			return nil, err
		}
	}
	return jutil.JObjectArray(env, arr, jObjectClass), nil
}

// prepareResults converts the provided Java result array into a Go slice of *vdl.Value.
func (i *invoker) prepareResults(env *C.JNIEnv, method string, jResults []unsafe.Pointer) ([]interface{}, error) {
	// Get Java result types.
	typesarr, err := jutil.CallObjectArrayMethod(env, i.jInvoker, "getResultTypes", []jutil.Sign{jutil.StringSign}, jutil.TypeSign, jutil.CamelCase(method))
	if err != nil {
		return nil, err
	}
	if len(typesarr) != len(jResults) {
		return nil, fmt.Errorf("wrong number of results for method %s, want %d, have %d", method, len(typesarr), len(jResults))
	}
	// VOM-encode Java results and decode into []*vdl.Value.
	ret := make([]interface{}, len(jResults))
	for i, jResult := range jResults {
		data, err := jutil.JVomEncode(env, jResult, typesarr[i])
		if err != nil {
			return nil, err
		}
		if ret[i], err = jutil.VomDecodeToValue(data); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

type javaGlobber struct {
	i *invoker
}

func (j javaGlobber) Glob__(ctx *context.T, call rpc.ServerCall, pattern string) (<-chan naming.GlobReply, error) {
	jEnv, freeFunc := jutil.GetEnv()
	defer freeFunc()
	env := (*C.JNIEnv)(jEnv)

	jServerCall, err := JavaServerCall(env, call)
	if err != nil {
		return nil, err
	}

	actualChannel := make(chan naming.GlobReply)
	readFunc := func(input interface{}) error {
		jEnv, freeFunc := jutil.GetEnv()
		defer freeFunc()
		env := (*C.JNIEnv)(jEnv)

		defer jutil.DeleteGlobalRef(env, input)
		var reply naming.GlobReply
		err := jutil.GoVomCopy(env, input, jGlobReplyClass, &reply)
		if err != nil {
			return err
		}
		actualChannel <- reply
		return nil
	}
	closeFunc := func() error {
		close(actualChannel)
		return nil
	}

	jOutputChannel, err := jchannel.JavaOutputChannel(env, readFunc, closeFunc)
	if err != nil {
		return nil, err
	}

	callSign := jutil.ClassSign("io.v.v23.rpc.ServerCall")
	channelSign := jutil.ClassSign("io.v.v23.OutputChannel")
	// Calls Java invoker's glob method.
	go func(jServerCallRef C.jobject, jOutputChannelRef C.jobject) {
		jEnv, freeFunc := jutil.GetEnv()
		defer freeFunc()
		env := (*C.JNIEnv)(jEnv)
		jutil.CallVoidMethod(env, j.i.jInvoker, "glob", []jutil.Sign{callSign, jutil.StringSign, channelSign}, jServerCallRef, pattern, jOutputChannelRef)
		jutil.DeleteGlobalRef(env, jServerCallRef)
		jutil.DeleteGlobalRef(env, jOutputChannelRef)
	}(C.jobject(jutil.NewGlobalRef(env, jServerCall)), C.jobject(jutil.NewGlobalRef(env, jOutputChannel)))

	return actualChannel, nil
}
