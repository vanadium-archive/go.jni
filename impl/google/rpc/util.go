// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"fmt"
	"unsafe"

	"v.io/v23/rpc"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaServer converts the provided Go Server into a Java Server object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServer(jEnv interface{}, server rpc.Server, jListenSpec interface{}) (unsafe.Pointer, error) {
	if server == nil {
		return nil, fmt.Errorf("Go Server value cannot be nil")
	}
	listenSpecSign := jutil.ClassSign("io.v.v23.rpc.ListenSpec")
	jServer, err := jutil.NewObject(jEnv, jServerClass, []jutil.Sign{jutil.LongSign, listenSpecSign}, int64(jutil.PtrValue(&server)), jListenSpec)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&server) // Un-refed when the Java Server object is finalized.
	return jServer, nil
}

// JavaClient converts the provided Go client into a Java Client object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaClient(jEnv interface{}, client rpc.Client) (unsafe.Pointer, error) {
	if client == nil {
		return nil, fmt.Errorf("Go Client value cannot be nil")
	}
	jClient, err := jutil.NewObject(jEnv, jClientClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&client)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&client) // Un-refed when the Java Client object is finalized.
	return jClient, nil
}

// javaStreamServerCall converts the provided Go serverCall into a Java StreamServerCall
// object.
func javaStreamServerCall(env *C.JNIEnv, call rpc.StreamServerCall) (C.jobject, error) {
	if call == nil {
		return nil, fmt.Errorf("Go StreamServerCall value cannot be nil")
	}
	jStream, err := javaStream(env, call)
	if err != nil {
		return nil, err
	}
	jServerCall, err := JavaServerCall(env, call)
	if err != nil {
		return nil, err
	}
	serverCallSign := jutil.ClassSign("io.v.impl.google.rpc.ServerCall")
	jStreamServerCall, err := jutil.NewObject(env, jStreamServerCallClass, []jutil.Sign{jutil.LongSign, streamSign, serverCallSign}, int64(jutil.PtrValue(&call)), jStream, jServerCall)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&call) // Un-refed when the Java StreamServerCall object is finalized.
	return C.jobject(jStreamServerCall), nil
}

// javaCall converts the provided Go Call value into a Java Call object.
func javaCall(env *C.JNIEnv, call rpc.ClientCall) (C.jobject, error) {
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
func javaStream(env *C.JNIEnv, stream rpc.Stream) (C.jobject, error) {
	jStream, err := jutil.NewObject(env, jStreamClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&stream)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&stream) // Un-refed when the Java stream object is finalized.
	return C.jobject(jStream), nil
}

// JavaServerStatus converts the provided rpc.ServerStatus value into a Java
// ServerStatus object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServerStatus(jEnv interface{}, status rpc.ServerStatus) (unsafe.Pointer, error) {
	// Create Java state enum value.
	jState, err := JavaServerState(jEnv, status.State)
	if err != nil {
		return nil, err
	}

	// Create Java array of mounts.
	mountarr := make([]interface{}, len(status.Mounts))
	for i, mount := range status.Mounts {
		var err error
		if mountarr[i], err = JavaMountStatus(jEnv, mount); err != nil {
			return nil, err
		}
	}
	jMounts := jutil.JObjectArray(jEnv, mountarr, jMountStatusClass)

	// Create Java array of endpoints.
	eps := make([]string, len(status.Endpoints))
	for i, ep := range status.Endpoints {
		eps[i] = ep.String()
	}
	jEndpoints := jutil.JStringArray(jEnv, eps)

	// Create Java array of proxies.
	proxarr := make([]interface{}, len(status.Proxies))
	for i, proxy := range status.Proxies {
		var err error
		if proxarr[i], err = JavaProxyStatus(jEnv, proxy); err != nil {
			return nil, err
		}
	}
	jProxies := jutil.JObjectArray(jEnv, proxarr, jProxyStatusClass)

	// Create final server status.
	mountStatusSign := jutil.ClassSign("io.v.v23.rpc.MountStatus")
	proxyStatusSign := jutil.ClassSign("io.v.v23.rpc.ProxyStatus")
	jServerStatus, err := jutil.NewObject(jEnv, jServerStatusClass, []jutil.Sign{serverStateSign, jutil.BoolSign, jutil.ArraySign(mountStatusSign), jutil.ArraySign(jutil.StringSign), jutil.ArraySign(proxyStatusSign)}, jState, status.ServesMountTable, jMounts, jEndpoints, jProxies)
	if err != nil {
		return nil, err
	}
	return jServerStatus, nil
}

// JavaServerState converts the provided rpc.ServerState value into a Java
// ServerState enum.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServerState(jEnv interface{}, state rpc.ServerState) (unsafe.Pointer, error) {
	var name string
	switch state {
	case rpc.ServerInit:
		name = "SERVER_INIT"
	case rpc.ServerActive:
		name = "SERVER_ACTIVE"
	case rpc.ServerStopping:
		name = "SERVER_STOPPING"
	case rpc.ServerStopped:
		name = "SERVER_STOPPED"
	default:
		return nil, fmt.Errorf("Unrecognized state: %d", state)
	}
	jState, err := jutil.CallStaticObjectMethod(jEnv, jServerStateClass, "valueOf", []jutil.Sign{jutil.StringSign}, serverStateSign, name)
	if err != nil {
		return nil, err
	}
	return jState, nil
}

