// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package v23_go_runner

import (
	"v.io/v23/context"
	"v.io/v23/rpc"
)

type echoServer struct{}

func (*echoServer) Echo(ctx *context.T, call rpc.ServerCall, arg string) (string, error) {
	ctx.Infof("echoServer got message '%s' from %v", arg, call.RemoteEndpoint())
	return arg, nil
}
