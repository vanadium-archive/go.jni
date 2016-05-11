// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package vango

import (
	"fmt"
	"time"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/rpc"
	"v.io/v23/security"
)

type echoServer struct{}

func (*echoServer) Echo(ctx *context.T, call rpc.ServerCall, arg string) (string, error) {
	ctx.Infof("echoServer got message '%s' from %v", arg, call.RemoteEndpoint())
	return arg, nil
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
