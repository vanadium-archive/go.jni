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
	jcontext "v.io/x/jni/v23/context"
	"v.io/x/ref/runtime/protocols/lib/framer"
)

const (
	bluetoothWithPortClassName = "BluetoothWithPort"
	bluetoothWithSdpClassName  = "BluetoothWithSdp"

	// TODO(jhahn,suharshs): Which one is more reliable?
	btClassName = bluetoothWithSdpClassName
)

var (
	contextSign  = jutil.ClassSign("io.v.v23.context.VContext")
	streamSign   = jutil.ClassSign("io.v.android.impl.google.rpc.protocols.bt." + btClassName + "$Stream")
	listenerSign = jutil.ClassSign("io.v.android.impl.google.rpc.protocols.bt." + btClassName + "$Listener")

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
	jBluetoothClass, err = jutil.JFindClass(env, "io/v/android/impl/google/rpc/protocols/bt/"+btClassName)
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

	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		return nil, err
	}
	jStream, err := jutil.CallStaticObjectMethod(env, jBluetoothClass, "dial", []jutil.Sign{contextSign, jutil.StringSign, jutil.DurationSign}, streamSign, jContext, address, timeout)
	if err != nil {
		return nil, err
	}
	// Reference Java Stream; it will be de-referenced when the Go connection
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jStream = jutil.NewGlobalRef(env, jStream)
	return newConnection(env, jStream), nil
}

func (btProtocol) Resolve(ctx *context.T, protocol, address string) (string, []string, error) {
	return protocol, []string{address}, nil
}

func (btProtocol) Listen(ctx *context.T, protocol, address string) (flow.Listener, error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		return nil, err
	}
	jListener, err := jutil.CallStaticObjectMethod(env, jBluetoothClass, "listen", []jutil.Sign{contextSign, jutil.StringSign}, listenerSign, jContext, address)
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

	jStream, err := jutil.CallObjectMethod(env, l.jListener, "accept", nil, streamSign)
	if err != nil {
		return nil, err
	}
	// Reference Java Stream; it will be de-referenced when the Go connection
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jStream = jutil.NewGlobalRef(env, jStream)
	return newConnection(env, jStream), nil
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

// newConnection creates a new Go connection.  The passed-in Java Stream object
// is assumed to hold a global reference.
func newConnection(env jutil.Env, jStream jutil.Object) flow.Conn {
	c := &btReadWriteCloser{jStream}
	runtime.SetFinalizer(c, func(c *btReadWriteCloser) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, c.jStream)
	})
	addrStr, _ := jutil.CallStringMethod(env, jStream, "localAddress", nil)
	localAddr := &btAddr{addrStr}
	return btConn{framer.New(c), localAddr}
}

type btConn struct {
	flow.MsgReadWriteCloser
	localAddr net.Addr
}

func (c btConn) LocalAddr() net.Addr {
	return c.localAddr
}

type btReadWriteCloser struct {
	jStream jutil.Object
}

func (c *btReadWriteCloser) Read(b []byte) (n int, err error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	data, err := jutil.CallByteArrayMethod(env, c.jStream, "read", []jutil.Sign{jutil.IntSign}, len(b))
	if err != nil {
		return 0, err
	}
	return copy(b, data), nil
}

func (c *btReadWriteCloser) Write(b []byte) (n int, err error) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	if err := jutil.CallVoidMethod(env, c.jStream, "write", []jutil.Sign{jutil.ByteArraySign}, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *btReadWriteCloser) Close() error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	return jutil.CallVoidMethod(env, c.jStream, "close", nil)
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
