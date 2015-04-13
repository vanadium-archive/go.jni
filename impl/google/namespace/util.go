// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package namespace

import (
	"unsafe"

	"v.io/v23/namespace"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaNamespace converts the provided Go Namespace into a Java Namespace
// object.
// NOTE: Because CGO creates package-local types and because this method may be
// invoked from a different package, Java types are passed in an empty interface
// and then cast into their package local types.
func JavaNamespace(jEnv interface{}, namespace namespace.T) (unsafe.Pointer, error) {
	jNamespace, err := jutil.NewObject(jEnv, jNamespaceImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&namespace)))
	if err != nil {
		return nil, err
	}
	jutil.GoRef(&namespace) // Un-refed when the Java PrincipalImpl is finalized.
	return jNamespace, nil
}
