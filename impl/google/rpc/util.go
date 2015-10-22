// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package rpc

import (
	"fmt"
	"net"
	"runtime"

	"v.io/v23/rpc"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaServer converts the provided Go Server into a Java Server object.
func JavaServer(env jutil.Env, server rpc.Server) (jutil.Object, error) {
	if server == nil {
		return jutil.NullObject, fmt.Errorf("Go Server value cannot be nil")
	}
	jServer, err := jutil.NewObject(env, jServerImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&server)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&server) // Un-refed when the Java Server object is finalized.
	return jServer, nil
}

// JavaClient converts the provided Go client into a Java Client object.
func JavaClient(env jutil.Env, client rpc.Client) (jutil.Object, error) {
	if client == nil {
		return jutil.NullObject, nil
	}
	jClient, err := jutil.NewObject(env, jClientImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&client)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&client) // Un-refed when the Java Client object is finalized.
	return jClient, nil
}

// javaStreamServerCall converts the provided Go serverCall into a Java StreamServerCall
// object.
func javaStreamServerCall(env jutil.Env, call rpc.StreamServerCall) (jutil.Object, error) {
	if call == nil {
		return jutil.NullObject, fmt.Errorf("Go StreamServerCall value cannot be nil")
	}
	jStream, err := javaStream(env, call)
	if err != nil {
		return jutil.NullObject, err
	}
	jServerCall, err := JavaServerCall(env, call)
	if err != nil {
		return jutil.NullObject, err
	}
	serverCallSign := jutil.ClassSign("io.v.v23.rpc.ServerCall")
	jStreamServerCall, err := jutil.NewObject(env, jStreamServerCallImplClass, []jutil.Sign{jutil.LongSign, streamSign, serverCallSign}, int64(jutil.PtrValue(&call)), jStream, jServerCall)
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&call) // Un-refed when the Java StreamServerCall object is finalized.
	return jStreamServerCall, nil
}

// javaCall converts the provided Go Call value into a Java Call object.
func javaCall(env jutil.Env, call rpc.ClientCall) (jutil.Object, error) {
	if call == nil {
		return jutil.NullObject, fmt.Errorf("Go Call value cannot be nil")
	}
	jStream, err := javaStream(env, call)
	if err != nil {
		return jutil.NullObject, err
	}
	jCall, err := jutil.NewObject(env, jClientCallImplClass, []jutil.Sign{jutil.LongSign, streamSign}, int64(jutil.PtrValue(&call)), jStream)
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&call) // Un-refed when the Java Call object is finalized.
	return jCall, nil
}

// javaStream converts the provided Go stream into a Java Stream object.
func javaStream(env jutil.Env, stream rpc.Stream) (jutil.Object, error) {
	jStream, err := jutil.NewObject(env, jStreamImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&stream)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&stream) // Un-refed when the Java stream object is finalized.
	return jStream, nil
}

// JavaServerStatus converts the provided rpc.ServerStatus value into a Java
// ServerStatus object.
func JavaServerStatus(env jutil.Env, status rpc.ServerStatus) (jutil.Object, error) {
	// Create Java state enum value.
	jState, err := JavaServerState(env, status.State)
	if err != nil {
		return jutil.NullObject, err
	}

	// Create Java array of mounts.
	mountarr := make([]jutil.Object, len(status.Mounts))
	for i, mount := range status.Mounts {
		var err error
		if mountarr[i], err = JavaMountStatus(env, mount); err != nil {
			return jutil.NullObject, err
		}
	}
	jMounts, err := jutil.JObjectArray(env, mountarr, jMountStatusClass)
	if err != nil {
		return jutil.NullObject, err
	}

	// Create an array of endpoint strings.
	eps := make([]string, len(status.Endpoints))
	for i, ep := range status.Endpoints {
		eps[i] = ep.String()
	}

	// Create Java array of proxies.
	proxarr := make([]jutil.Object, len(status.Proxies))
	for i, proxy := range status.Proxies {
		var err error
		if proxarr[i], err = JavaProxyStatus(env, proxy); err != nil {
			return jutil.NullObject, err
		}
	}
	jProxies, err := jutil.JObjectArray(env, proxarr, jProxyStatusClass)
	if err != nil {
		return jutil.NullObject, err
	}

	// Create final server status.
	mountStatusSign := jutil.ClassSign("io.v.v23.rpc.MountStatus")
	proxyStatusSign := jutil.ClassSign("io.v.v23.rpc.ProxyStatus")
	jServerStatus, err := jutil.NewObject(env, jServerStatusClass, []jutil.Sign{serverStateSign, jutil.BoolSign, jutil.ArraySign(mountStatusSign), jutil.ArraySign(jutil.StringSign), jutil.ArraySign(proxyStatusSign)}, jState, status.ServesMountTable, jMounts, eps, jProxies)
	if err != nil {
		return jutil.NullObject, err
	}
	return jServerStatus, nil
}

