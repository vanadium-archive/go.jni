// Copyright 2016 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package ble

import (
	"runtime"

	"v.io/v23/context"

	"v.io/x/ref/lib/discovery/plugins/ble"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

type driver struct {
	jDriver jutil.Object
}

func (d *driver) AddService(uuid string, characteristics map[string][]byte) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	csObjMap := make(map[jutil.Object]jutil.Object, len(characteristics))
	for uuid, characteristic := range characteristics {
		jUuid := jutil.JString(env, uuid)
		jCharacteristic, err := jutil.JByteArray(env, characteristic)
		if err != nil {
			return err
		}
		csObjMap[jUuid] = jCharacteristic
	}
	jCharacteristics, err := jutil.JObjectMap(env, csObjMap)
	if err != nil {
		return err
	}
	return jutil.CallVoidMethod(env, d.jDriver, "addService", []jutil.Sign{jutil.StringSign, jutil.MapSign}, uuid, jCharacteristics)
}

func (d *driver) RemoveService(uuid string) {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jutil.CallVoidMethod(env, d.jDriver, "removeService", []jutil.Sign{jutil.StringSign}, uuid)
}

func (d *driver) StartScan(uuids []string, baseUuid, maskUuid string, handler ble.ScanHandler) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	handlerRef := jutil.GoNewRef(&handler) // Un-refed when jNativeScanHandler is finalized.
	jNativeScanHandler, err := jutil.NewObject(env, jNativeScanHandlerClass, []jutil.Sign{jutil.LongSign}, int64(handlerRef))
	if err != nil {
		jutil.GoDecRef(handlerRef)
		return err
	}
	err = jutil.CallVoidMethod(env, d.jDriver, "startScan", []jutil.Sign{jutil.ArraySign(jutil.StringSign), jutil.StringSign, jutil.StringSign, scanHandlerSign}, uuids, baseUuid, maskUuid, jNativeScanHandler)
	if err != nil {
		jutil.GoDecRef(handlerRef)
		return err
	}
	return nil
}

func (d *driver) StopScan() {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jutil.CallVoidMethod(env, d.jDriver, "stopScan", nil)
}

func (d *driver) DebugString() string {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()

	s, _ := jutil.CallStringMethod(env, d.jDriver, "debugString", nil)
	return s
}

func initDriverFactory(env jutil.Env) error {
	jDriverClass, err := jutil.JFindClass(env, "io/v/android/impl/google/discovery/plugins/ble/Driver")
	if err != nil {
		return err
	}
	factory := func(ctx *context.T, _ string) (ble.Driver, error) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()

		jCtx, err := jcontext.JavaContext(env, ctx, nil)
		if err != nil {
			return nil, err
		}
		jDriver, err := jutil.NewObject(env, jDriverClass, []jutil.Sign{contextSign}, jCtx)
		if err != nil {
			return nil, err
		}
		// Reference the driver; it will be de-referenced when the driver is garbage-collected.
		jDriver = jutil.NewGlobalRef(env, jDriver)
		d := &driver{jDriver}
		runtime.SetFinalizer(d, func(*driver) {
			env, freeFunc := jutil.GetEnv()
			jutil.DeleteGlobalRef(env, d.jDriver)
			freeFunc()
		})
		return d, nil
	}
	ble.SetDriverFactory(factory)
	return nil
}
