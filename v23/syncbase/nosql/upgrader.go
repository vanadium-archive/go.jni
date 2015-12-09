// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package nosql

import (
	"fmt"
	"runtime"

	"v.io/v23/syncbase/nosql"

	jutil "v.io/x/jni/util"
)

// GoUpgrader converts a provided Java SchemaUpgrader into a Go SchemaUpgrader.
func GoUpgrader(env jutil.Env, jUpgrader jutil.Object) nosql.SchemaUpgrader {
	if jUpgrader.IsNull() {
		return nil
	}
	// Reference Java upgrader; it will be de-referenced when the Go upgrader
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jUpgrader = jutil.NewGlobalRef(env, jUpgrader)
	upgrader := &jniUpgrader{
		jUpgrader: jUpgrader,
	}
	runtime.SetFinalizer(upgrader, func(u *jniUpgrader) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, u.jUpgrader)
	})
	return upgrader
}

type jniUpgrader struct {
	jUpgrader jutil.Object
}

func (u *jniUpgrader) Run(db nosql.Database, oldVersion, newVersion int32) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jdb, ok := db.(*jniDatabase)
	if !ok {
		return fmt.Errorf("Unknown database type: %T", db)
	}
	jDatabase, err := javaDatabase(env, jdb)
	if err != nil {
		return err
	}
	databaseSign := jutil.ClassSign("io.v.v23.syncbase.nosql.Database")
	return jutil.CallVoidMethod(env, u.jUpgrader, "run", []jutil.Sign{databaseSign, jutil.IntSign, jutil.IntSign}, jDatabase, oldVersion, newVersion)
}