// JavaServerState converts the provided rpc.ServerState value into a Java
// ServerState enum.
func JavaServerState(env jutil.Env, state rpc.ServerState) (jutil.Object, error) {
	var name string
	switch state {
	case rpc.ServerActive:
		name = "SERVER_ACTIVE"
	case rpc.ServerStopping:
		name = "SERVER_STOPPING"
	case rpc.ServerStopped:
		name = "SERVER_STOPPED"
	default:
		return jutil.NullObject, fmt.Errorf("Unrecognized state: %d", state)
	}
	jState, err := jutil.CallStaticObjectMethod(env, jServerStateClass, "valueOf", []jutil.Sign{jutil.StringSign}, serverStateSign, name)
	if err != nil {
		return jutil.NullObject, err
	}
	return jState, nil
}

// JavaMountStatus converts the provided rpc.MountStatus value into a Java
// MountStatus object.
func JavaMountStatus(env jutil.Env, status rpc.MountStatus) (jutil.Object, error) {
	jStatus, err := jutil.NewObject(env, jMountStatusClass, []jutil.Sign{jutil.StringSign, jutil.StringSign, jutil.DateTimeSign, jutil.VExceptionSign, jutil.DurationSign, jutil.DateTimeSign, jutil.VExceptionSign}, status.Name, status.Server, status.LastMount, status.LastMountErr, status.TTL, status.LastUnmount, status.LastUnmountErr)
	if err != nil {
		return jutil.NullObject, err
	}
	return jStatus, nil
}

// JavaProxyStatus converts the provided rpc.ProxyStatus value into a Java
// ProxyStatus object.
func JavaProxyStatus(env jutil.Env, status rpc.ProxyStatus) (jutil.Object, error) {
	jStatus, err := jutil.NewObject(env, jProxyStatusClass, []jutil.Sign{jutil.StringSign, jutil.StringSign, jutil.VExceptionSign}, status.Proxy, status.Endpoint.String(), status.Error)
	if err != nil {
		return jutil.NullObject, err
	}
	return jStatus, nil
}

// GoListenSpec converts the provided Java ListenSpec into a Go ListenSpec.
func GoListenSpec(env jutil.Env, jSpec jutil.Object) (rpc.ListenSpec, error) {
	jAddrs, err := jutil.CallObjectMethod(env, jSpec, "getAddresses", nil, jutil.ArraySign(listenAddrSign))
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	addrs, err := GoListenAddrs(env, jAddrs)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	proxy, err := jutil.CallStringMethod(env, jSpec, "getProxy", nil)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	jChooser, err := jutil.CallObjectMethod(env, jSpec, "getChooser", nil, addressChooserSign)
	if err != nil {
		return rpc.ListenSpec{}, err
	}
	chooser := GoAddressChooser(env, jChooser)
	return rpc.ListenSpec{
		Addrs:          addrs,
		Proxy:          proxy,
		AddressChooser: chooser,
	}, nil
}

// GoListenSpec converts the provided Go ListenSpec into a Java ListenSpec.
func JavaListenSpec(env jutil.Env, spec rpc.ListenSpec) (jutil.Object, error) {
	jAddrs, err := JavaListenAddrArray(env, spec.Addrs)
	if err != nil {
		return jutil.NullObject, err
	}
	jChooser, err := JavaAddressChooser(env, spec.AddressChooser)
	if err != nil {
		return jutil.NullObject, err
	}
	addressSign := jutil.ClassSign("io.v.v23.rpc.ListenSpec$Address")
	jSpec, err := jutil.NewObject(env, jListenSpecClass, []jutil.Sign{jutil.ArraySign(addressSign), jutil.StringSign, addressChooserSign}, jAddrs, spec.Proxy, jChooser)
	if err != nil {
		return jutil.NullObject, err
	}
	return jSpec, nil
}

// JavaListenAddrArray converts Go rpc.ListenAddrs into a Java
// ListenSpec$Address array.
func JavaListenAddrArray(env jutil.Env, addrs rpc.ListenAddrs) (jutil.Object, error) {
	addrarr := make([]jutil.Object, len(addrs))
	for i, addr := range addrs {
		var err error
		if addrarr[i], err = jutil.NewObject(env, jListenSpecAddressClass, []jutil.Sign{jutil.StringSign, jutil.StringSign}, addr.Protocol, addr.Address); err != nil {
			return jutil.NullObject, err
		}
	}
	return jutil.JObjectArray(env, addrarr, jListenSpecAddressClass)
}