// JavaMountStatus converts the provided rpc.MountStatus value into a Java
// MountStatus object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaMountStatus(jEnv interface{}, status rpc.MountStatus) (unsafe.Pointer, error) {
	jStatus, err := jutil.NewObject(jEnv, jMountStatusClass, []jutil.Sign{jutil.StringSign, jutil.StringSign, jutil.DateTimeSign, jutil.VExceptionSign, jutil.DurationSign, jutil.DateTimeSign, jutil.VExceptionSign}, status.Name, status.Server, status.LastMount, status.LastMountErr, status.TTL, status.LastUnmount, status.LastUnmountErr)
	if err != nil {
		return nil, err
	}
	return jStatus, nil
}

// JavaProxyStatus converts the provided rpc.ProxyStatus value into a Java
// ProxyStatus object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaProxyStatus(jEnv interface{}, status rpc.ProxyStatus) (unsafe.Pointer, error) {
	jStatus, err := jutil.NewObject(jEnv, jProxyStatusClass, []jutil.Sign{jutil.StringSign, jutil.StringSign, jutil.VExceptionSign}, status.Proxy, status.Endpoint.String(), status.Error)
	if err != nil {
		return nil, err
	}
	return jStatus, nil
}

var (
	roamingSpec rpc.ListenSpec
)

// SetRoamingSpec saves the provided roaming spec for later use.
func SetRoamingSpec(spec rpc.ListenSpec) {
	roamingSpec = spec
}

// GoListenSpec converts the provided Java ListenSpec into a Go ListenSpec.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoListenSpec(jEnv, jSpec interface{}) (rpc.ListenSpec, error) {
	addrArr, err := jutil.CallObjectArrayMethod(jEnv, jSpec, "getAddresses", nil, listenAddrSign)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	addrs := make(rpc.ListenAddrs, len(addrArr))
	for i, jAddr := range addrArr {
		var err error
		addrs[i].Protocol, err = jutil.CallStringMethod(jEnv, jAddr, "getProtocol", nil)
		if err != nil {
			return rpc.ListenSpec{}, err
		}
		addrs[i].Address, err = jutil.CallStringMethod(jEnv, jAddr, "getAddress", nil)
		if err != nil {
			return rpc.ListenSpec{}, err
		}
	}
	proxy, err := jutil.CallStringMethod(jEnv, jSpec, "getProxy", nil)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	isRoaming, err := jutil.CallBooleanMethod(jEnv, jSpec, "isRoaming", nil)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	// TODO(spetrovic): fix this roaming hack, probably by implementing
	// Publisher and AddressChooser in Java as well (ugh!).
	var spec rpc.ListenSpec
	if isRoaming {
		spec = roamingSpec
	}
	spec.Addrs = addrs
	spec.Proxy = proxy
	return spec, nil
}

// GoListenSpec converts the provided Go ListenSpec into a Java ListenSpec.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaListenSpec(jEnv interface{}, spec rpc.ListenSpec) (unsafe.Pointer, error) {
	addrarr := make([]interface{}, len(spec.Addrs))
	for i, addr := range spec.Addrs {
		var err error
		if addrarr[i], err = jutil.NewObject(jEnv, jListenSpecAddressClass, []jutil.Sign{jutil.StringSign, jutil.StringSign}, addr.Protocol, addr.Address); err != nil {
			return nil, err
		}
	}
	jAddrs := jutil.JObjectArray(jEnv, addrarr, jListenSpecAddressClass)
	isRoaming := false
	if spec.StreamPublisher != nil || spec.AddressChooser != nil {
		// Our best guess that this is a roaming spec.
		isRoaming = true
	}
	addressSign := jutil.ClassSign("io.v.v23.rpc.ListenSpec$Address")
	jSpec, err := jutil.NewObject(jEnv, jListenSpecClass, []jutil.Sign{jutil.ArraySign(addressSign), jutil.StringSign, jutil.BoolSign}, jAddrs, spec.Proxy, isRoaming)
	if err != nil {
		return nil, err
	}
	return jSpec, nil
}

// JavaNetworkChange converts the Go NetworkChange value into a Java NetworkChange object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaNetworkChange(jEnv interface{}, change rpc.NetworkChange) (unsafe.Pointer, error) {
	jTime, err := jutil.JTime(jEnv, change.Time)
	if err != nil {
		return nil, err
	}
	jState, err := JavaServerState(jEnv, change.State)
	if err != nil {
		return nil, err
	}
	setting := fmt.Sprintf("%v", change.Setting)
	changedEndpointStrs := make([]string, len(change.Changed))
	for i, ep := range change.Changed {
		changedEndpointStrs[i] = fmt.Sprintf("%v", ep)
	}
	jNetworkChange, err := jutil.NewObject(jEnv, jNetworkChangeClass, []jutil.Sign{jutil.DateTimeSign, serverStateSign, jutil.ArraySign(jutil.StringSign), jutil.StringSign, jutil.VExceptionSign}, jTime, jState, changedEndpointStrs, setting, change.Error)
	if err != nil {
		return nil, err
	}
	return jNetworkChange, nil
}

// JavaServerCall converts a Go rpc.ServerCall into a Java ServerCall object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServerCall(jEnv interface{}, serverCall rpc.ServerCall) (C.jobject, error) {
	jServerCall, err := jutil.NewObject(jEnv, jServerCallClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&serverCall)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&serverCall) // Un-refed when the Java ServerCall object is finalized.
	return C.jobject(jServerCall), nil
}
