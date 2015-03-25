// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build android

package security

import (
	"v.io/v23/security"
	jutil "v.io/x/jni/util"
)

// #include "jni.h"
import "C"

func caveatValidator(call security.Call, caveat security.Caveat) error {
	jEnv, freeFunc := jutil.GetEnv()
	defer freeFunc()

	jCall, err := JavaCall(jEnv, call)
	if err != nil {
		return err
	}
	jCaveat, err := JavaCaveat(jEnv, caveat)
	if err != nil {
		return err
	}
	return jutil.CallStaticVoidMethod(jEnv, jCaveatRegistryClass, "validate", []jutil.Sign{callSign, caveatSign}, jCall, jCaveat)
}