// GoListenAddrs converts Java ListenSpec$Address array into a Go
// rpc.ListenAddrs value.
func GoListenAddrs(env jutil.Env, jAddrs jutil.Object) (rpc.ListenAddrs, error) {
	addrarr, err := jutil.GoObjectArray(env, jAddrs)
	if err != nil {
		return nil, err
	}
	addrs := make(rpc.ListenAddrs, len(addrarr))
	for i, jAddr := range addrarr {
		var err error
		addrs[i].Protocol, err = jutil.CallStringMethod(env, jAddr, "getProtocol", nil)
		if err != nil {
			return nil, err
		}
		addrs[i].Address, err = jutil.CallStringMethod(env, jAddr, "getAddress", nil)
		if err != nil {
			return nil, err
		}
	}
	return addrs, nil
}

// JavaServerCall converts a Go rpc.ServerCall into a Java ServerCall object.
func JavaServerCall(env jutil.Env, serverCall rpc.ServerCall) (jutil.Object, error) {
	jServerCall, err := jutil.NewObject(env, jServerCallImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&serverCall)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&serverCall) // Un-refed when the Java ServerCall object is finalized.
	return jServerCall, nil
}

// JavaNetworkAddress converts a Go net.Addr into a Java NetworkAddress object.
func JavaNetworkAddress(env jutil.Env, addr net.Addr) (jutil.Object, error) {
	return jutil.NewObject(env, jNetworkAddressClass, []jutil.Sign{jutil.StringSign, jutil.StringSign}, addr.Network(), addr.String())
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
func GoNetworkAddress(env jutil.Env, jAddr jutil.Object) (net.Addr, error) {
	network, err := jutil.CallStringMethod(env, jAddr, "network", nil)
	if err != nil {
		return nil, err
	}
	addr, err := jutil.CallStringMethod(env, jAddr, "address", nil)
	if err != nil {
		return nil, err
	}
	return &jniAddr{network, addr}, nil
}

// JavaNetworkAddressArray converts a Go slice of net.Addr values into a Java
// array of NetworkAddress objects.
func JavaNetworkAddressArray(env jutil.Env, addrs []net.Addr) (jutil.Object, error) {
	arr := make([]jutil.Object, len(addrs))
	for i, addr := range addrs {
		var err error
		if arr[i], err = JavaNetworkAddress(env, addr); err != nil {
			return jutil.NullObject, err
		}
	}
	return jutil.JObjectArray(env, arr, jNetworkAddressClass)
}

// GoNetworkAddressArray converts a Java array of NetworkAddress objects into a
// Go slice of net.Addr values.
func GoNetworkAddressArray(env jutil.Env, jAddrs jutil.Object) ([]net.Addr, error) {
	arr, err := jutil.GoObjectArray(env, jAddrs)
	if err != nil {
		return nil, err
	}
	ret := make([]net.Addr, len(arr))
	for i, jAddr := range arr {
		var err error
		if ret[i], err = GoNetworkAddress(env, jAddr); err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// JavaAddressChooser converts a Go address chooser function into a Java
// AddressChooser object.
func JavaAddressChooser(env jutil.Env, chooser rpc.AddressChooser) (jutil.Object, error) {
	jAddressChooser, err := jutil.NewObject(env, jAddressChooserImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&chooser)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&chooser) // Un-refed when the Java AddressChooser object is finalized.
	return jAddressChooser, nil
}

type jniAddressChooser struct {
	jChooser jutil.Object
}

func (chooser *jniAddressChooser) ChooseAddresses(protocol string, candidates []net.Addr) ([]net.Addr, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jCandidates, err := JavaNetworkAddressArray(env, candidates)
	if err != nil {
		return nil, err
	}
	addrsSign := jutil.ArraySign(jutil.ClassSign("io.v.v23.rpc.NetworkAddress"))
	jAddrs, err := jutil.CallObjectMethod(env, chooser.jChooser, "choose", []jutil.Sign{jutil.StringSign, addrsSign}, addrsSign, protocol, jCandidates)
	if err != nil {
		return nil, err
	}
	return GoNetworkAddressArray(env, jAddrs)
}

// GoAddressChooser converts a Java AddressChooser object into a Go address
// chooser function.
func GoAddressChooser(env jutil.Env, jChooser jutil.Object) rpc.AddressChooser {
	// Reference Java chooser; it will be de-referenced when the go function
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	chooser := &jniAddressChooser{
		jChooser: jutil.NewGlobalRef(env, jChooser),
	}
	runtime.SetFinalizer(chooser, func(chooser *jniAddressChooser) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, chooser.jChooser)
	})
	return chooser
}
