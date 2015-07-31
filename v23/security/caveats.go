// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build java android

package security

import (
	"v.io/v23/context"
	"v.io/v23/security"
	jutil "v.io/x/jni/util"
	jcontext "v.io/x/jni/v23/context"
)

// #include "jni.h"
import "C"

func caveatValidator(context *context.T, call security.Call, sets [][]security.Caveat) []error {
	env, freeFunc := jutil.GetEnv()
	defer freeFunc()
	jContext, err := jcontext.JavaContext(env, context, nil)
	if err != nil {
		return errors(err, len(sets))
	}
	jCall, err := JavaCall(env, call)
	if err != nil {
		return errors(err, len(sets))
	}
	ret := make([]error, len(sets))
	for i, set := range sets {
		for _, caveat := range set {
			jCaveat, err := JavaCaveat(env, caveat)
			if err != nil {
				ret[i] = err
				break
			}
			if err := jutil.CallStaticVoidMethod(env, jCaveatRegistryClass, "validate", []jutil.Sign{contextSign, callSign, caveatSign}, jContext, jCall, jCaveat); err != nil {
				ret[i] = err
				break
			}
		}
	}
	return ret
}

func errors(err error, len int) []error {
	ret := make([]error, len)
	for i := 0; i < len; i++ {
		ret[i] = err
	}
	return ret
}
