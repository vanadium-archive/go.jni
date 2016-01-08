// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package nosql

import (
	"runtime"

	"v.io/v23/context"
	"v.io/v23/syncbase/nosql"

	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// GoResolver converts a provided Java ConflictResolver into a Go ConflictResolver.
func GoResolver(env jutil.Env, jResolver jutil.Object) nosql.ConflictResolver {
	if jResolver.IsNull() {
		return nil
	}
	// Reference Java resolver; it will be de-referenced when the Go resolver
	// created below is garbage-collected (through the finalizer callback we
	// setup just below).
	jResolver = jutil.NewGlobalRef(env, jResolver)
	resolver := &jniResolver{
		jResolver: jResolver,
	}
	runtime.SetFinalizer(resolver, func(r *jniResolver) {
		env, freeFunc := jutil.GetEnv()
		defer freeFunc()
		jutil.DeleteGlobalRef(env, r.jResolver)
	})
	return resolver
}

type jniResolver struct {
	jResolver jutil.Object
}

func (r *jniResolver) OnConflict(ctx *context.T, conflict *nosql.Conflict) nosql.Resolution {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jContext, err := jcontext.JavaContext(env, ctx, nil)
	if err != nil {
		panic("Couldn't create Java context: " + err.Error())
	}
	jConflict, err := jutil.JVomCopy(env, *conflict, jConflictClass)
	if err != nil {
		panic("Couldn't create Java Conflict object: " + err.Error())
	}
	contextSign := jutil.ClassSign("io.v.v23.context.VContext")
	conflictSign := jutil.ClassSign("io.v.v23.syncbase.nosql.Conflict")
	resolutionSign := jutil.ClassSign("io.v.v23.syncbase.nosql.Resolution")
	jResolution, err := jutil.CallObjectMethod(env, r.jResolver, "onConflict", []jutil.Sign{contextSign, conflictSign}, resolutionSign, jContext, jConflict)
	if err != nil {
		panic("Error invoking Java ConflictResolver: " + err.Error())
	}
	var resolution nosql.Resolution
	if err := jutil.GoVomCopy(env, jResolution, jResolutionClass, &resolution); err != nil {
		panic("Couldn't create Go Resolution: " + err.Error())
	}
	return resolution
}
