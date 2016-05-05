// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package v23_go_runner

import (
	"fmt"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/rpc"
	"v.io/v23/security"
)

const (
	tcpServerName = "tmp/tcpServerName"
)

// v23GoRunnerFuncs is a map containing go functions keys by unique strings
// intended to be run by java/android applications using V23GoRunner.run(key).
// Users must add function entries to this map and rebuild lib/android-lib in
// the vanadium java repository.
var v23GoRunnerFuncs = map[string]func(*context.T) error{
	"tcp-server": tcpServerFunc,
	"tcp-client": tcpClientFunc,
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
	ctx.Info("Client successufl executed rpc")
	return nil
}
