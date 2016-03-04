// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package plugins

import (
	"runtime"

	"v.io/v23/context"

	idiscovery "v.io/x/ref/lib/discovery"
	dfactory "v.io/x/ref/lib/discovery/factory"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

type plugin struct {
	jPlugin jutil.Object
	trigger *idiscovery.Trigger
}

func (p *plugin) Advertise(ctx *context.T, adinfo *idiscovery.AdInfo, done func()) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jAdInfo, err := jutil.JVomCopy(env, adinfo, jAdInfoClass)
	if err != nil {
		done()
		return err
	}
	err = jutil.CallVoidMethod(env, p.jPlugin, "startAdvertising", []jutil.Sign{adInfoSign}, jAdInfo)
	if err != nil {
		done()
		return err
	}

	jAdInfo = jutil.NewGlobalRef(env, jAdInfo)
	stop := func() {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()

		jutil.CallVoidMethod(env, p.jPlugin, "stopAdvertising", []jutil.Sign{adInfoSign}, jAdInfo)
		jutil.DeleteGlobalRef(env, jAdInfo)
		done()
	}
	p.trigger.Add(stop, ctx.Done())
	return nil
}

func (p *plugin) Scan(ctx *context.T, interfaceName string, ch chan<- *idiscovery.AdInfo, done func()) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	chRef := jutil.GoNewRef(&ch) // Un-refed when jNativeScanHandler is finalized.
	jNativeScanHandler, err := jutil.NewObject(env, jNativeScanHandlerClass, []jutil.Sign{jutil.LongSign}, int64(chRef))
	if err != nil {
		jutil.GoDecRef(chRef)
		done()
		return err
	}
	err = jutil.CallVoidMethod(env, p.jPlugin, "startScan", []jutil.Sign{jutil.StringSign, scanHandlerSign}, interfaceName, jNativeScanHandler)
	if err != nil {
		done()
		return err
	}

	jNativeScanHandler = jutil.NewGlobalRef(env, jNativeScanHandler)
	stop := func() {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()

		jutil.CallVoidMethod(env, p.jPlugin, "stopScan", []jutil.Sign{scanHandlerSign}, jNativeScanHandler)
		jutil.DeleteGlobalRef(env, jNativeScanHandler)
		done()
	}
	p.trigger.Add(stop, ctx.Done())
	return nil
}

func newPluginFactory(env jutil.Env, jPluginClass jutil.Class) func(*context.T, string) (idiscovery.Plugin, error) {
	return func(ctx *context.T, host string) (idiscovery.Plugin, error) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()

		// We pass the global context of the android vanadium runtime, since the context of the discovery
		// factory does not have an Android context.
		jCtx, err := jcontext.JavaContext(env, ctx, nil)
		if err != nil {
			return nil, err
		}
		jPlugin, err := jutil.NewObject(env, jPluginClass, []jutil.Sign{contextSign, jutil.StringSign}, jCtx, host)
		if err != nil {
			return nil, err
		}
		// Reference Plugin; it will be de-referenced when the plugin is garbage-collected.
		jPlugin = jutil.NewGlobalRef(env, jPlugin)
		p := &plugin{
			jPlugin: jPlugin,
			trigger: idiscovery.NewTrigger(),
		}
		runtime.SetFinalizer(p, func(*plugin) {
			env, freeFunc := jutil.GetEnv()
			jutil.DeleteGlobalRef(env, p.jPlugin)
			freeFunc()
		})
		return p, nil
	}
}

func initPluginFactories(env jutil.Env) error {
	jPluginClass, err := jutil.JFindClass(env, "io/v/android/impl/google/discovery/plugins/ble/BlePlugin")
	if err != nil {
		return err
	}
	dfactory.SetPluginFactory("ble", newPluginFactory(env, jPluginClass))
	return nil
}
