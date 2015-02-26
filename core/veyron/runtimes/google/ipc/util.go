// +build android

package ipc

import (
	"fmt"

	jutil "v.io/jni/util"
	jcontext "v.io/jni/v23/context"
	jsecurity "v.io/jni/v23/security"
	"v.io/v23/ipc"
)

// #cgo LDFLAGS: -ljniwrapper
// #include "jni_wrapper.h"
import "C"

// JavaServer converts the provided Go Server into a Java Server object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServer(jEnv interface{}, server ipc.Server, jListenSpec interface{}) (C.jobject, error) {
	if server == nil {
		return nil, fmt.Errorf("Go Server value cannot be nil")
	}
	listenSpecSign := jutil.ClassSign("io.v.v23.ipc.ListenSpec")
	jServer, err := jutil.NewObject(jEnv, jServerClass, []jutil.Sign{jutil.LongSign, listenSpecSign}, int64(jutil.PtrValue(&server)), jListenSpec)
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
	jContext, err := jcontext.JavaContext(env, call.Context(), nil)
	if err != nil {
		return nil, err
	}
	jSecurityContext, err := jsecurity.JavaContext(env, call)
	if err != nil {
		return nil, err
	}
	contextSign := jutil.ClassSign("io.v.v23.context.VContext")
	securityContextSign := jutil.ClassSign("io.v.v23.security.VContext")
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

// JavaServerStatus converts the provided ipc.ServerStatus value into a Java
// ServerStatus object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServerStatus(jEnv interface{}, status ipc.ServerStatus) (C.jobject, error) {
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
	mountStatusSign := jutil.ClassSign("io.v.v23.ipc.MountStatus")
	proxyStatusSign := jutil.ClassSign("io.v.v23.ipc.ProxyStatus")
	jServerStatus, err := jutil.NewObject(jEnv, jServerStatusClass, []jutil.Sign{serverStateSign, jutil.BoolSign, jutil.ArraySign(mountStatusSign), jutil.ArraySign(jutil.StringSign), jutil.ArraySign(proxyStatusSign)}, jState, status.ServesMountTable, jMounts, jEndpoints, jProxies)
	if err != nil {
		return nil, err
	}
	return C.jobject(jServerStatus), nil
}

// JavaServerState converts the provided ipc.ServerState value into a Java
// ServerState enum.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaServerState(jEnv interface{}, state ipc.ServerState) (C.jobject, error) {
	var name string
	switch state {
	case ipc.ServerInit:
		name = "SERVER_INIT"
	case ipc.ServerActive:
		name = "SERVER_ACTIVE"
	case ipc.ServerStopping:
		name = "SERVER_STOPPING"
	case ipc.ServerStopped:
		name = "SERVER_STOPPED"
	default:
		return nil, fmt.Errorf("Unrecognized state: %d", state)
	}
	jState, err := jutil.CallStaticObjectMethod(jEnv, jServerStateClass, "valueOf", []jutil.Sign{jutil.StringSign}, serverStateSign, name)
	if err != nil {
		return nil, err
	}
	return C.jobject(jState), nil
}

// JavaMountStatus converts the provided ipc.MountStatus value into a Java
// MountStatus object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaMountStatus(jEnv interface{}, status ipc.MountStatus) (C.jobject, error) {
	jStatus, err := jutil.NewObject(jEnv, jMountStatusClass, []jutil.Sign{jutil.StringSign, jutil.StringSign, jutil.DateTimeSign, jutil.VExceptionSign, jutil.DurationSign, jutil.DateTimeSign, jutil.VExceptionSign}, status.Name, status.Server, status.LastMount, status.LastMountErr, status.TTL, status.LastUnmount, status.LastUnmountErr)
	if err != nil {
		return nil, err
	}
	return C.jobject(jStatus), nil
}

// JavaProxyStatus converts the provided ipc.ProxyStatus value into a Java
// ProxyStatus object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaProxyStatus(jEnv interface{}, status ipc.ProxyStatus) (C.jobject, error) {
	jStatus, err := jutil.NewObject(jEnv, jProxyStatusClass, []jutil.Sign{jutil.StringSign, jutil.StringSign, jutil.VExceptionSign}, status.Proxy, status.Endpoint.String(), status.Error)
	if err != nil {
		return nil, err
	}
	return C.jobject(jStatus), nil
}

var (
	roamingSpec ipc.ListenSpec
)

// SetRoamingSpec saves the provided roaming spec for later use.
func SetRoamingSpec(spec ipc.ListenSpec) {
	roamingSpec = spec
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
	// TODO(spetrovic): fix this roaming hack, probably by implementing
	// Publisher and AddressChooser in Java as well (ugh!).
	var spec ipc.ListenSpec
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
func JavaListenSpec(jEnv interface{}, spec ipc.ListenSpec) (C.jobject, error) {
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
	addressSign := jutil.ClassSign("io.v.v23.ipc.ListenSpec$Address")
	jSpec, err := jutil.NewObject(jEnv, jListenSpecClass, []jutil.Sign{jutil.ArraySign(addressSign), jutil.StringSign, jutil.BoolSign}, jAddrs, spec.Proxy, isRoaming)
	if err != nil {
		return nil, err
	}
	return C.jobject(jSpec), nil
}

// JavaNetworkChange converts the Go NetworkChange value into a Java NetworkChange object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaNetworkChange(jEnv interface{}, change ipc.NetworkChange) (C.jobject, error) {
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
	return C.jobject(jNetworkChange), nil
}