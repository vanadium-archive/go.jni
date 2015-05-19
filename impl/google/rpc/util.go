// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"fmt"
	"net"
	"runtime"
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
func JavaServer(jEnv interface{}, server rpc.Server) (unsafe.Pointer, error) {
	if server == nil {
		return nil, fmt.Errorf("Go Server value cannot be nil")
	}
	jServer, err := jutil.NewObject(jEnv, jServerImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&server)))
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
	jClient, err := jutil.NewObject(jEnv, jClientImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&client)))
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
	serverCallSign := jutil.ClassSign("io.v.v23.rpc.ServerCall")
	jStreamServerCall, err := jutil.NewObject(env, jStreamServerCallImplClass, []jutil.Sign{jutil.LongSign, streamSign, serverCallSign}, int64(jutil.PtrValue(&call)), jStream, jServerCall)
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
	jCall, err := jutil.NewObject(env, jClientCallImplClass, []jutil.Sign{jutil.LongSign, streamSign}, int64(jutil.PtrValue(&call)), jStream)
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&call) // Un-refed when the Java Call object is finalized.
	return C.jobject(jCall), nil
}

// javaStream converts the provided Go stream into a Java Stream object.
func javaStream(env *C.JNIEnv, stream rpc.Stream) (C.jobject, error) {
	jStream, err := jutil.NewObject(env, jStreamImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&stream)))
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

// GoListenSpec converts the provided Java ListenSpec into a Go ListenSpec.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoListenSpec(jEnv, jSpec interface{}) (rpc.ListenSpec, error) {
	jAddrs, err := jutil.CallObjectMethod(jEnv, jSpec, "getAddresses", nil, jutil.ArraySign(listenAddrSign))
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	addrs, err := GoListenAddrs(jEnv, jAddrs)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	proxy, err := jutil.CallStringMethod(jEnv, jSpec, "getProxy", nil)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	jChooser, err := jutil.CallObjectMethod(jEnv, jSpec, "getChooser", nil, addressChooserSign)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	chooser := GoAddressChooser(jEnv, jChooser)
	return rpc.ListenSpec{
		Addrs:          addrs,
		Proxy:          proxy,
		AddressChooser: chooser,
	}, nil
}

// GoListenSpec converts the provided Go ListenSpec into a Java ListenSpec.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaListenSpec(jEnv interface{}, spec rpc.ListenSpec) (unsafe.Pointer, error) {
	jAddrs, err := JavaListenAddrArray(jEnv, spec.Addrs)
	if err != nil {
		return nil, err
	}
	jProxy := jutil.JString(jEnv, spec.Proxy)
	jChooser, err := JavaAddressChooser(jEnv, spec.AddressChooser)
	if err != nil {
		return nil, err
	}
	addressSign := jutil.ClassSign("io.v.v23.rpc.ListenSpec$Address")
	jSpec, err := jutil.NewObject(jEnv, jListenSpecClass, []jutil.Sign{jutil.ArraySign(addressSign), jutil.StringSign, addressChooserSign}, jAddrs, jProxy, jChooser)
	if err != nil {
		return nil, err
	}
	return jSpec, nil
}

// JavaListenAddrArray converts Go rpc.ListenAddrs into a Java
// ListenSpec$Address array.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaListenAddrArray(jEnv interface{}, addrs rpc.ListenAddrs) (unsafe.Pointer, error) {
	addrarr := make([]interface{}, len(addrs))
	for i, addr := range addrs {
		var err error
		if addrarr[i], err = jutil.NewObject(jEnv, jListenSpecAddressClass, []jutil.Sign{jutil.StringSign, jutil.StringSign}, addr.Protocol, addr.Address); err != nil {
			return nil, err
		}
	}
	return jutil.JObjectArray(jEnv, addrarr, jListenSpecAddressClass), nil
}

// GoListenAddrs converts Java ListenSpec$Address array into a Go
// rpc.ListenAddrs value.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoListenAddrs(jEnv, jAddrs interface{}) (rpc.ListenAddrs, error) {
	addrarr := jutil.GoObjectArray(jEnv, jAddrs)
	addrs := make(rpc.ListenAddrs, len(addrarr))
	for i, jAddr := range addrarr {
		var err error
		addrs[i].Protocol, err = jutil.CallStringMethod(jEnv, jAddr, "getProtocol", nil)
		if err != nil {
			return nil, err
		}
		addrs[i].Address, err = jutil.CallStringMethod(jEnv, jAddr, "getAddress", nil)
		if err != nil {
			return nil, err
		}
	}
	return addrs, nil
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
	jAddedAddrs, err := JavaNetworkAddressArray(jEnv, change.AddedAddrs)
	if err != nil {
		return nil, err
	}
	jRemovedAddrs, err := JavaNetworkAddressArray(jEnv, change.RemovedAddrs)
	if err != nil {
		return nil, err
	}
	changedEndpointStrs := make([]string, len(change.Changed))
	for i, ep := range change.Changed {
		changedEndpointStrs[i] = fmt.Sprintf("%v", ep)
	}
	addrSign := jutil.ClassSign("io.v.v23.rpc.NetworkAddress")
	jNetworkChange, err := jutil.NewObject(jEnv, jNetworkChangeClass, []jutil.Sign{jutil.DateTimeSign, serverStateSign, jutil.ArraySign(addrSign), jutil.ArraySign(addrSign), jutil.ArraySign(jutil.StringSign), jutil.VExceptionSign}, jTime, jState, jAddedAddrs, jRemovedAddrs, changedEndpointStrs, change.Error)
	if err != nil {
		return nil, err
	}
	return jNetworkChange, nil
}

