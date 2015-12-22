// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package bt

import (
	"net"
	"runtime"
	"time"

	"v.io/v23/context"
	"v.io/v23/flow"

	jutil "v.io/x/jni/util"
)

var (
	connectionSign = jutil.ClassSign("io.v.android.impl.google.rpc.protocols.bt.Bluetooth$Connection")
	listenerSign   = jutil.ClassSign("io.v.android.impl.google.rpc.protocols.bt.Bluetooth$Listener")

	// Global reference for io.v.impl.google.rpc.protocols.bt.Bluetooth class.
	jBluetoothClass jutil.Class
)

// Init initializes the JNI code with the given Java environment. This method
// must be called from the main Java thread.
func Init(env jutil.Env) error {
	// Cache global references to all Java classes used by the package.  This is
	// necessary because JNI gets access to the class loader only in the system
	// thread, so we aren't able to invoke FindClass in other threads.
	var err error
	jBluetoothClass, err = jutil.JFindClass(env, "io/v/android/impl/google/rpc/protocols/bt/Bluetooth")
	if err != nil {
		return err
	}
	flow.RegisterProtocol("bt", btProtocol{})
	return nil
}

type btProtocol struct{}

func (btProtocol) Dial(ctx *context.T, protocol, address string, timeout time.Duration) (flow.Conn, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jConnection, err := jutil.CallStaticObjectMethod(env, jBluetoothClass, "dial", []jutil.Sign{jutil.StringSign, jutil.DurationSign}, connectionSign, address, timeout)
	if err != nil {
		return nil, err
	}
	return newConnection(env, jConnection), nil
}

func (btProtocol) Resolve(ctx *context.T, protocol, address string) (string, string, error) {
	return protocol, address, nil
}

func (btProtocol) Listen(ctx *context.T, protocol, address string) (flow.Listener, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jListener, err := jutil.CallStaticObjectMethod(env, jBluetoothClass, "listen", []jutil.Sign{jutil.StringSign}, listenerSign, address)
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
	l := &btListener{jListener}
	runtime.SetFinalizer(l, func(l *btListener) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, l.jListener)
	})
	return l
}

type btListener struct {
	jListener jutil.Object
}

func (l *btListener) Accept(ctx *context.T) (flow.Conn, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jConn, err := jutil.CallObjectMethod(env, l.jListener, "accept", nil, connectionSign)
	if err != nil {
		return nil, err
	}
	return newConnection(env, jConn), nil
}

func (l *btListener) Addr() net.Addr {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	addr, err := jutil.CallStringMethod(env, l.jListener, "address", nil)
	if err != nil {
		return &btAddr{""}
	}
	return &btAddr{addr}
}

func (l *btListener) Close() error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	return jutil.CallVoidMethod(env, l.jListener, "close", nil)
}

func newConnection(env jutil.Env, jConnection jutil.Object) flow.Conn {
	// Reference Java Connection; it will be de-referenced when the Go Conn
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jConnection = jutil.NewGlobalRef(env, jConnection)
	c := &btConn{jConnection}
	runtime.SetFinalizer(c, func(c *btConn) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, c.jConnection)
	})
	return c
}

type btConn struct {
	jConnection jutil.Object
}

func (c *btConn) ReadMsg() ([]byte, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	return jutil.CallByteArrayMethod(env, c.jConnection, "read", nil)
}

func (c *btConn) WriteMsg(bs ...[]byte) (n int, err error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	total := 0
	for _, b := range bs {
		if err := jutil.CallVoidMethod(env, c.jConnection, "write", []jutil.Sign{jutil.ByteArraySign}, b); err != nil {
			return 0, err
		}
		total += len(b)
	}
	return total, nil
}

func (c *btConn) Close() error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	return jutil.CallVoidMethod(env, c.jConnection, "close", nil)
}

func (c *btConn) LocalAddr() net.Addr {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	addr, err := jutil.CallStringMethod(env, c.jConnection, "localAddress", nil)
	if err != nil {
		return &btAddr{""}
	}
	return &btAddr{addr}
}

type btAddr struct {
	addr string
}

func (a *btAddr) Network() string {
	return "bt"
}

func (a *btAddr) String() string {
	return a.addr
}
