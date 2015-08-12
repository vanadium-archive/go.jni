// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"fmt"
	"runtime"

	"v.io/v23/context"
	"v.io/v23/glob"
	"v.io/v23/naming"
	"v.io/v23/rpc"
	"v.io/v23/vdl"
	"v.io/v23/vdlroot/signature"
	"v.io/v23/vom"

	jchannel "v.io/x/jni/impl/google/channel"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

func goInvoker(env jutil.Env, obj jutil.Object) (rpc.Invoker, error) {
	// See if the Java object is an invoker.
	var jInvoker jutil.Object
	if jutil.IsInstanceOf(env, obj, jInvokerClass) {
		jInvoker = obj
	} else {
		// Create a new Java ReflectInvoker object.
		jReflectInvoker, err := jutil.NewObject(env, jReflectInvokerClass, []jutil.Sign{jutil.ObjectSign}, obj)
		if err != nil {
			return nil, fmt.Errorf("error creating Java ReflectInvoker object: %v", err)
		}
		jInvoker = jReflectInvoker
	}

	// Reference Java invoker; it will be de-referenced when the go invoker
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jInvoker = jutil.NewGlobalRef(env, jInvoker)
	i := &invoker{
		jInvoker: jInvoker,
	}
	runtime.SetFinalizer(i, func(i *invoker) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, i.jInvoker)
	})
	return i, nil
}

type invoker struct {
	jInvoker jutil.Object
}

func (i *invoker) Prepare(ctx *context.T, method string, numArgs int) (argptrs []interface{}, tags []*vdl.Value, err error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	// Have all input arguments be decoded into *vdl.Value.
	argptrs = make([]interface{}, numArgs)
	for i := 0; i < numArgs; i++ {
		value := new(vdl.Value)
		argptrs[i] = &value
	}
	jVomTags, err := jutil.CallStaticObjectMethod(env, jUtilClass, "getMethodTags", []jutil.Sign{invokerSign, jutil.StringSign}, jutil.ArraySign(jutil.ByteArraySign), i.jInvoker, jutil.CamelCase(method))
	if err != nil {
		return nil, nil, err
	}
	vomTags, err := jutil.GoByteArrayArray(env, jVomTags)
	if err != nil {
		return nil, nil, err
	}
	tags = make([]*vdl.Value, len(vomTags))
	for i, vomTag := range vomTags {
		var err error
		if tags[i], err = jutil.VomDecodeToValue(vomTag); err != nil {
			return nil, nil, err
		}
	}
	return
}

func (i *invoker) Invoke(ctx *context.T, call rpc.StreamServerCall, method string, argptrs []interface{}) (results []interface{}, err error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		return nil, err
	}
	jStreamServerCall, err := javaStreamServerCall(env, call)
	if err != nil {
		return nil, err
	}
	vomArgs := make([][]byte, len(argptrs))
	for i, argptr := range argptrs {
		arg := interface{}(jutil.DerefOrDie(argptr))
		var err error
		if vomArgs[i], err = vom.Encode(arg); err != nil {
			return nil, err
		}
	}
	jVomArgs, err := jutil.JByteArrayArray(env, vomArgs)
	if err != nil {
		return nil, err
	}
	jVomResults, err := jutil.CallStaticObjectMethod(env, jUtilClass, "invoke", []jutil.Sign{invokerSign, contextSign, streamServerCallSign, jutil.StringSign, jutil.ArraySign(jutil.ByteArraySign)}, jutil.ArraySign(jutil.ByteArraySign), i.jInvoker, jContext, jStreamServerCall, jutil.CamelCase(method), jVomArgs)
	if err != nil {
		return nil, err
	}
	vomResults, err := jutil.GoByteArrayArray(env, jVomResults)
	if err != nil {
		return nil, err
	}
	results = make([]interface{}, len(vomResults))
	for i, vomResult := range vomResults {
		var err error
		if results[i], err = jutil.VomDecodeToValue(vomResult); err != nil {
			return nil, err
		}
	}
	return
}

func (i *invoker) Signature(ctx *context.T, call rpc.ServerCall) ([]signature.Interface, error) {
	env, freeFunc := jutil.GetEnv()
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
		err = jutil.GoVomCopy(env, jInterface, jInterfaceClass, &result[i])
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (i *invoker) MethodSignature(ctx *context.T, call rpc.ServerCall, method string) (signature.Method, error) {
	env, freeFunc := jutil.GetEnv()
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
	err = jutil.GoVomCopy(env, jMethod, jMethodClass, &result)
	if err != nil {
		return signature.Method{}, err
	}
	return result, nil
}

func (i *invoker) Globber() *rpc.GlobState {
	return &rpc.GlobState{AllGlobber: javaGlobber{i}}
}

type javaGlobber struct {
	i *invoker
}

func (j javaGlobber) Glob__(ctx *context.T, call rpc.GlobServerCall, g *glob.Glob) error {
	// TODO(sjr,rthellend): Update the Java API to match the new GO API.
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jServerCall, err := JavaServerCall(env, call)
	if err != nil {
		return err
	}

	readFunc := func(input jutil.Object) error {
		env, freeFunc := jutil.GetEnv()
		defer jutil.DeleteGlobalRef(env, input)
		var reply naming.GlobReply
		err := jutil.GoVomCopy(env, input, jGlobReplyClass, &reply)
		if err != nil {
			freeFunc()
			return err
		}
		freeFunc()
		call.SendStream().Send(reply)
		return nil
	}
	closeFunc := func() error {
		return nil
	}

	jOutputChannel, err := jchannel.JavaOutputChannel(env, readFunc, closeFunc)
	if err != nil {
		return err
	}

	callSign := jutil.ClassSign("io.v.v23.rpc.ServerCall")
	channelSign := jutil.ClassSign("io.v.v23.OutputChannel")

	jServerCall = jutil.NewGlobalRef(env, jServerCall)
	jOutputChannel = jutil.NewGlobalRef(env, jOutputChannel)
	// Calls Java invoker's glob method.
	jutil.CallVoidMethod(env, j.i.jInvoker, "glob", []jutil.Sign{callSign, jutil.StringSign, channelSign}, jServerCall, g.String(), jOutputChannel)
	jutil.DeleteGlobalRef(env, jServerCall)
	jutil.DeleteGlobalRef(env, jOutputChannel)
	return nil
}