// JavaServerCall converts a Go rpc.ServerCall into a Java ServerCall object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServerCall(jEnv interface{}, serverCall rpc.ServerCall) (unsafe.Pointer, error) {
	jServerCall, err := jutil.NewObject(jEnv, jServerCallImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&serverCall)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&serverCall) // Un-refed when the Java ServerCall object is finalized.
	return jServerCall, nil
}

// JavaNetworkAddress converts a Go net.Addr into a Java NetworkAddress object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaNetworkAddress(jEnv interface{}, addr net.Addr) (unsafe.Pointer, error) {
	return jutil.NewObject(jEnv, jNetworkAddressClass, []jutil.Sign{jutil.StringSign, jutil.StringSign}, addr.Network(), addr.String())
}

type jniAddr struct {
	network, addr string
}

func (a *jniAddr) Network() string {
	return a.network
}

func (a *jniAddr) String() string {
	return a.addr
}

// GoNetworkAddress converts a Java NetworkAddress object into Go net.Addr.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoNetworkAddress(jEnv, jAddr interface{}) (net.Addr, error) {
	network, err := jutil.CallStringMethod(jEnv, jAddr, "network", nil)
	if err != nil {
		return nil, err
	}
	addr, err := jutil.CallStringMethod(jEnv, jAddr, "address", nil)
	if err != nil {
		return nil, err
	}
	return &jniAddr{network, addr}, nil
}

// JavaNetworkAddressArray converts a Go slice of net.Addr values into a Java
// array of NetworkAddress objects.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaNetworkAddressArray(jEnv interface{}, addrs []net.Addr) (unsafe.Pointer, error) {
	arr := make([]interface{}, len(addrs))
	for i, addr := range addrs {
		var err error
		if arr[i], err = JavaNetworkAddress(jEnv, addr); err != nil {
			return nil, err
		}
	}
	return jutil.JObjectArray(jEnv, arr, jNetworkAddressClass), nil
}

// GoNetworkAddressArray converts a Java array of NetworkAddress objects into a
// Go slice of net.Addr values.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoNetworkAddressArray(jEnv, jAddrs interface{}) ([]net.Addr, error) {
	arr := jutil.GoObjectArray(jEnv, jAddrs)
	ret := make([]net.Addr, len(arr))
	for i, jAddr := range arr {
		var err error
		if ret[i], err = GoNetworkAddress(jEnv, jAddr); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// JavaAddressChooser converts a Go address chooser function into a Java
// AddressChooser object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaAddressChooser(jEnv interface{}, chooser func(protocol string, candidates []net.Addr) ([]net.Addr, error)) (unsafe.Pointer, error) {
	jAddressChooser, err := jutil.NewObject(jEnv, jAddressChooserImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&chooser)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&chooser) // Un-refed when the Java AddressChooser object is finalized.
	return jAddressChooser, nil
}

// GoAddressChooser converts a Java AddressChooser object into a Go address
// chooser function.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoAddressChooser(jEnv, jChooserObj interface{}) func(protocol string, candidates []net.Addr) ([]net.Addr, error) {
	// Reference Java chooser; it will be de-referenced when the go function
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jChooser := jutil.NewGlobalRef(jEnv, jChooserObj)
	f := func(protocol string, candidates []net.Addr) ([]net.Addr, error) {
		javaEnv, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jCandidates, err := JavaNetworkAddressArray(javaEnv, candidates)
		if err != nil {
			return nil, err
		}
		addrsSign := jutil.ArraySign(jutil.ClassSign("io.v.v23.rpc.NetworkAddress"))
		jAddrs, err := jutil.CallObjectMethod(javaEnv, jChooser, "choose", []jutil.Sign{jutil.StringSign, addrsSign}, addrsSign, protocol, jCandidates)
		if err != nil {
			return nil, err
		}
		return GoNetworkAddressArray(javaEnv, jAddrs)
	}
	runtime.SetFinalizer(&f, func(f *func(protocol string, candidates []net.Addr) ([]net.Addr, error)) {
		javaEnv, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(javaEnv, jChooser)
	})
	return f
}
