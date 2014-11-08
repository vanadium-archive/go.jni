// +build android

package ipc

import (
	"fmt"
	"unsafe"

	jutil "veyron.io/jni/util"
	jcontext "veyron.io/jni/veyron2/context"
	jsecurity "veyron.io/jni/veyron2/security"
	"veyron.io/veyron/veyron/profiles/roaming"
	"veyron.io/veyron/veyron2/ipc"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaServer converts the provided Go Server into a Java Server object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServer(jEnv interface{}, server ipc.Server) (C.jobject, error) {
	if server == nil {
		return nil, nil
	}
	jServer, err := jutil.NewObject(jEnv, jServerClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&server)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&server) // Un-refed when the Java Server object is finalized.
	return C.jobject(jServer), nil
}

// JavaClient converts the provided Go client into a Java Client object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaClient(jEnv interface{}, client ipc.Client) (C.jobject, error) {
	if client == nil {
		return nil, nil
	}
	c := newClient(client)
	jClient, err := jutil.NewObject(jEnv, jClientClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(c)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(c) // Un-refed when the Java Client object is finalized.
	return C.jobject(jClient), nil
}

// javaServerCall converts the provided Go serverCall into a Java ServerCall
// object.
func javaServerCall(env *C.JNIEnv, call *serverCall) (C.jobject, error) {
	if call == nil {
		return nil, nil
	}
	jStream, err := javaStream(env, call)
	if err != nil {
		return nil, err
	}
	jContext, err := jcontext.JavaContext(env, call, nil)
	if err != nil {
		return nil, err
	}
	jSecurityContext, err := jsecurity.JavaContext(env, call)
	if err != nil {
		return nil, err
	}
	contextSign := jutil.ClassSign("io.veyron.veyron.veyron2.context.Context")
	securityContextSign := jutil.ClassSign("io.veyron.veyron.veyron2.security.Context")
	jServerCall, err := jutil.NewObject(env, jServerCallClass, []jutil.Sign{jutil.LongSign, streamSign, contextSign, securityContextSign}, int64(jutil.PtrValue(call)), jStream, jContext, jSecurityContext)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(call) // Un-refed when the Java ServerCall object is finalized.
	return C.jobject(jServerCall), nil
}

// javaServerCall converts the provided Go clientCall into a Java ClientCall
// object.
func javaClientCall(env *C.JNIEnv, call *clientCall) (C.jobject, error) {
	if call == nil {
		return nil, nil
	}
	jStream, err := javaStream(env, call)
	if err != nil {
		return nil, err
	}
	jClientCall, err := jutil.NewObject(env, jClientCallClass, []jutil.Sign{jutil.LongSign, streamSign}, int64(jutil.PtrValue(call)), jStream)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(call) // Un-refed when the Java ClientCall object is finalized.
	return C.jobject(jClientCall), nil
}

// javaStream converts the provided Go stream into a Java Stream object.
func javaStream(env *C.JNIEnv, streamIn interface{}) (C.jobject, error) {
	s := (*stream)(unsafe.Pointer(jutil.PtrValue(streamIn)))
	if s == nil {
		return nil, nil
	}
	jStream, err := jutil.NewObject(env, jStreamClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(s)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(s) // Un-refed when the Java stream object is finalized.
	return C.jobject(jStream), nil
}

// GoListenSpec converts the provided Java ListenSpec into a Go ListenSpec.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoListenSpec(jEnv, jSpec interface{}) (ipc.ListenSpec, error) {
	protocol, err := jutil.CallStringMethod(jEnv, jSpec, "getProtocol", nil)
	if err != nil {
		return ipc.ListenSpec{}, err
	}
	proxy, err := jutil.CallStringMethod(jEnv, jSpec, "getProxy", nil)
	if err != nil {
		return ipc.ListenSpec{}, err
	}
	var spec ipc.ListenSpec
	switch protocol {
	case "tcp", "tcp4", "tcp6":
		spec = roaming.ListenSpec
	default:
		return ipc.ListenSpec{}, fmt.Errorf("Unsupported protocol: %s", protocol)
	}
	spec.Proxy = proxy
	return spec, nil
}
