// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package discovery
import (
	"bytes"
	"encoding/binary"
	"runtime"

	"v.io/v23/context"

	"v.io/x/ref/lib/discovery"

	jcontext "v.io/x/jni/v23/context"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

func NewBleCreator(env jutil.Env, context jutil.Object) func(string) (discovery.Plugin, error) {
	// Reference Android Context; it will be de-referenced when the plugin
	// created below is garbage-collected (through the finalizer callback we
	// setup in the function below).  This function should only be executed once.
	jContext := jutil.NewGlobalRef(env, context)
	return func(host string) (discovery.Plugin, error) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jPlugin, err := jutil.NewObject(env, jBlePluginClass, []jutil.Sign{androidContextSign}, jContext)

		jutil.DeleteGlobalRef(env, jContext)
		if err != nil {
			return nil, err
		}
		// Reference Android BlePlugin; it will be de-referenced when the plugin
		// created below is garbage-collected (through the finalizer callback we
		// setup below).
		jPlugin = jutil.NewGlobalRef(env, jPlugin)
		p := &plugin{
			trigger: discovery.NewTrigger(),
			jPlugin: jPlugin,
		}
		runtime.SetFinalizer(p, func(p *plugin) {
			env, freeFunc := jutil.GetEnv()
			defer freeFunc()
			jutil.DeleteGlobalRef(env, p.jPlugin)
		})
		return p, nil
	}

}

type plugin struct {
	trigger *discovery.Trigger
	jPlugin jutil.Object
}

func (p *plugin) Advertise(ctx *context.T, ad discovery.Advertisement, done func()) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jContext, err := jcontext.JavaContext(env, ctx)
	if err != nil {
		return err
	}
	jAdv, err := jutil.JVomCopy(env, ad, jAdvertisementClass)
	if err != nil {
		return err
	}

	err = jutil.CallVoidMethod(env, p.jPlugin, "addAdvertisement", []jutil.Sign{contextSign, advertisementSign}, jContext, jAdv)

	p.trigger.Add(done, ctx.Done())

	return err
}

func (p *plugin) Scan(ctx *context.T, serviceUuid discovery.Uuid, ch chan<- discovery.Advertisement, done func()) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jContext, err := jcontext.JavaContext(env, ctx)
	if err != nil {
		return err
	}

	jUuid, err := JavaUUID(env, serviceUuid)

	if err != nil {
		return err
	}


	jutil.GoRef(&ch)
	jNativeScanHandler, err := jutil.NewObject(env, jNativeScanHandlerClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&ch)))
	if err != nil {
		return err
	}

	err = jutil.CallVoidMethod(env, p.jPlugin, "addScanner", []jutil.Sign{contextSign, uuidSign, scanHandlerSign},
		jContext, jUuid, jNativeScanHandler)

	if err != nil {
		return err
	}
	p.trigger.Add(done, ctx.Done())
	return nil
}


// JavaUUID converts a Go UUID type to a Java UUID object.
func JavaUUID(env jutil.Env, uuid discovery.Uuid) (jutil.Object, error) {
	buf := bytes.NewBuffer(uuid)
	var high, low int64
	binary.Read(buf, binary.BigEndian, &high)
	binary.Read(buf, binary.BigEndian, &low)
	return jutil.NewObject(env, jUUIDClass, []jutil.Sign{jutil.LongSign, jutil.LongSign}, high, low)
}
