// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"io"
	"log"
	"net"
	"unsafe"

	"v.io/v23/options"
	"v.io/v23/rpc"
	"v.io/v23/vdl"
	"v.io/v23/vom"

	jchannel "v.io/x/jni/impl/google/channel"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
	jsecurity "v.io/x/jni/v23/security"
)

// #include "jni.h"
import "C"

var (
	contextSign          = jutil.ClassSign("io.v.v23.context.VContext")
	invokerSign          = jutil.ClassSign("io.v.v23.rpc.Invoker")
	serverCallSign       = jutil.ClassSign("io.v.v23.rpc.ServerCall")
	streamServerCallSign = jutil.ClassSign("io.v.v23.rpc.StreamServerCall")
	listenAddrSign       = jutil.ClassSign("io.v.v23.rpc.ListenSpec$Address")
	addressChooserSign   = jutil.ClassSign("io.v.v23.rpc.AddressChooser")
	serverStateSign      = jutil.ClassSign("io.v.v23.rpc.ServerState")
	streamSign           = jutil.ClassSign("io.v.v23.rpc.Stream")
	optionsSign          = jutil.ClassSign("io.v.v23.Options")
	// Global reference for io.v.impl.google.rpc.AddressChooserImpl class.
	jAddressChooserImplClass C.jclass
	// Global reference for io.v.impl.google.rpc.ServerImpl class.
	jServerImplClass C.jclass
	// Global reference for io.v.impl.google.rpc.ClientImpl class.
	jClientImplClass C.jclass
	// Global reference for io.v.impl.google.rpc.ClientCallImpl class.
	jClientCallImplClass C.jclass
	// Global reference for io.v.impl.google.rpc.StreamServerCallImpl class.
	jStreamServerCallImplClass C.jclass
	// Global reference for io.v.impl.google.rpc.ServerCallImpl class.
	jServerCallImplClass C.jclass
	// Global reference for io.v.impl.google.rpc.StreamImpl class.
	jStreamImplClass C.jclass
	// Global reference for io.v.impl.google.rpc.Util class.
	jUtilClass C.jclass
	// Global reference for io.v.v23.rpc.Invoker class.
	jInvokerClass C.jclass
	// Global reference for io.v.v23.rpc.ListenSpec class.
	jListenSpecClass C.jclass
	// Global reference for io.v.v23.rpc.ListenSpec$Address class.
	jListenSpecAddressClass C.jclass
	// Global reference for io.v.v23.rpc.MountStatus class.
	jMountStatusClass C.jclass
	// Global reference for io.v.v23.rpc.NetworkAddress class.
	jNetworkAddressClass C.jclass
	// Global reference for io.v.v23.rpc.NetworkChange class.
	jNetworkChangeClass C.jclass
	// Global reference for io.v.v23.rpc.ProxyStatus class.
	jProxyStatusClass C.jclass
	// Global reference for io.v.v23.rpc.ReflectInvoker class.
	jReflectInvokerClass C.jclass
	// Global reference for io.v.v23.rpc.ServerStatus class.
	jServerStatusClass C.jclass
	// Global reference for io.v.v23.rpc.ServerState class.
	jServerStateClass C.jclass
	// Global reference for io.v.v23.OptionDefs class.
	jOptionDefsClass C.jclass
	// Global reference for java.io.EOFException class.
	jEOFExceptionClass C.jclass
	// Global reference for java.lang.String class.
	jStringClass C.jclass
	// Global reference for io.v.v23.vdlroot.signature.Interface class.
	jInterfaceClass C.jclass
	// Global reference for io.v.v23.vdlroot.signature.Method class.
	jMethodClass C.jclass
	// Global reference for io.v.v23.naming.GlobReply
	jGlobReplyClass C.jclass
	// Global reference for java.lang.Object class.
	jObjectClass C.jclass
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java environment is passed in an empty
// interface and then cast into the package-local environment type.
func Init(jEnv interface{}) error {
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	class, err := jutil.JFindClass(jEnv, "io/v/impl/google/rpc/AddressChooserImpl")
	if err != nil {
		return err
	}
	jAddressChooserImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/ServerImpl")
	if err != nil {
		return err
	}
	jServerImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/ClientImpl")
	if err != nil {
		return err
	}
	jClientImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/ClientCallImpl")
	if err != nil {
		return err
	}
	jClientCallImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/StreamServerCallImpl")
	if err != nil {
		return err
	}
	jStreamServerCallImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/ServerCallImpl")
	if err != nil {
		return err
	}
	jServerCallImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/StreamImpl")
	if err != nil {
		return err
	}
	jStreamImplClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/impl/google/rpc/Util")
	if err != nil {
		return err
	}
	jUtilClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/Invoker")
	if err != nil {
		return err
	}
	jInvokerClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ListenSpec")
	if err != nil {
		return err
	}
	jListenSpecClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ListenSpec$Address")
	if err != nil {
		return err
	}
	jListenSpecAddressClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/MountStatus")
	if err != nil {
		return err
	}
	jMountStatusClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/NetworkAddress")
	if err != nil {
		return err
	}
	jNetworkAddressClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/NetworkChange")
	if err != nil {
		return err
	}
	jNetworkChangeClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ProxyStatus")
	if err != nil {
		return err
	}
	jProxyStatusClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ReflectInvoker")
	if err != nil {
		return err
	}
	jReflectInvokerClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ServerStatus")
	if err != nil {
		return err
	}
	jServerStatusClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/rpc/ServerState")
	if err != nil {
		return err
	}
	jServerStateClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/OptionDefs")
	if err != nil {
		return err
	}
	jOptionDefsClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "java/io/EOFException")
	if err != nil {
		return err
	}
	jEOFExceptionClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "java/lang/String")
	if err != nil {
		return err
	}
	jStringClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/vdlroot/signature/Interface")
	if err != nil {
		return err
	}
	jInterfaceClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/vdlroot/signature/Method")
	if err != nil {
		return err
	}
	jMethodClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "io/v/v23/naming/GlobReply")
	if err != nil {
		return err
	}
	jGlobReplyClass = C.jclass(class)
	class, err = jutil.JFindClass(jEnv, "java/lang/Object")
	if err != nil {
		return err
	}
	jObjectClass = C.jclass(class)
	return nil
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeListen
func Java_io_v_impl_google_rpc_ServerImpl_nativeListen(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jSpec C.jobject) C.jobjectArray {
	spec, err := GoListenSpec(env, jSpec)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	eps, err := (*(*rpc.Server)(jutil.Ptr(goPtr))).Listen(spec)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	epStrs := make([]string, len(eps))
	for i, ep := range eps {
		epStrs[i] = ep.String()
	}
	return C.jobjectArray(jutil.JStringArray(env, epStrs))
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeServe
func Java_io_v_impl_google_rpc_ServerImpl_nativeServe(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jName C.jstring, jDispatcher C.jobject) {
	name := jutil.GoString(env, jName)
	d, err := goDispatcher(env, jDispatcher)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*rpc.Server)(jutil.Ptr(goPtr))).ServeDispatcher(name, d); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeAddName
