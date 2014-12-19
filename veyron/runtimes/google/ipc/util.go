// +build android

package ipc

import (
	"fmt"

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
func JavaServer(jEnv interface{}, server ipc.Server, jOptions interface{}) (C.jobject, error) {
	if server == nil {
		return nil, fmt.Errorf("Go Server value cannot be nil")
	}
	jServer, err := jutil.NewObject(jEnv, jServerClass, []jutil.Sign{jutil.LongSign, optionsSign}, int64(jutil.PtrValue(&server)), jOptions)
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
		return nil, fmt.Errorf("Go Client value cannot be nil")
	}
	jClient, err := jutil.NewObject(jEnv, jClientClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&client)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&client) // Un-refed when the Java Client object is finalized.
	return C.jobject(jClient), nil
}

// javaServerCall converts the provided Go serverCall into a Java ServerCall
// object.
func javaServerCall(env *C.JNIEnv, call ipc.ServerCall) (C.jobject, error) {
	if call == nil {
		return nil, fmt.Errorf("Go ServerCall value cannot be nil")
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
	jServerCall, err := jutil.NewObject(env, jServerCallClass, []jutil.Sign{jutil.LongSign, streamSign, contextSign, securityContextSign}, int64(jutil.PtrValue(&call)), jStream, jContext, jSecurityContext)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&call) // Un-refed when the Java ServerCall object is finalized.
	return C.jobject(jServerCall), nil
}

// javaCall converts the provided Go Call value into a Java Call object.
func javaCall(env *C.JNIEnv, call ipc.Call) (C.jobject, error) {
	if call == nil {
		return nil, fmt.Errorf("Go Call value cannot be nil")
	}
	jStream, err := javaStream(env, call)
	if err != nil {
		return nil, err
	}
	jCall, err := jutil.NewObject(env, jCallClass, []jutil.Sign{jutil.LongSign, streamSign}, int64(jutil.PtrValue(&call)), jStream)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&call) // Un-refed when the Java Call object is finalized.
	return C.jobject(jCall), nil
}

// javaStream converts the provided Go stream into a Java Stream object.
func javaStream(env *C.JNIEnv, stream ipc.Stream) (C.jobject, error) {
	jStream, err := jutil.NewObject(env, jStreamClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&stream)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&stream) // Un-refed when the Java stream object is finalized.
	return C.jobject(jStream), nil
}

// GoListenSpec converts the provided Java ListenSpec into a Go ListenSpec.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoListenSpec(jEnv, jSpec interface{}) (ipc.ListenSpec, error) {
	addrArr, err := jutil.CallObjectArrayMethod(jEnv, jSpec, "getAddresses", nil, listenAddrSign)
	if err != nil {
		return ipc.ListenSpec{}, err
	}
	addrs := make(ipc.ListenAddrs, len(addrArr))
	for i, jAddr := range addrArr {
		var err error
		addrs[i].Protocol, err = jutil.CallStringMethod(jEnv, jAddr, "getProtocol", nil)
		if err != nil {
			return ipc.ListenSpec{}, err
		}
		addrs[i].Address, err = jutil.CallStringMethod(jEnv, jAddr, "getAddress", nil)
		if err != nil {
			return ipc.ListenSpec{}, err
		}
	}
	proxy, err := jutil.CallStringMethod(jEnv, jSpec, "getProxy", nil)
	if err != nil {
		return ipc.ListenSpec{}, err
	}
	isRoaming, err := jutil.CallBooleanMethod(jEnv, jSpec, "isRoaming", nil)
	if err != nil {
		return ipc.ListenSpec{}, err
	}
	var spec ipc.ListenSpec
	if isRoaming {
		spec = roaming.ListenSpec
	}
	spec.Addrs = addrs
	spec.Proxy = proxy
	return spec, nil
}
