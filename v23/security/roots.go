// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package security

import (
	"log"
	"runtime"
	"unsafe"

	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaBlessingRoots creates an instance of Java BlessingRoots that uses the provided Go
// BlessingRoots as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaBlessingRoots(jEnv interface{}, roots security.BlessingRoots) (unsafe.Pointer, error) {
	jObj, err := jutil.NewObject(jEnv, jBlessingRootsImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&roots)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&roots) // Un-refed when the Java BlessingRootsImpl is finalized.
	return jObj, nil
}

// GoBlessingRoots creates an instance of security.BlessingRoots that uses the
// provided Java BlessingRoots as its underlying implementation.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func GoBlessingRoots(jEnv, jBlessingRootsObj interface{}) (security.BlessingRoots, error) {
	if jBlessingRootsObj == nil {
		return nil, nil
	}
	// Reference Java BlessingRoots; it will be de-referenced when the Go
	// BlessingRoots created below is garbage-collected (through the finalizer
	// callback we setup just below).
	jBlessingRoots := C.jobject(jutil.NewGlobalRef(jEnv, jBlessingRootsObj))
	r := &blessingRoots{
		jBlessingRoots: jBlessingRoots,
	}
	runtime.SetFinalizer(r, func(r *blessingRoots) {
		envPtr, freeFunc := jutil.GetEnv()
		env := (*C.JNIEnv)(envPtr)
		defer freeFunc()
		jutil.DeleteGlobalRef(env, r.jBlessingRoots)
	})
	return r, nil
}

type blessingRoots struct {
	jBlessingRoots C.jobject
}

func (r *blessingRoots) Add(root security.PublicKey, pattern security.BlessingPattern) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jRoot, err := JavaPublicKey(env, root)
	if err != nil {
		return err
	}
	jPattern, err := JavaBlessingPattern(env, pattern)
	if err != nil {
		return err
	}
	return jutil.CallVoidMethod(env, r.jBlessingRoots, "add", []jutil.Sign{publicKeySign, blessingPatternSign}, jRoot, jPattern)
}

func (r *blessingRoots) Recognized(root security.PublicKey, blessing string) error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jRoot, err := JavaPublicKey(env, root)
	if err != nil {
		return err
	}
	return jutil.CallVoidMethod(env, r.jBlessingRoots, "recognized", []jutil.Sign{publicKeySign, jutil.StringSign}, jRoot, blessing)
}

func (r *blessingRoots) DebugString() string {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jString, err := jutil.CallStringMethod(env, r.jBlessingRoots, "debugString", nil)
	if err != nil {
		log.Printf("Coudln't get Java DebugString: %v", err)
		return ""
	}
	return jutil.GoString(env, jString)
}
