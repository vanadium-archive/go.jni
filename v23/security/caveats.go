// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package security

import (
	"v.io/v23/context"
	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

func caveatValidator(context *context.T, caveat security.Caveat) error {
	jEnv, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jContext, err := JavaContext(jEnv, context, nil)
	if err != nil {
		return err
	}
	jCaveat, err := JavaCaveat(jEnv, caveat)
	if err != nil {
		return err
	}
	return jutil.CallStaticVoidMethod(jEnv, jCaveatRegistryClass, "validate", []jutil.Sign{contextSign, caveatSign}, jContext, jCaveat)
}
