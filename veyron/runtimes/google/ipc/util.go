// +build android

package ipc

import (
	"unsafe"

	"veyron.io/jni/util"
	jsecurity "veyron.io/jni/veyron/runtimes/google/security"
	jcontext "veyron.io/jni/veyron2/context"
	"veyron.io/veyron/veyron2/ipc"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// javaServer converts the provided Go Server into a Java Server object.
func javaServer(env *C.JNIEnv, server ipc.Server) (C.jobject, error) {
	if server == nil {
		return nil, nil
	}
	jServer, err := util.NewObject(env, jServerClass, []util.Sign{util.LongSign}, int64(util.PtrValue(&server)))
	if err != nil {
		return nil, err
	}
	util.GoRef(&server) // Un-refed when the Java Server object is finalized.
	return C.jobject(jServer), nil
}

// javaClient converts the provided Go client into a Java Client object.
func javaClient(env *C.JNIEnv, c *client) (C.jobject, error) {
	if c == nil {
		return nil, nil
	}
	jClient, err := util.NewObject(env, jClientClass, []util.Sign{util.LongSign}, int64(util.PtrValue(c)))
	if err != nil {
		return nil, err
	}
	util.GoRef(c) // Un-refed when the Java Client object is finalized.
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
	contextSign := util.ClassSign("io.veyron.veyron.veyron2.context.Context")
	securityContextSign := util.ClassSign("io.veyron.veyron.veyron.runtimes.google.security.Context")
	jServerCall, err := util.NewObject(env, jServerCallClass, []util.Sign{util.LongSign, streamSign, contextSign, securityContextSign}, int64(util.PtrValue(call)), jStream, jContext, jSecurityContext)
	if err != nil {
		return nil, err
	}
	util.GoRef(call) // Un-refed when the Java ServerCall object is finalized.
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
	jClientCall, err := util.NewObject(env, jClientCallClass, []util.Sign{util.LongSign, streamSign}, int64(util.PtrValue(call)), jStream)
	if err != nil {
		return nil, err
	}
	util.GoRef(call) // Un-refed when the Java ClientCall object is finalized.
	return C.jobject(jClientCall), nil
}

// javaStream converts the provided Go stream into a Java Stream object.
func javaStream(env *C.JNIEnv, streamIn interface{}) (C.jobject, error) {
	s := (*stream)(unsafe.Pointer(util.PtrValue(streamIn)))
	if s == nil {
		return nil, nil
	}
	jStream, err := util.NewObject(env, jStreamClass, []util.Sign{util.LongSign}, C.jlong(util.PtrValue(s)))
	if err != nil {
		return nil, err
	}
	util.GoRef(s) // Un-refed when the Java stream object is finalized.
	return C.jobject(jStream), nil
}
