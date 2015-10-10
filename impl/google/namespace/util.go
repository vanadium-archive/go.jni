// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package namespace

import (
	"v.io/v23/namespace"
	"v.io/v23/naming"
	"v.io/v23/options"
	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

// JavaNamespace converts the provided Go Namespace into a Java Namespace
// object.
func JavaNamespace(env jutil.Env, namespace namespace.T) (jutil.Object, error) {
	if namespace == nil {
		return jutil.NullObject, nil
	}
	jNamespace, err := jutil.NewObject(env, jNamespaceImplClass, []jutil.Sign{jutil.LongSign}, int64(jutil.PtrValue(&namespace)))
	if err != nil {
		return jutil.NullObject, err
	}
	jutil.GoRef(&namespace) // Un-refed when the Java NamespaceImpl is finalized.
	return jNamespace, nil
}

func javaToGoOptions(env jutil.Env, key string, jValue jutil.Object) (interface{}, error) {
	switch key {
	case "io.v.v23.SKIP_SERVER_ENDPOINT_AUTHORIZATION":
		value, err := jutil.CallBooleanMethod(env, jValue, "booleanValue", []jutil.Sign{})
		if err != nil {
			return nil, err
		}
		if value {
			// TODO(ashankar): The Java APIs need to reflect the
			// change in the Go APIs: any authorization policy can
			// be providfed as an option?
			return options.NameResolutionAuthorizer{security.AllowEveryone()}, nil
		}
	}
	// Otherwise we don't know what this option is, ignore it.
	return nil, jutil.SkipOption
}

func namespaceOptions(env jutil.Env, jOptions jutil.Object) ([]naming.NamespaceOpt, error) {
	opts, err := jutil.GoOptions(env, jOptions, javaToGoOptions)
	if err != nil {
		return nil, err
	}
	var actualOpts []naming.NamespaceOpt
	for _, opt := range opts {
		switch opt := opt.(type) {
		case naming.NamespaceOpt:
			actualOpts = append(actualOpts, opt)
		}
	}
	return actualOpts, nil
}