func Java_io_v_impl_google_rpc_ServerImpl_nativeAddName(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jName C.jstring) {
	name := jutil.GoString(env, jName)
	if err := (*(*rpc.Server)(jutil.Ptr(goPtr))).AddName(name); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeRemoveName
func Java_io_v_impl_google_rpc_ServerImpl_nativeRemoveName(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jName C.jstring) {
	name := jutil.GoString(env, jName)
	(*(*rpc.Server)(jutil.Ptr(goPtr))).RemoveName(name)
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeGetStatus
func Java_io_v_impl_google_rpc_ServerImpl_nativeGetStatus(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong) C.jobject {
	status := (*(*rpc.Server)(jutil.Ptr(goPtr))).Status()
	jStatus, err := JavaServerStatus(env, status)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jStatus)
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeWatchNetwork
func Java_io_v_impl_google_rpc_ServerImpl_nativeWatchNetwork(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong) C.jobject {
	networkChan := make(chan rpc.NetworkChange, 100)
	(*(*rpc.Server)(jutil.Ptr(goPtr))).WatchNetwork(networkChan)
	retChan := make(chan C.jobject, 100)
	go func() {
		for change := range networkChan {
			jEnv, freeFunc := jutil.GetEnv()
			env := (*C.JNIEnv)(jEnv)
			defer freeFunc()
			jChangeObj, err := JavaNetworkChange(env, change)
			if err != nil {
				log.Printf("Couldn't convert Go NetworkChange %v to Java\n", change)
				continue
			}
			jChange := C.jobject(jutil.NewGlobalRef(env, jChangeObj))
			retChan <- jChange
		}
		close(retChan)
	}()
	jInputChannel, err := jchannel.JavaInputChannel(env, &retChan, &networkChan)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jInputChannel)
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeUnwatchNetwork
func Java_io_v_impl_google_rpc_ServerImpl_nativeUnwatchNetwork(env *C.JNIEnv, jServer C.jobject, goPtr C.jlong, jInputChannel C.jobject) {
	goNetworkChanPtr, err := jutil.CallLongMethod(env, jInputChannel, "getSourceNativePtr", nil)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	networkChan := *(*chan rpc.NetworkChange)(unsafe.Pointer(uintptr(goNetworkChanPtr)))
	(*(*rpc.Server)(jutil.Ptr(goPtr))).UnwatchNetwork(networkChan)
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeStop
func Java_io_v_impl_google_rpc_ServerImpl_nativeStop(env *C.JNIEnv, server C.jobject, goPtr C.jlong) {
	s := (*rpc.Server)(jutil.Ptr(goPtr))
	if err := (*s).Stop(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_ServerImpl_nativeFinalize
func Java_io_v_impl_google_rpc_ServerImpl_nativeFinalize(env *C.JNIEnv, server C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goPtr))
}

//export Java_io_v_impl_google_rpc_ClientImpl_nativeStartCall
func Java_io_v_impl_google_rpc_ClientImpl_nativeStartCall(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong, jContext C.jobject, jName C.jstring, jMethod C.jstring, jVomArgs C.jobjectArray, jSkipServerEndpointAuthorization C.jboolean) C.jobject {
	name := jutil.GoString(env, jName)
	method := jutil.GoString(env, jMethod)
	context, err := jcontext.GoContext(env, jContext)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	vomArgs := jutil.GoByteArrayArray(env, jVomArgs)
	// VOM-decode each arguments into a *vdl.Value.
	args := make([]interface{}, len(vomArgs))
	for i := 0; i < len(vomArgs); i++ {
		var err error
		if args[i], err = jutil.VomDecodeToValue(vomArgs[i]); err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
	}
	var opts []rpc.CallOpt
	if jSkipServerEndpointAuthorization == C.JNI_TRUE {
		opts = append(opts, options.SkipServerEndpointAuthorization{})
	}

	// Invoke StartCall
	call, err := (*(*rpc.Client)(jutil.Ptr(goPtr))).StartCall(context, name, method, args, opts...)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jCall, err := javaCall(env, call)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return jCall
}

//export Java_io_v_impl_google_rpc_ClientImpl_nativeClose
func Java_io_v_impl_google_rpc_ClientImpl_nativeClose(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong) {
	(*(*rpc.Client)(jutil.Ptr(goPtr))).Close()
}

//export Java_io_v_impl_google_rpc_ClientImpl_nativeFinalize
func Java_io_v_impl_google_rpc_ClientImpl_nativeFinalize(env *C.JNIEnv, jClient C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goPtr))
}

//export Java_io_v_impl_google_rpc_StreamImpl_nativeSend
func Java_io_v_impl_google_rpc_StreamImpl_nativeSend(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong, jVomItem C.jbyteArray) {
	vomItem := jutil.GoByteArray(env, jVomItem)
	item, err := jutil.VomDecodeToValue(vomItem)
	if err != nil {
		jutil.JThrowV(env, err)
		return
	}
	if err := (*(*rpc.Stream)(jutil.Ptr(goPtr))).Send(item); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_StreamImpl_nativeRecv
func Java_io_v_impl_google_rpc_StreamImpl_nativeRecv(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) C.jbyteArray {
	result := new(vdl.Value)
	if err := (*(*rpc.Stream)(jutil.Ptr(goPtr))).Recv(&result); err != nil {
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
	return C.jbyteArray(jutil.JByteArray(env, vomResult))
}

//export Java_io_v_impl_google_rpc_StreamImpl_nativeFinalize
func Java_io_v_impl_google_rpc_StreamImpl_nativeFinalize(env *C.JNIEnv, jStream C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goPtr))
}

//export Java_io_v_impl_google_rpc_ClientCallImpl_nativeCloseSend
func Java_io_v_impl_google_rpc_ClientCallImpl_nativeCloseSend(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	if err := (*(*rpc.ClientCall)(jutil.Ptr(goPtr))).CloseSend(); err != nil {
		jutil.JThrowV(env, err)
		return
	}
}

//export Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinish
func Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinish(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong, jNumResults C.jint) C.jobjectArray {
	// Have all the results be decoded into *vdl.Value.
	numResults := int(jNumResults)
	resultPtrs := make([]interface{}, numResults)
	for i := 0; i < numResults; i++ {
		value := new(vdl.Value)
		resultPtrs[i] = &value
	}
	if err := (*(*rpc.ClientCall)(jutil.Ptr(goPtr))).Finish(resultPtrs...); err != nil {
		// Invocation error.
		jutil.JThrowV(env, err)
		return nil
	}

	// VOM-encode the results.
	vomResults := make([][]byte, numResults)
	for i, resultPtr := range resultPtrs {
		// Remove the pointer from the result.  Simply *resultPtr doesn't work
		// as resultPtr is of type interface{}.
		result := interface{}(jutil.DerefOrDie(resultPtr))
		var err error
		if vomResults[i], err = vom.Encode(result); err != nil {
			jutil.JThrowV(env, err)
			return nil
		}
	}
	return C.jobjectArray(jutil.JByteArrayArray(env, vomResults))
}

//export Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinalize
func Java_io_v_impl_google_rpc_ClientCallImpl_nativeFinalize(env *C.JNIEnv, jCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goPtr))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeSecurity
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeSecurity(env *C.JNIEnv, jServerCallClass C.jclass, goPtr C.jlong) C.jobject {
	securityCall := (*(*rpc.ServerCall)(jutil.Ptr(goPtr))).Security()
	if securityCall == nil {
		return nil
	}
	jSecurityCall, err := jsecurity.JavaCall(env, securityCall)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jSecurityCall)
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeSuffix
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeSuffix(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*rpc.ServerCall)(jutil.Ptr(goPtr))).Suffix()))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeLocalEndpoint
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeLocalEndpoint(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*rpc.ServerCall)(jutil.Ptr(goPtr))).LocalEndpoint().String()))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeRemoteEndpoint
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeRemoteEndpoint(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jstring {
	return C.jstring(jutil.JString(env, (*(*rpc.ServerCall)(jutil.Ptr(goPtr))).RemoteEndpoint().String()))
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeGrantedBlessings
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeGrantedBlessings(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	blessings := (*(*rpc.ServerCall)(jutil.Ptr(goPtr))).GrantedBlessings()
	jBlessings, err := jsecurity.JavaBlessings(env, blessings)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jBlessings)
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeServer
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeServer(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) C.jobject {
	server := (*(*rpc.ServerCall)(jutil.Ptr(goPtr))).Server()
	jServer, err := JavaServer(env, server)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobject(jServer)
}

//export Java_io_v_impl_google_rpc_ServerCallImpl_nativeFinalize
func Java_io_v_impl_google_rpc_ServerCallImpl_nativeFinalize(env *C.JNIEnv, jServerCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goPtr))
}

