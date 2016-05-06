// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package v23_go_runner

import (
	"fmt"
	"strconv"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/naming"
	"v.io/v23/rpc"
	"v.io/v23/security"
)

const (
	tcpServerName = "tmp/tcpServerName"

	// TODO(suharshs): Currently we hard code the server phone's MAC address, email blessing, and port number because:
	// (1) We haven't plugged in discovery yet.
	// (2) Android APIs have removed a way to get the MAC address of a server.
	btAddress   = "a0:91:69:99:00:f0"
	btPortNum   = 13
	btBlessings = "dev.v.io:o:608941808256-43vtfndets79kf5hac8ieujto8837660.apps.googleusercontent.com:vanadium.testphone@gmail.com"
)

var btServerName = naming.Endpoint{
	Protocol: "bt",
	Address:  btAddress + "/" + strconv.Itoa(btPortNum),
}.WithBlessingNames([]string{btBlessings}).String()

// v23GoRunnerFuncs is a map containing go functions keys by unique strings
// intended to be run by java/android applications using V23GoRunner.run(key).
// Users must add function entries to this map and rebuild lib/android-lib in
// the vanadium java repository.
var v23GoRunnerFuncs = map[string]func(*context.T) error{
	"tcp-server": tcpServerFunc,
	"tcp-client": tcpClientFunc,
	"bt-client":  btClientFunc,
	"bt-server":  btServerFunc,
}

func tcpServerFunc(ctx *context.T) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Proxy: "proxy"})
	if _, _, err := v23.WithNewServer(ctx, tcpServerName, &echoServer{}, security.AllowEveryone()); err != nil {
		return err
	}
	return nil
}

func tcpClientFunc(ctx *context.T) error {
	message := "hi there"
	var got string
	if err := v23.GetClient(ctx).Call(ctx, tcpServerName, "Echo", []interface{}{message}, []interface{}{&got}); err != nil {
		return err
	}
	if want := message; got != want {
		return fmt.Errorf("got %s, want %s", got, want)
	}
	ctx.Info("Client successfully executed rpc")
	return nil
}

func btServerFunc(ctx *context.T) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Addrs: rpc.ListenAddrs{{Protocol: "bt", Address: "/" + strconv.Itoa(btPortNum)}}})
	_, server, err := v23.WithNewServer(ctx, "", &echoServer{}, security.AllowEveryone())
	if err != nil {
		return err
	}
	ctx.Infof("Server listening on %v", server.Status().Endpoints)
	ctx.Infof("Server listen errors: %v", server.Status().ListenErrors)
	return nil
}

func btClientFunc(ctx *context.T) error {
	message := "hi there"
	var got string
	if err := v23.GetClient(ctx).Call(ctx, btServerName, "Echo", []interface{}{message}, []interface{}{&got}); err != nil {
		return err
	}
	if want := message; got != want {
		return fmt.Errorf("got %s, want %s", got, want)
	}
	ctx.Info("Client successfully executed rpc")
	return nil
}
