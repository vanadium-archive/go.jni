// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vango

import (
	"fmt"
	"io"
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
	vangoFuncs = map[string]func(*context.T, io.Writer) error{
		"tcp-client": tcpClientFunc,
		"tcp-server": tcpServerFunc,
		"bt-client":  btClientFunc,
		"bt-server":  btServerFunc,
		"ble-client": bleClientFunc,
		"ble-server": bleServerFunc,
		"all":        AllFunc,
	}
)

func tcpServerFunc(ctx *context.T, _ io.Writer) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Proxy: "proxy"})
	return runServer(ctx, tcpServerName)
}

func tcpClientFunc(ctx *context.T, _ io.Writer) error {
	return runClient(ctx, tcpServerName)
}

func btServerFunc(ctx *context.T, _ io.Writer) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Addrs: rpc.ListenAddrs{{Protocol: "bt", Address: "/0"}}})
	return runServer(ctx, btServerName)
}

func btClientFunc(ctx *context.T, _ io.Writer) error {
	return runClient(ctx, btServerName)
}

func bleServerFunc(ctx *context.T, _ io.Writer) error {
	ctx = v23.WithListenSpec(ctx, rpc.ListenSpec{Addrs: rpc.ListenAddrs{{Protocol: "ble", Address: "na"}}})
	return runServer(ctx, "")
}

func bleClientFunc(ctx *context.T, _ io.Writer) error {
	return runClient(ctx, bleServerName)
}

// AllFunc runs a server, advertises it, scans for other servers and makes an
// Echo RPC to every advertised remote server.
func AllFunc(ctx *context.T, output io.Writer) error {
	ls := rpc.ListenSpec{Proxy: "proxy"}
	addRegisteredProto(&ls, "tcp", ":0")
	addRegisteredProto(&ls, "bt", "/0")
	fmt.Fprintf(output, "Listening on: %+v (and proxy)", ls.Addrs)
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
	var (
		status      = server.Status()
		counter     = 0
		peerByAdId  = make(map[discovery.AdId]*peer)
		lastCall    = make(map[discovery.AdId]time.Time)
		callResults = make(chan string)
		activeCalls = 0
		quit        = false
		myaddrs     = serverAddrs(status)
		ticker      = time.NewTicker(time.Second)
		call        = func(p *peer) {
			counter++
			activeCalls++
			lastCall[p.adId] = time.Now()
			go func(msg string) {
				summary, err := p.call(ctx, msg)
				if err != nil {
					callResults <- fmt.Sprintf("ERROR calling [%v]: %v", p.description, err)
					return
				}
				callResults <- summary
			}(fmt.Sprintf("Hello #%d", counter))
		}
	)
	defer ticker.Stop()
	fmt.Fprintln(output, "My AdID:", ad.Id)
	fmt.Fprintln(output, "My addrs:", myaddrs)
	ctx.Infof("SERVER STATUS: %+v", status)
	for !quit {
		select {
		case <-ctx.Done():
			quit = true
		case <-status.Dirty:
			status = server.Status()
			newaddrs := serverAddrs(status)
			changed := len(newaddrs) != len(myaddrs)
			if !changed {
				for i := range newaddrs {
					if newaddrs[i] != myaddrs[i] {
						changed = true
						break
					}
				}
			}
			if changed {
				myaddrs = newaddrs
				fmt.Fprintln(output, "My addrs:", myaddrs)
			}
			ctx.Infof("SERVER STATUS: %+v", status)
		case u, scanning := <-updates:
			if !scanning {
				fmt.Fprintln(output, "SCANNING STOPPED")
				quit = true
				break
			}
			if u.IsLost() {
				if p, ok := peerByAdId[u.Id()]; ok {
					fmt.Fprintln(output, "LOST:", p.description)
				}
				delete(peerByAdId, u.Id())
				delete(lastCall, u.Id())
				break
			}
			p, err := newPeer(ctx, u)
			if err != nil {
				ctx.Info(err)
				break
			}
			peerByAdId[p.adId] = p
			fmt.Fprintln(output, "FOUND:", p.description)
			call(p)
		case r := <-callResults:
			activeCalls--
			fmt.Fprintln(output, r)
		case <-stoppedAd:
			fmt.Fprintln(output, "STOPPED ADVERTISING")
			stoppedAd = nil
		case <-ticker.C:
			// Call all peers that haven't been called in a while
			now := time.Now()
			for id, t := range lastCall {
				if now.Sub(t) > rpcTimeout {
					call(peerByAdId[id])
				}
			}
		}
	}
	fmt.Println(output, "EXITING: Cleaning up")
	for activeCalls > 0 {
		<-callResults
		activeCalls--
	}
	// Exhaust the scanned updates queue.
	// (The channel will be closed as a by-product of the context being Done).
	for range updates {
	}
	fmt.Fprintln(output, "EXITING: Done")
	return nil
}