//export Java_io_v_impl_google_rpc_StreamServerCallImpl_nativeFinalize
func Java_io_v_impl_google_rpc_StreamServerCallImpl_nativeFinalize(env *C.JNIEnv, jStreamServerCall C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goPtr))
}

//export Java_io_v_impl_google_rpc_AddressChooserImpl_nativeChoose
func Java_io_v_impl_google_rpc_AddressChooserImpl_nativeChoose(env *C.JNIEnv, jAddressChooser C.jobject, goPtr C.jlong, jProtocol C.jstring, jCandidates C.jobjectArray) C.jobjectArray {
	protocol := jutil.GoString(env, jProtocol)
	candidates, err := GoNetworkAddressArray(env, jCandidates)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	addrs, err := (*((*func(protocol string, candidates []net.Addr) ([]net.Addr, error))(jutil.Ptr(goPtr))))(protocol, candidates)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	jAddrs, err := JavaNetworkAddressArray(env, addrs)
	if err != nil {
		jutil.JThrowV(env, err)
		return nil
	}
	return C.jobjectArray(jAddrs)
}

//export Java_io_v_impl_google_rpc_AddressChooserImpl_nativeFinalize
func Java_io_v_impl_google_rpc_AddressChooserImpl_nativeFinalize(env *C.JNIEnv, jAddressChooser C.jobject, goPtr C.jlong) {
	jutil.GoUnref(jutil.Ptr(goPtr))
}

//export Java_io_v_impl_google_rpc_Util_nativeGoInvoker
func Java_io_v_impl_google_rpc_Util_nativeGoInvoker(env *C.JNIEnv, jUtil C.jclass, jServiceObject C.jobject) C.jlong {
	invoker, err := goInvoker(env, jServiceObject)
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	jutil.GoRef(&invoker) // Un-refed when the Go invoker is returned to the Go runtime
	return C.jlong(jutil.PtrValue(&invoker))
}

//export Java_io_v_impl_google_rpc_Util_nativeGoAuthorizer
func Java_io_v_impl_google_rpc_Util_nativeGoAuthorizer(env *C.JNIEnv, jUtil C.jclass, jAuthorizer C.jobject) C.jlong {
	auth, err := jsecurity.GoAuthorizer(env, jAuthorizer)
	if err != nil {
		jutil.JThrowV(env, err)
		return C.jlong(0)
	}
	jutil.GoRef(&auth) // Un-refed when the Go authorizer is returned to the Go runtime
	return C.jlong(jutil.PtrValue(&auth))
}
