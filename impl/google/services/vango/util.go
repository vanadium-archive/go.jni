// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vango

import (
	"fmt"
	"time"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/conventions"
	"v.io/v23/flow"
	"v.io/v23/naming"
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
	summary, err := runTimedCall(ctx, name, "new connection")
	if err != nil {
		return err
	}
	ctx.Infof("Client success: %v", summary)

	summary, err = runTimedCall(ctx, name, "cached connection")
	if err != nil {
		return err
	}
	ctx.Infof("Client success: %v", summary)
	return nil
}

func runTimedCall(ctx *context.T, name, message string) (string, error) {
	summary := fmt.Sprintf("[%s] to %v", message, name)
	start := time.Now()
	call, err := v23.GetClient(ctx).StartCall(ctx, name, "Echo", []interface{}{message})
	if err != nil {
		return summary, err
	}
	var recvd string
	if err := call.Finish(&recvd); err != nil {
		return summary, err
	}
	elapsed := time.Now().Sub(start)
	if recvd != message {
		return summary, fmt.Errorf("got [%s], want [%s]", recvd, message)
	}
	me := security.LocalBlessingNames(ctx, call.Security())
	them, _ := call.RemoteBlessings()
	return fmt.Sprintf("%s in %v (THEM:%v EP:%v) (ME:%v)", summary, elapsed, them, call.Security().RemoteEndpoint(), me), nil
}

func mountName(ctx *context.T, addendums ...string) string {
	var (
		p     = v23.GetPrincipal(ctx)
		b, _  = p.BlessingStore().Default()
		names = conventions.ParseBlessingNames(security.BlessingNames(p, b)...)
	)
	if len(names) == 0 {
		return ""
	}
	return naming.Join(append([]string{names[0].Home()}, addendums...)...)
}

func addRegisteredProto(ls *rpc.ListenSpec, proto, addr string) {
	for _, p := range flow.RegisteredProtocols() {
		if p == proto {
			ls.Addrs = append(ls.Addrs, rpc.ListenAddrs{{Protocol: p, Address: addr}}...)
		}
	}
}
