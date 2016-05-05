// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package ble

import (
	"net"
	"runtime"
	"time"

	"v.io/v23/context"
	"v.io/v23/flow"
	"v.io/x/ref/runtime/protocols/lib/framer"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

var (
	contextSign  = jutil.ClassSign("io.v.v23.context.VContext")
	streamSign   = jutil.ClassSign("io.v.android.impl.google.rpc.protocols.ble.BLE$Stream")
	listenerSign = jutil.ClassSign("io.v.android.impl.google.rpc.protocols.ble.BLE$Listener")

	// Global reference for io.v.impl.google.rpc.protocols.ble.BLE class.
	jBleClass jutil.Class
)

const ble = "ble"

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
func Init(env jutil.Env) error {
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	var err error
	jBleClass, err = jutil.JFindClass(env, "io/v/android/impl/google/rpc/protocols/ble/BLE")
	if err != nil {
		return err
	}
	flow.RegisterProtocol(ble, bleProtocol{})
	return nil
}

// TODO(mattr): We should unify this struct with btProtocol by introducing Protocol java interfaces and allowing
// you to initialize one with Init and a class name.  I'll keep this separate until I get ble working.
type bleProtocol struct{}

func (bleProtocol) Dial(ctx *context.T, protocol, address string, timeout time.Duration) (flow.Conn, error) {
	env, freeFunc := jutil.GetEnv()
	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		freeFunc()
		return nil, err
	}
	// This method will invoke the freeFunc().
	jStream, err := jutil.CallStaticCallbackMethod(env, freeFunc, jBleClass, "dial", []jutil.Sign{contextSign, jutil.StringSign, jutil.DurationSign}, jContext, address, timeout)
	if err != nil {
		return nil, err
	}
	env, freeFunc = jutil.GetEnv()
	defer freeFunc()
	return newConnection(env, jStream), nil
}

func (bleProtocol) Resolve(ctx *context.T, protocol, address string) (string, []string, error) {
	return protocol, []string{address}, nil
}

func (bleProtocol) Listen(ctx *context.T, protocol, address string) (flow.Listener, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		return nil, err
	}
	jListener, err := jutil.CallStaticObjectMethod(env, jBleClass, "listen", []jutil.Sign{contextSign, jutil.StringSign}, listenerSign, jContext, address)
	if err != nil {
		return nil, err
	}
	return newListener(env, jListener), nil
}

func newListener(env jutil.Env, jListener jutil.Object) flow.Listener {
	// Reference Java Listener; it will be de-referenced when the Go Listener
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jListener = jutil.NewGlobalRef(env, jListener)
	l := &bleListener{jListener}
	runtime.SetFinalizer(l, func(l *bleListener) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, l.jListener)
	})
	return l
}

type bleListener struct {
	jListener jutil.Object
}

func (l *bleListener) Accept(ctx *context.T) (flow.Conn, error) {
	env, freeFunc := jutil.GetEnv()
	// This method will invoke the freeFunc().
	jStream, err := jutil.CallCallbackMethod(env, freeFunc, l.jListener, "accept", nil)
	if err != nil {
		return nil, err
	}
	env, freeFunc = jutil.GetEnv()
	defer freeFunc()
	return newConnection(env, jStream), nil
}

func (l *bleListener) Addr() net.Addr {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	addr, err := jutil.CallStringMethod(env, l.jListener, "address", nil)
	if err != nil {
		return bleAddr("")
	}
	return bleAddr(addr)
}

func (l *bleListener) Close() error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	return jutil.CallVoidMethod(env, l.jListener, "close", nil)
}

// newConnection creates a new Go connection.  The passed-in Java Stream object
// is assumed to hold a global reference.
func newConnection(env jutil.Env, jStream jutil.Object) flow.Conn {
	c := &bleReadWriteCloser{jStream}
	runtime.SetFinalizer(c, func(c *bleReadWriteCloser) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, c.jStream)
	})
	addrStr, _ := jutil.CallStringMethod(env, jStream, "localAddress", nil)
	localAddr := bleAddr(addrStr)
	return bleConn{framer.New(c), localAddr}
}

type bleConn struct {
	flow.MsgReadWriteCloser
	localAddr net.Addr
}

func (c bleConn) LocalAddr() net.Addr {
	return c.localAddr
}

type bleReadWriteCloser struct {
	jStream jutil.Object
}

func (c *bleReadWriteCloser) Read(b []byte) (n int, err error) {
	env, freeFunc := jutil.GetEnv()
	// This method will invoke the freeFunc().
	jResult, err := jutil.CallCallbackMethod(env, freeFunc, c.jStream, "read", []jutil.Sign{jutil.IntSign}, len(b))
	if err != nil {
		return 0, err
	}
	env, freeFunc = jutil.GetEnv()
	defer freeFunc()
	defer jutil.DeleteGlobalRef(env, jResult)
	data := jutil.GoByteArray(env, jResult)
	return copy(b, data), nil
}

func (c *bleReadWriteCloser) Write(b []byte) (n int, err error) {
	env, freeFunc := jutil.GetEnv()
	// This method will invoke the freeFunc().
	if _, err := jutil.CallCallbackMethod(env, freeFunc, c.jStream, "write", []jutil.Sign{jutil.ByteArraySign}, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *bleReadWriteCloser) Close() error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	return jutil.CallVoidMethod(env, c.jStream, "close", nil)
}

type bleAddr string

func (bleAddr) Network() string {
	return ble
}

func (a bleAddr) String() string {
	return string(a)
}
