// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package vango

import (
	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/rpc"
)

const (
	tcpServerName = "tmp/tcpServerName"
	btServerName  = "tmp/btServerName"
)

var bleServerName = naming.Endpoint{
	Protocol: "ble",
}.Name()

// vangoFuncs is a map containing go functions keys by unique strings
// intended to be run by java/android applications using Vango.run(key).
// Users must add function entries to this map and rebuild lib/android-lib in
// the vanadium java repository.
var vangoFuncs = map[string]func(*context.T) error{
	"tcp-client": tcpClientFunc,
	"tcp-server": tcpServerFunc,
	"bt-client":  btClientFunc,
	"bt-server":  btServerFunc,
	"ble-client": bleClientFunc,
	"ble-server": bleServerFunc,
}

func tcpServerFunc(ctx *context.T) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Proxy: "proxy"})
	return runServer(ctx, tcpServerName)
}

func tcpClientFunc(ctx *context.T) error {
	return runClient(ctx, tcpServerName)
}

func btServerFunc(ctx *context.T) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Addrs: rpc.ListenAddrs{{Protocol: "bt", Address: "/0"}}})
	return runServer(ctx, btServerName)
}

func btClientFunc(ctx *context.T) error {
	return runClient(ctx, btServerName)
}

func bleServerFunc(ctx *context.T) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Addrs: rpc.ListenAddrs{{Protocol: "ble", Address: "na"}}})
	return runServer(ctx, "")
}

func bleClientFunc(ctx *context.T) error {
	return runClient(ctx, bleServerName)
}
