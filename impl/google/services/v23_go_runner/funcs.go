// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package v23_go_runner

import (
	"fmt"
	"time"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/rpc"
	"v.io/v23/security"
)

const (
	tcpServerName = "tmp/tcpServerName"
	btServerName  = "tmp/btServerName"
)

var bleServerName = naming.Endpoint{
	Protocol: "ble",
}.Name()

// v23GoRunnerFuncs is a map containing go functions keys by unique strings
// intended to be run by java/android applications using V23GoRunner.run(key).
// Users must add function entries to this map and rebuild lib/android-lib in
// the vanadium java repository.
var v23GoRunnerFuncs = map[string]func(*context.T) error{
	"tcp-server": tcpServerFunc,
	"tcp-client": tcpClientFunc,
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

func runServer(ctx *context.T, name string) error {
	_, server, err := v23.WithNewServer(ctx, name, &echoServer{}, security.AllowEveryone())
	if err != nil {
		return err
	}
	ctx.Infof("Server listening on %v", server.Status().Endpoints)
	ctx.Infof("Server listen errors: %v", server.Status().ListenErrors)
	return nil
}

func runClient(ctx *context.T, name string) error {
	elapsed, err := runTimedCall(ctx, name)
	if err != nil {
		return err
	}
	ctx.Infof("Client successfully executed rpc on new connection in %s.", elapsed.String())

	elapsed, err = runTimedCall(ctx, name)
	if err != nil {
		return err
	}
	ctx.Infof("Client successfully executed rpc on cached connection in %s.", elapsed.String())
	return nil
}

func runTimedCall(ctx *context.T, name string) (time.Duration, error) {
	message := "hi there"
	var got string
	start := time.Now()
	if err := v23.GetClient(ctx).Call(ctx, name, "Echo", []interface{}{message}, []interface{}{&got}); err != nil {
		return 0, err
	}
	elapsed := time.Now().Sub(start)
	if want := message; got != want {
		return 0, fmt.Errorf("got %s, want %s", got, want)
	}
	return elapsed, nil
}

func bleServerFunc(ctx *context.T) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Addrs: rpc.ListenAddrs{{Protocol: "ble"}}})
	return runServer(ctx, "")
}

func bleClientFunc(ctx *context.T) error {
	return runClient(ctx, bleServerName)
}
