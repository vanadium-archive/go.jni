// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vango

import (
	"fmt"
	"time"

	"v.io/v23"
	"v.io/v23/context"
	"v.io/v23/discovery"
	"v.io/v23/naming"
	"v.io/v23/rpc"
	"v.io/v23/security"
	libdiscovery "v.io/x/ref/lib/discovery"
)

const (
	rpcTimeout    = 10 * time.Second
	tcpServerName = "tmp/vango/tcp"
	btServerName  = "tmp/vango/bt"
)

var (
	bleServerName = naming.Endpoint{Protocol: "ble"}.Name()

	// vangoFuncs is a map containing go functions keys by unique strings
	// intended to be run by java/android applications using Vango.run(key).
	// Users must add function entries to this map and rebuild lib/android-lib in
	// the vanadium java repository.
	vangoFuncs = map[string]func(*context.T) error{
		"tcp-client": tcpClientFunc,
		"tcp-server": tcpServerFunc,
		"bt-client":  btClientFunc,
		"bt-server":  btServerFunc,
		"ble-client": bleClientFunc,
		"ble-server": bleServerFunc,
		"all":        AllFunc,
	}
)

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

// AllFunc runs a server, advertises it, scans for other servers and makes an
// Echo RPC to every advertised remote server.
func AllFunc(ctx *context.T) error {
	ls := rpc.ListenSpec{Proxy: "proxy"}
	addRegisteredProto(&ls, "tcp", ":0")
	addRegisteredProto(&ls, "bt", "/0")
	ctx.Infof("ListenSpec: %#v", ls)
	ctx, server, err := v23.WithNewServer(
		v23.WithListenSpec(ctx, ls),
		mountName(ctx, "all"),
		&echoServer{},
		security.AllowEveryone())
	if err != nil {
		return err
	}
	const interfaceName = "v.io/x/jni/impl/google/services/vango.EchoServer"
	ad := &discovery.Advertisement{
		InterfaceName: interfaceName,
		Attributes: discovery.Attributes{
			"Hello": "There",
		},
	}
	d, err := v23.NewDiscovery(ctx)
	if err != nil {
		return err
	}
	stoppedAd, err := libdiscovery.AdvertiseServer(ctx, d, server, "", ad, nil)
	if err != nil {
		return err
	}
	updates, err := d.Scan(ctx, "")
	if err != nil {
		return err
	}
	status := server.Status()
	ctx.Infof("My AdID: %v", ad.Id)
	ctx.Infof("Status: %+v", status)
	counter := 0
	onDiscovered := func(addr, message string) {
		ctx, cancel := context.WithTimeout(ctx, rpcTimeout)
		defer cancel()
		summary, err := runTimedCall(ctx, addr, message)
		if err != nil {
			ctx.Infof("%s: ERROR: %v", summary, err)
			return
		}
		ctx.Infof("%s: SUCCESS", summary)
	}
	for {
		select {
		case <-ctx.Done():
			ctx.Infof("EXITING")
			return nil
		case <-status.Dirty:
			status = server.Status()
			ctx.Infof("Status Changed: %+v", status)
		case u := <-updates:
			if u.IsLost() {
				ctx.Infof("LOST: %v", u.Id())
				break
			}
			counter++
			ctx.Infof("FOUND(%d): %+v", counter, u.Advertisement())
			for _, addr := range u.Addresses() {
				go onDiscovered(addr, fmt.Sprintf("CALL #%03d", counter))
			}
		case <-stoppedAd:
			ctx.Infof("Stopped advertising")
			return fmt.Errorf("stopped advertising")
		}
	}
}
