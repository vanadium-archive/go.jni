// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package plugins

import (
	"runtime"

	"v.io/v23/context"

	idiscovery "v.io/x/ref/lib/discovery"

	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

func NewBlePluginFactory(env jutil.Env, jAndroidContext jutil.Object) func(string) (idiscovery.Plugin, error) {
	return newPluginFactory(env, jAndroidContext, jBlePluginClass)
}

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

	jNativeScanHandler, err := jutil.NewObject(env, jNativeScanHandlerClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&ch)))
	if err != nil {
		done()
		return err
	}
	err = jutil.CallVoidMethod(env, p.jPlugin, "startScan", []jutil.Sign{jutil.StringSign, scanHandlerSign}, interfaceName, jNativeScanHandler)
	if err != nil {
		done()
		return err
	}

	jutil.GoRef(&ch) // Will be unrefed when jNativeScanHandler is finalized.
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

func newPluginFactory(env jutil.Env, jAndroidContext jutil.Object, jPluginClass jutil.Class) func(string) (idiscovery.Plugin, error) {
	// Reference Android Context; it will be de-referenced when the factory
	// is garbage-collected.
	jAndroidContext = jutil.NewGlobalRef(env, jAndroidContext)
	factory := func(host string) (idiscovery.Plugin, error) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()

		jHost := jutil.JString(env, host)
		jPlugin, err := jutil.NewObject(env, jPluginClass, []jutil.Sign{jutil.StringSign, androidContextSign}, jHost, jAndroidContext)
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
	runtime.SetFinalizer(&factory, func(*func(string) (idiscovery.Plugin, error)) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, jAndroidContext)
	})
	return factory
}
