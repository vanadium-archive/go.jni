// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package rpc

import (
	"fmt"
	"runtime"

	"v.io/v23/naming"
	"v.io/v23/rpc"
	"v.io/v23/vdl"
	"v.io/v23/vdlroot/signature"
	"v.io/v23/vom"
	jchannel "v.io/x/jni/impl/google/channel"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

func goInvoker(env *C.JNIEnv, jObj C.jobject) (rpc.Invoker, error) {
	// Create a new Java VDLInvoker object.
	jInvokerObj, err := jutil.NewObject(env, jVDLInvokerClass, []jutil.Sign{jutil.ObjectSign}, jObj)
	if err != nil {
		return nil, fmt.Errorf("error creating Java VDLInvoker object: %v", err)
	}
	// Reference Java invoker; it will be de-referenced when the go invoker
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jInvoker := C.jobject(jutil.NewGlobalRef(env, jInvokerObj))
	i := &invoker{
		jInvoker: jInvoker,
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

func (i *invoker) Invoke(method string, call rpc.StreamServerCall, argptrs []interface{}) (results []interface{}, err error) {
	jEnv, freeFunc := jutil.GetEnv()
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	jStreamServerCall, err := javaStreamServerCall(env, call)
	if err != nil {
		return nil, err
	}

	// VOM-encode the input arguments.
	jVomArgs, err := encodeArgs(env, argptrs)
	if err != nil {
		return nil, err
	}
	// Invoke the method.
	callSign := jutil.ClassSign("io.v.v23.rpc.StreamServerCall")
	replySign := jutil.ClassSign("io.v.impl.google.rpc.VDLInvoker$InvokeReply")
	jReply, err := jutil.CallObjectMethod(env, i.jInvoker, "invoke", []jutil.Sign{jutil.StringSign, callSign, jutil.ArraySign(jutil.ArraySign(jutil.ByteSign))}, replySign, jutil.CamelCase(method), jStreamServerCall, jVomArgs)
	if err != nil {
		return nil, fmt.Errorf("error invoking Java method %q: %v", method, err)
	}
	// Decode and return results.
	return decodeResults(env, C.jobject(jReply))
}

type javaGlobber struct {
	i *invoker
}

func (j javaGlobber) Glob__(call rpc.ServerCall, pattern string) (<-chan naming.GlobReply, error) {
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
	// Invoke the VDLInvoker's glob method.
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

func (i *invoker) Globber() *rpc.GlobState {
	return &rpc.GlobState{AllGlobber: javaGlobber{i}}
}

func (i *invoker) Signature(ctx rpc.ServerCall) ([]signature.Interface, error) {
	jEnv, freeFunc := jutil.GetEnv()
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	replySign := jutil.ClassSign("io.v.v23.vdlroot.signature.Interface")

	interfacesArr, err := jutil.CallObjectArrayMethod(env, i.jInvoker, "getSignature", nil, replySign)
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

func (i *invoker) MethodSignature(ctx rpc.ServerCall, method string) (signature.Method, error) {
	jEnv, freeFunc := jutil.GetEnv()
	env := (*C.JNIEnv)(jEnv)
	defer freeFunc()

	replySign := jutil.ClassSign("io.v.v23.vdlroot.signature.Method")

	jMethod, err := jutil.CallObjectMethod(env, i.jInvoker, "getMethodSignature", []jutil.Sign{jutil.StringSign}, replySign, method)
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

// encodeArgs VOM-encodes the provided arguments pointers and returns them as a
// Java array of byte arrays.
func encodeArgs(env *C.JNIEnv, argptrs []interface{}) (C.jobjectArray, error) {
	vomArgs := make([][]byte, len(argptrs))
	for i, argptr := range argptrs {
		arg := interface{}(jutil.DerefOrDie(argptr))
		var err error
		if vomArgs[i], err = vom.Encode(arg); err != nil {
			return nil, err
		}
	}
	return C.jobjectArray(jutil.JByteArrayArray(env, vomArgs)), nil
}

// decodeResults VOM-decodes replies stored in the Java reply object into
// an array *vdl.Value.
func decodeResults(env *C.JNIEnv, jReply C.jobject) ([]interface{}, error) {
	// Unpack the replies.
	results, err := jutil.JByteArrayArrayField(env, jReply, "results")
	if err != nil {
		return nil, err
	}
	vomAppErr, err := jutil.JByteArrayField(env, jReply, "vomAppError")
	if err != nil {
		return nil, err
	}
	// Check for app error.
	if vomAppErr != nil {
		var appErr error
		if err := vom.Decode(vomAppErr, &appErr); err != nil {
			return nil, err
		}
		return nil, appErr
	}
	// VOM-decode results into *vdl.Value instances.
	ret := make([]interface{}, len(results))
	for i, result := range results {
		var err error
		if ret[i], err = jutil.VomDecodeToValue(result); err != nil {
			return nil, err
		}
	}
	return ret, nil
}
