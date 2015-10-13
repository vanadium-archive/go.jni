// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"fmt"
	"io"
	"unsafe"

	"v.io/v23/context"
	"v.io/v23/options"
	"v.io/v23/rpc"
	"v.io/v23/security"
	"v.io/v23/vdl"
	"v.io/v23/vom"

	jchannel "v.io/x/jni/impl/google/channel"
	jbt "v.io/x/jni/impl/google/rpc/protocols/bt"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	jnaming "v.io/x/jni/v23/naming"
	jsecurity "v.io/x/jni/v23/security"
)

// #include "jni.h"
import "C"

var (
	contextSign          = jutil.ClassSign("io.v.v23.context.VContext")
	invokerSign          = jutil.ClassSign("io.v.v23.rpc.Invoker")
	clientCallSign       = jutil.ClassSign("io.v.v23.rpc.ClientCall")
	serverCallSign       = jutil.ClassSign("io.v.v23.rpc.ServerCall")
	streamServerCallSign = jutil.ClassSign("io.v.v23.rpc.StreamServerCall")
	listenAddrSign       = jutil.ClassSign("io.v.v23.rpc.ListenSpec$Address")
	addressChooserSign   = jutil.ClassSign("io.v.v23.rpc.AddressChooser")
	serverStateSign      = jutil.ClassSign("io.v.v23.rpc.ServerState")
	streamSign           = jutil.ClassSign("io.v.v23.rpc.Stream")
	optionsSign          = jutil.ClassSign("io.v.v23.Options")
	// Global reference for io.v.impl.google.rpc.AddressChooserImpl class.
	jAddressChooserImplClass jutil.Class
	// Global reference for io.v.impl.google.rpc.ServerImpl class.
	jServerImplClass jutil.Class
	// Global reference for io.v.impl.google.rpc.ClientImpl class.
	jClientImplClass jutil.Class
	// Global reference for io.v.impl.google.rpc.ClientCallImpl class.
	jClientCallImplClass jutil.Class
	// Global reference for io.v.impl.google.rpc.StreamServerCallImpl class.
	jStreamServerCallImplClass jutil.Class
	// Global reference for io.v.impl.google.rpc.ServerCallImpl class.
	jServerCallImplClass jutil.Class
	// Global reference for io.v.impl.google.rpc.StreamImpl class.
	jStreamImplClass jutil.Class
	// Global reference for io.v.impl.google.rpc.Util class.
	jUtilClass jutil.Class
	// Global reference for io.v.v23.rpc.Invoker class.
	jInvokerClass jutil.Class
	// Global reference for io.v.v23.rpc.ListenSpec class.
	jListenSpecClass jutil.Class
	// Global reference for io.v.v23.rpc.ListenSpec$Address class.
	jListenSpecAddressClass jutil.Class
	// Global reference for io.v.v23.rpc.MountStatus class.
	jMountStatusClass jutil.Class
	// Global reference for io.v.v23.rpc.NetworkAddress class.
	jNetworkAddressClass jutil.Class
	// Global reference for io.v.v23.rpc.NetworkChange class.
	jNetworkChangeClass jutil.Class
	// Global reference for io.v.v23.rpc.ProxyStatus class.
	jProxyStatusClass jutil.Class
	// Global reference for io.v.v23.rpc.ReflectInvoker class.
	jReflectInvokerClass jutil.Class
	// Global reference for io.v.v23.rpc.ServerStatus class.
	jServerStatusClass jutil.Class
	// Global reference for io.v.v23.rpc.ServerState class.
	jServerStateClass jutil.Class
	// Global reference for io.v.v23.OptionDefs class.
	jOptionDefsClass jutil.Class
	// Global reference for java.io.EOFException class.
	jEOFExceptionClass jutil.Class
	// Global reference for io.v.v23.naming.Endpoint.
	jEndpointClass jutil.Class
	// Global reference for io.v.v23.vdlroot.signature.Interface class.
	jInterfaceClass jutil.Class
	// Global reference for io.v.v23.vdlroot.signature.Method class.
	jMethodClass jutil.Class
	// Global reference for io.v.v23.naming.GlobReply
	jGlobReplyClass jutil.Class
	// Global reference for java.lang.Object class.
	jObjectClass jutil.Class
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
func Init(env jutil.Env) error {
	if err := jbt.Init(env); err != nil {
		return err
	}
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	var err error
	jAddressChooserImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rpc/AddressChooserImpl")
	if err != nil {
		return err
	}
	jServerImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rpc/ServerImpl")
	if err != nil {
		return err
	}
	jClientImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rpc/ClientImpl")
	if err != nil {
		return err
	}
	jClientCallImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rpc/ClientCallImpl")
	if err != nil {
		return err
	}
	jStreamServerCallImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rpc/StreamServerCallImpl")
	if err != nil {
		return err
	}
	jServerCallImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rpc/ServerCallImpl")
	if err != nil {
		return err
	}
	jStreamImplClass, err = jutil.JFindClass(env, "io/v/impl/google/rpc/StreamImpl")
	if err != nil {
		return err
	}
	jUtilClass, err = jutil.JFindClass(env, "io/v/impl/google/rpc/Util")
	if err != nil {
		return err
	}
	jInvokerClass, err = jutil.JFindClass(env, "io/v/v23/rpc/Invoker")
	if err != nil {
		return err
	}
	jListenSpecClass, err = jutil.JFindClass(env, "io/v/v23/rpc/ListenSpec")
	if err != nil {
		return err
	}
	jListenSpecAddressClass, err = jutil.JFindClass(env, "io/v/v23/rpc/ListenSpec$Address")
	if err != nil {
		return err
	}
	jMountStatusClass, err = jutil.JFindClass(env, "io/v/v23/rpc/MountStatus")
	if err != nil {
		return err
	}
	jNetworkAddressClass, err = jutil.JFindClass(env, "io/v/v23/rpc/NetworkAddress")
	if err != nil {
		return err
	}
	jNetworkChangeClass, err = jutil.JFindClass(env, "io/v/v23/rpc/NetworkChange")
	if err != nil {
		return err
	}
	jProxyStatusClass, err = jutil.JFindClass(env, "io/v/v23/rpc/ProxyStatus")
	if err != nil {
		return err
	}
	jReflectInvokerClass, err = jutil.JFindClass(env, "io/v/v23/rpc/ReflectInvoker")
	if err != nil {
		return err
	}
	jServerStatusClass, err = jutil.JFindClass(env, "io/v/v23/rpc/ServerStatus")
	if err != nil {
		return err
	}
	jServerStateClass, err = jutil.JFindClass(env, "io/v/v23/rpc/ServerState")
	if err != nil {
		return err
	}
	jOptionDefsClass, err = jutil.JFindClass(env, "io/v/v23/OptionDefs")
	if err != nil {
		return err
	}
	jEOFExceptionClass, err = jutil.JFindClass(env, "java/io/EOFException")
	if err != nil {
		return err
	}
	jEndpointClass, err = jutil.JFindClass(env, "io/v/v23/naming/Endpoint")
	if err != nil {
		return err
	}
	jInterfaceClass, err = jutil.JFindClass(env, "io/v/v23/vdlroot/signature/Interface")
	if err != nil {
		return err
	}
	jMethodClass, err = jutil.JFindClass(env, "io/v/v23/vdlroot/signature/Method")
	if err != nil {
		return err
	}
	jGlobReplyClass, err = jutil.JFindClass(env, "io/v/v23/naming/GlobReply")
	if err != nil {
		return err
	}
	jObjectClass, err = jutil.JFindClass(env, "java/lang/Object")
	if err != nil {
		return err
	}
	return nil
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeAddName
func Java_io_v_impl_google_rpc_ServerImpl_nativeAddName(jenv *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jName C.jstring) {
	env := jutil.WrapEnv(jenv)
	name := jutil.GoString(env, jutil.WrapObject(jName))
	if err := (*(*rpc.Server)(jutil.NativePtr(goPtr))).AddName(name); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeRemoveName
func Java_io_v_impl_google_rpc_ServerImpl_nativeRemoveName(jenv *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jName C.jstring) {
	env := jutil.WrapEnv(jenv)
	name := jutil.GoString(env, jutil.WrapObject(jName))
	(*(*rpc.Server)(jutil.NativePtr(goPtr))).RemoveName(name)
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeGetStatus
func Java_io_v_impl_google_rpc_ServerImpl_nativeGetStatus(jenv *C.JNIEnv, jServer C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	status := (*(*rpc.Server)(jutil.NativePtr(goPtr))).Status()
	jStatus, err := JavaServerStatus(env, status)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jStatus))
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeWatchNetwork
func Java_io_v_impl_google_rpc_ServerImpl_nativeWatchNetwork(jenv *C.JNIEnv, jServer C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	networkChan := make(chan rpc.NetworkChange, 100)
	(*(*rpc.Server)(jutil.NativePtr(goPtr))).WatchNetwork(networkChan)
	retChan := make(chan jutil.Object, 100)
	go func() {
		for change := range networkChan {
			env, freeFunc := jutil.GetEnv()
			jChange, err := JavaNetworkChange(env, change)
			if err != nil {
				freeFunc()
				continue
			}
			jChange = jutil.NewGlobalRef(env, jChange)
			freeFunc()
			retChan <- jChange
		}
		close(retChan)
	}()
	jIterable, err := jchannel.JavaIterable(env, &retChan, &networkChan)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jIterable))
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeUnwatchNetwork
func Java_io_v_impl_google_rpc_ServerImpl_nativeUnwatchNetwork(jenv *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jChannelIterable C.jobject) {
	env := jutil.WrapEnv(jenv)
	goNetworkChanPtr, err := jutil.CallLongMethod(env, jutil.WrapObject(jChannelIterable), "getSourceNativePtr", nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	networkChan := *(*chan rpc.NetworkChange)(unsafe.Pointer(uintptr(goNetworkChanPtr)))
	(*(*rpc.Server)(jutil.NativePtr(goPtr))).UnwatchNetwork(networkChan)
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeStop
func Java_io_v_impl_google_rpc_ServerImpl_nativeStop(jenv *C.JNIEnv, jServer C.jobject, goPtr C.jlong) {
	env := jutil.WrapEnv(jenv)
	s := (*rpc.Server)(jutil.NativePtr(goPtr))
	if err := (*s).Stop(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeFinalize
func Java_io_v_impl_google_rpc_ServerImpl_nativeFinalize(jenv *C.JNIEnv, jServer C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

func decodeArgs(env jutil.Env, jVomArgs C.jobjectArray) ([]interface{}, error) {
	vomArgs, err := jutil.GoByteArrayArray(env, jutil.WrapObject(jVomArgs))
	if err != nil {
		return nil, err
	}
	// VOM-decode each arguments into a *vdl.Value.
	args := make([]interface{}, len(vomArgs))
	for i := 0; i < len(vomArgs); i++ {
		var err error
		if args[i], err = jutil.VomDecodeToValue(vomArgs[i]); err != nil {
			return nil, err
		}
	}
	return args, nil
}

func doStartCall(env jutil.Env, context *context.T, name, method string, skipServerAuth bool, goPtr C.jlong, args []interface{}) (jutil.Object, error) {
	var opts []rpc.CallOpt
	if skipServerAuth {
		opts = append(opts,
			options.NameResolutionAuthorizer{security.AllowEveryone()},
			options.ServerAuthorizer{security.AllowEveryone()})
	}
	// Invoke StartCall
	call, err := (*(*rpc.Client)(jutil.NativePtr(goPtr))).StartCall(context, name, method, args, opts...)
	if err != nil {
		return jutil.NullObject, err
	}
	jCall, err := javaCall(env, call)
	if err != nil {
		return jutil.NullObject, err
	}
	return jCall, nil
}

//export Java_io_v_impl_google_rpc_ClientImpl_nativeStartCall
func Java_io_v_impl_google_rpc_ClientImpl_nativeStartCall(jenv *C.JNIEnv, jClient C.jobject, goPtr C.jlong, jContext C.jobject, jName C.jstring, jMethod C.jstring, jVomArgs C.jobjectArray, jSkipServerAuth C.jboolean) C.jobject {
	env := jutil.WrapEnv(jenv)
	name := jutil.GoString(env, jutil.WrapObject(jName))
	method := jutil.GoString(env, jutil.WrapObject(jMethod))
	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	args, err := decodeArgs(env, jVomArgs)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	result, err := doStartCall(env, context, name, method, jSkipServerAuth == C.JNI_TRUE, goPtr, args)
	if err != nil {
		jutil.JThrowV(env, err)
	}
	return C.jobject(unsafe.Pointer(result))
}

func callOnFailure(env jutil.Env, callback jutil.Object, err error) {
	if err := jutil.CallVoidMethod(env, callback, "onFailure", []jutil.Sign{jutil.VExceptionSign}, err); err != nil {
		panic(fmt.Sprintf("couldn't call Java onFailure method: %v", err))
	}
}

func callOnSuccess(env jutil.Env, callback jutil.Object, jClientCall jutil.Object) {
	if err := jutil.CallVoidMethod(env, callback, "onSuccess", []jutil.Sign{jutil.ObjectSign}, jClientCall); err != nil {
		panic(fmt.Sprintf("couldn't call Java onSuccess method: %v", err))
	}
}

//export Java_io_v_impl_google_rpc_ClientImpl_nativeStartCallAsync
func Java_io_v_impl_google_rpc_ClientImpl_nativeStartCallAsync(jenv *C.JNIEnv, jClient C.jobject, goPtr C.jlong, jContext C.jobject, jName C.jstring, jMethod C.jstring, jVomArgs C.jobjectArray, jSkipServerAuth C.jboolean, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	name := jutil.GoString(env, jutil.WrapObject(jName))
	method := jutil.GoString(env, jutil.WrapObject(jMethod))

	context, err := jcontext.GoContext(env, jutil.WrapObject(jContext))
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	args, err := decodeArgs(env, jVomArgs)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	skipServerAuth := jSkipServerAuth == C.JNI_TRUE
	go func(jCallback jutil.Object) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		defer jutil.DeleteGlobalRef(env, jCallback)
		if jCall, err := doStartCall(env, context, name, method, skipServerAuth, goPtr, args); err != nil {
			callOnFailure(env, jCallback, err)
		} else {
			callOnSuccess(env, jCallback, jCall)
		}
	}(jutil.NewGlobalRef(env, jutil.WrapObject(jCallback)))
}

//export Java_io_v_impl_google_rpc_ClientImpl_nativeClose
func Java_io_v_impl_google_rpc_ClientImpl_nativeClose(jenv *C.JNIEnv, jClient C.jobject, goPtr C.jlong) {
	(*(*rpc.Client)(jutil.NativePtr(goPtr))).Close()
}

//export Java_io_v_impl_google_rpc_ClientImpl_nativeFinalize
func Java_io_v_impl_google_rpc_ClientImpl_nativeFinalize(jenv *C.JNIEnv, jClient C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_impl_google_rpc_StreamImpl_nativeSend
func Java_io_v_impl_google_rpc_StreamImpl_nativeSend(jenv *C.JNIEnv, jStream C.jobject, goPtr C.jlong, jVomItem C.jbyteArray) {
	env := jutil.WrapEnv(jenv)
	vomItem := jutil.GoByteArray(env, jutil.WrapObject(jVomItem))
	item, err := jutil.VomDecodeToValue(vomItem)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*rpc.Stream)(jutil.NativePtr(goPtr))).Send(item); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_StreamImpl_nativeRecv
func Java_io_v_impl_google_rpc_StreamImpl_nativeRecv(jenv *C.JNIEnv, jStream C.jobject, goPtr C.jlong) C.jbyteArray {
	env := jutil.WrapEnv(jenv)
	result := new(vdl.Value)
	if err := (*(*rpc.Stream)(jutil.NativePtr(goPtr))).Recv(&result); err != nil {
		if err == io.EOF {
			jutil.JThrow(env, jEOFExceptionClass, err.Error())
			return nil
		}
		jutil.JThrowV(env, err)
		return nil
	}
	vomResult, err := vom.Encode(result)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jArr, err := jutil.JByteArray(env, vomResult)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jbyteArray(unsafe.Pointer(jArr))
}

//export Java_io_v_impl_google_rpc_StreamImpl_nativeFinalize
func Java_io_v_impl_google_rpc_StreamImpl_nativeFinalize(jenv *C.JNIEnv, jStream C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_impl_google_rpc_ClientCallImpl_nativeCloseSend
func Java_io_v_impl_google_rpc_ClientCallImpl_nativeCloseSend(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	env := jutil.WrapEnv(jenv)
	if err := (*(*rpc.ClientCall)(jutil.NativePtr(goPtr))).CloseSend(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

func doFinish(env jutil.Env, goPtr C.jlong, numResults int) (jutil.Object, error) {
	// Have all the results be decoded into *vdl.Value.
	resultPtrs := make([]interface{}, numResults)
	for i := 0; i < numResults; i++ {
		value := new(vdl.Value)
		resultPtrs[i] = &value
	}
	if err := (*(*rpc.ClientCall)(jutil.NativePtr(goPtr))).Finish(resultPtrs...); err != nil {
		// Invocation error.
		return jutil.NullObject, err
	}
	// VOM-encode the results.
	vomResults := make([][]byte, numResults)
	for i, resultPtr := range resultPtrs {
		// Remove the pointer from the result.  Simply *resultPtr doesn't work
		// as resultPtr is of type interface{}.
		result := interface{}(jutil.DerefOrDie(resultPtr))
		var err error
		if vomResults[i], err = vom.Encode(result); err != nil {
			return jutil.NullObject, err
		}
	}
	jArr, err := jutil.JByteArrayArray(env, vomResults)
	if err != nil {
		return jutil.NullObject, err
	}
	return jArr, nil
}

//export Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinish
func Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinish(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong, jNumResults C.jint) C.jobjectArray {
	env := jutil.WrapEnv(jenv)
	numResults := int(jNumResults)
	result, err := doFinish(env, goPtr, numResults)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(unsafe.Pointer(result))
}

//export Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinishAsync
func Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinishAsync(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong, jNumResults C.jint, jCallback C.jobject) {
	env := jutil.WrapEnv(jenv)
	numResults := int(jNumResults)
	go func(jCallback jutil.Object) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		defer jutil.DeleteGlobalRef(env, jCallback)
		if result, err := doFinish(env, goPtr, numResults); err != nil {
			callOnFailure(env, jCallback, err)
		} else {
			callOnSuccess(env, jCallback, result)
		}
	}(jutil.NewGlobalRef(env, jutil.WrapObject(jCallback)))
}

//export Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinalize
func Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinalize(jenv *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeSecurity
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeSecurity(jenv *C.JNIEnv, jServerCallClass C.jclass, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	securityCall := (*(*rpc.ServerCall)(jutil.NativePtr(goPtr))).Security()
	if securityCall == nil {
		return nil
	}
	jSecurityCall, err := jsecurity.JavaCall(env, securityCall)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jSecurityCall))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeSuffix
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeSuffix(jenv *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jstring {
	env := jutil.WrapEnv(jenv)
	jSuffix := jutil.JString(env, (*(*rpc.ServerCall)(jutil.NativePtr(goPtr))).Suffix())
	return C.jstring(unsafe.Pointer(jSuffix))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeLocalEndpoint
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeLocalEndpoint(jenv *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	jEndpoint, err := jnaming.JavaEndpoint(env, (*(*rpc.ServerCall)(jutil.NativePtr(goPtr))).LocalEndpoint())
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jEndpoint))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeRemoteEndpoint
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeRemoteEndpoint(jenv *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	jEndpoint, err := jnaming.JavaEndpoint(env, (*(*rpc.ServerCall)(jutil.NativePtr(goPtr))).RemoteEndpoint())
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jEndpoint))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeGrantedBlessings
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeGrantedBlessings(jenv *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	blessings := (*(*rpc.ServerCall)(jutil.NativePtr(goPtr))).GrantedBlessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jBlessings))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeServer
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeServer(jenv *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	env := jutil.WrapEnv(jenv)
	server := (*(*rpc.ServerCall)(jutil.NativePtr(goPtr))).Server()
	jServer, err := JavaServer(env, server)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(unsafe.Pointer(jServer))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeFinalize
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeFinalize(jenv *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_impl_google_rpc_StreamServerCallImpl_nativeFinalize
func Java_io_v_impl_google_rpc_StreamServerCallImpl_nativeFinalize(jenv *C.JNIEnv, jStreamServerCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_impl_google_rpc_AddressChooserImpl_nativeChoose
func Java_io_v_impl_google_rpc_AddressChooserImpl_nativeChoose(jenv *C.JNIEnv, jAddressChooser C.jobject, goPtr C.jlong, jProtocol C.jstring, jCandidates C.jobjectArray) C.jobjectArray {
	env := jutil.WrapEnv(jenv)
	protocol := jutil.GoString(env, jutil.WrapObject(jProtocol))
	candidates, err := GoNetworkAddressArray(env, jutil.WrapObject(jCandidates))
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	addrs, err := (*(*rpc.AddressChooser)(jutil.NativePtr(goPtr))).ChooseAddresses(protocol, candidates)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jAddrs, err := JavaNetworkAddressArray(env, addrs)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(unsafe.Pointer(jAddrs))
}

//export Java_io_v_impl_google_rpc_AddressChooserImpl_nativeFinalize
func Java_io_v_impl_google_rpc_AddressChooserImpl_nativeFinalize(jenv *C.JNIEnv, jAddressChooser C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.NativePtr(goPtr))
}

//export Java_io_v_impl_google_rpc_Util_nativeGoInvoker
func Java_io_v_impl_google_rpc_Util_nativeGoInvoker(jenv *C.JNIEnv, jUtil C.jclass, jServiceObject C.jobject) C.jlong {
	env := jutil.WrapEnv(jenv)
	invoker, err := goInvoker(env, jutil.WrapObject(jServiceObject))
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	jutil.GoRef(&invoker) // Un-refed when the Go invoker is returned to the Go runtime
	return C.jlong(jutil.PtrValue(&invoker))
}

//export Java_io_v_impl_google_rpc_Util_nativeGoAuthorizer
func Java_io_v_impl_google_rpc_Util_nativeGoAuthorizer(jenv *C.JNIEnv, jUtil C.jclass, jAuthorizer C.jobject) C.jlong {
	env := jutil.WrapEnv(jenv)
	auth, err := jsecurity.GoAuthorizer(env, jutil.WrapObject(jAuthorizer))
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	jutil.GoRef(&auth) // Un-refed when the Go authorizer is returned to the Go runtime
	return C.jlong(jutil.PtrValue(&auth))
}
